package service

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"strings"
	"time"
)

var orderCreatedEventABI = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"orderId","type":"bytes32"},{"indexed":false,"name":"seller","type":"address"},{"indexed":false,"name":"nftToken","type":"address"},{"indexed":false,"name":"tokenId","type":"uint256"},{"indexed":false,"name":"price","type":"uint256"},{"indexed":false,"name":"fee","type":"uint256"},{"indexed":false,"name":"isBid","type":"bool"},{"indexed":false,"name":"isCollectionBid","type":"bool"}],"name":"OrderCreated","type":"event"}]`

// SyncOrderEventsPolling 主节点优先订单同步（生产级，面向对象）
func (s *MultiNodeSyncService) SyncOrderEventsPolling(ctx context.Context, bizCtx *config.Context) *big.Int {
	mainClient := s.MultiNode.Clients[0]
	ethClient := blockchain.NewEthClient(mainClient)
	latestBlock, err := ethClient.GetBlockNumber(ctx)
	if err != nil {
		log.Printf("[order_sync] 主节点获取最新区块失败: %v", err)
		return s.lastSyncedBlock
	}
	confirmBlocks := bizCtx.Config.Sync.ConfirmBlocks
	safeBlock := new(big.Int).Sub(latestBlock, big.NewInt(int64(confirmBlocks)))
	startBlock := new(big.Int).Add(s.lastSyncedBlock, big.NewInt(1))
	if safeBlock.Cmp(startBlock) < 0 {
		log.Println("[order_sync] 无新区块达到安全确认高度，无需轮询")
		return s.lastSyncedBlock
	}
	orderContracts := bizCtx.Config.OrderContracts

	// 事件topic hash，与eth.go保持一致
	orderCreatedSig := "OrderCreated(address,address,uint256,uint256,uint256,bool,bool)"
	orderCancelledSig := "OrderCancelled(bytes32,address)"
	orderFilledSig := "OrderFilled(bytes32,bytes32,address,address,uint256,uint256,uint256)"
	createdTopic := crypto.Keccak256Hash([]byte(orderCreatedSig))
	cancelledTopic := crypto.Keccak256Hash([]byte(orderCancelledSig))
	filledTopic := crypto.Keccak256Hash([]byte(orderFilledSig))

	orderCreatedABI, err := abi.JSON(strings.NewReader(orderCreatedEventABI))
	if err != nil {
		log.Printf("[order_sync] 订单事件ABI解析失败: %v", err)
		return s.lastSyncedBlock
	}

	topics := []common.Hash{createdTopic, cancelledTopic, filledTopic}
	for _, contract := range orderContracts {
		logs, err := ethClient.FetchOrderEvents(ctx, contract, startBlock, safeBlock, topics)
		if err != nil {
			log.Printf("[order_sync] 主节点订单事件拉取失败: %v，尝试其他节点补全", err)
			for i := 1; i < len(s.MultiNode.Clients); i++ {
				logs, err = blockchain.NewEthClient(s.MultiNode.Clients[i]).
					FetchOrderEvents(ctx, contract, startBlock, safeBlock, topics)
				if err == nil && len(logs) > 0 {
					break
				}
			}
			if err != nil {
				log.Printf("[order_sync] 所有节点订单事件拉取失败: %v", err)
				continue
			}
		}
		for _, vLog := range logs {
			blockTime := int64(0)
			block, err := ethClient.GetBlockByNumber(ctx, vLog.BlockNumber)
			if err == nil {
				blockTime = int64(block.Time())
			}
			if len(vLog.Topics) == 0 {
				continue
			}
			topic0 := vLog.Topics[0]
			if topic0 == createdTopic {
				var createdLog struct {
					orderId         [32]byte
					seller          common.Address
					nftToken        common.Address
					tokenId         *big.Int
					price           *big.Int
					fee             *big.Int
					isBid           bool
					isCollectionBid bool
				}
				// 订单创建事件解析（ABI解包）
				if err := orderCreatedABI.UnpackIntoInterface(&createdLog, "OrderCreated", vLog.Data); err != nil {
					log.Printf("[order_sync] 订单事件ABI解包失败: %v", err)
					continue
				}
				var orderType string
				if createdLog.isBid {
					if createdLog.isCollectionBid {
						orderType = dao.OrderTypeCollectionBid
					} else {
						orderType = dao.OrderTypeItemBid
					}
				} else {
					orderType = dao.OrderTypeListing
				}
				order := dao.Order{
					OrderID:     common.BytesToHash(createdLog.orderId[:]).Hex(),
					NFTToken:    createdLog.nftToken.Hex(),
					Seller:      createdLog.seller.Hex(),
					Status:      dao.OrderStatusListed,
					TxHash:      vLog.TxHash.Hex(),
					BlockNumber: vLog.BlockNumber,
					BlockTime:   blockTime,
					CreatedAt:   time.Unix(blockTime, 0),
					UpdatedAt:   time.Unix(blockTime, 0),
					Price:       decimal.NewFromBigInt(createdLog.price, 0),
					Fee:         decimal.NewFromBigInt(createdLog.fee, 0),
					OrderType:   orderType,
				}
				if err := s.Dao.CreateOrderIgnoreConflict(&order); err != nil {
					log.Printf("[order_sync] 新订单插入失败: %v", err)
				} else {
					log.Printf("[order_sync] 新订单已同步: %s, orderId: %s", order.TxHash, order.OrderID)
				}
			} else if topic0 == cancelledTopic {
				// 订单取消事件解析，orderId直接从topics[1]获取
				if len(vLog.Topics) < 2 {
					log.Printf("[order_sync] 取消事件topics不足: txHash=%s", vLog.TxHash.Hex())
					continue
				}
				orderId := vLog.Topics[1].Hex()
				// 查询订单是否存在
				order, err := s.Dao.GetOrderByOrderID(orderId)
				if err != nil {
					log.Printf("[order_sync] 查询订单失败: %v", err)
					continue
				}
				if order == nil {
					log.Printf("[order_sync] 取消事件未找到订单: orderId=%s", orderId)
					continue
				}
				// 更新订单状态为取消
				if err := s.Dao.UpdateOrderStatusByOrderID(orderId, dao.OrderStatusCancelled); err != nil {
					log.Printf("[order_sync] 取消订单更新失败: %v", err)
				} else {
					log.Printf("[order_sync] 取消订单已同步: orderId=%s", orderId)
				}
			} else if topic0 == filledTopic {
				// 订单完成事件，orderId从topics[1]、topics[2]获取
				if len(vLog.Topics) < 3 {
					log.Printf("[order_sync] 成交事件topics不足: txHash=%s", vLog.TxHash.Hex())
					continue
				}
				sellerOrderId := vLog.Topics[1].Hex()
				buyerOrderId := vLog.Topics[2].Hex()
				// 卖家订单状态更新
				sellerOrder, err := s.Dao.GetOrderByOrderID(sellerOrderId)
				if err != nil {
					log.Printf("[order_sync] 查询卖家订单失败: %v", err)
				} else if sellerOrder != nil {
					if err := s.Dao.UpdateOrderStatusByOrderID(sellerOrderId, dao.OrderStatusCompleted); err != nil {
						log.Printf("[order_sync] 卖家订单状态更新失败: %v", err)
					} else {
						log.Printf("[order_sync] 卖家订单已完成: orderId=%s", sellerOrderId)
					}
				} else {
					log.Printf("[order_sync] 卖家订单不存在: orderId=%s", sellerOrderId)
				}
				// 买家订单状态更新
				buyerOrder, err := s.Dao.GetOrderByOrderID(buyerOrderId)
				if err != nil {
					log.Printf("[order_sync] 查询买家订单失败: %v", err)
				} else if buyerOrder != nil {
					if err := s.Dao.UpdateOrderStatusByOrderID(buyerOrderId, dao.OrderStatusCompleted); err != nil {
						log.Printf("[order_sync] 买家订单状态更新失败: %v", err)
					} else {
						log.Printf("[order_sync] 买家订单已完成: orderId=%s", buyerOrderId)
					}
				} else {
					log.Printf("[order_sync] 买家订单不存在: orderId=%s", buyerOrderId)
				}
			}
		}
	}
	log.Printf("[order_sync] 订单轮询同步完成，已安全同步到区块 %v", safeBlock)
	return safeBlock
}
