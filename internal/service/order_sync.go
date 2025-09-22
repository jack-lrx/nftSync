package service

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"gorm.io/gorm"
	"log"
	"math/big"
	"strconv"
	"time"
)

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
	orderCreatedSig := "OrderCreated(address,address,uint256,uint256,uint256)"
	orderCancelledSig := "OrderCancelled(bytes32,address)"
	orderFilledSig := "OrderFilled(address,address,uint256,uint256,uint256)"
	createdTopic := crypto.Keccak256Hash([]byte(orderCreatedSig))
	cancelledTopic := crypto.Keccak256Hash([]byte(orderCancelledSig))
	filledTopic := crypto.Keccak256Hash([]byte(orderFilledSig))

	var createOrders, cancelOrders, matchOrders []dao.Order

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
				// 订单创建事件解析
				// topics: [OrderCreated, seller, nftToken, ...]
				if len(vLog.Topics) < 3 {
					continue
				}
				order := dao.Order{
					NFTToken:    vLog.Topics[2].Hex(),
					Seller:      vLog.Topics[1].Hex(),
					Status:      dao.OrderStatusListed,
					TxHash:      vLog.TxHash.Hex(),
					BlockNumber: vLog.BlockNumber,
					BlockTime:   blockTime,
					CreatedAt:   time.Unix(blockTime, 0),
					UpdatedAt:   time.Unix(blockTime, 0),
				}
				createOrders = append(createOrders, order)
			} else if topic0 == cancelledTopic {
				// 订单取消事件解析
				order := dao.Order{
					TxHash:      vLog.TxHash.Hex(),
					Status:      dao.OrderStatusCancelled,
					BlockNumber: vLog.BlockNumber,
					BlockTime:   blockTime,
					UpdatedAt:   time.Unix(blockTime, 0),
				}
				cancelOrders = append(cancelOrders, order)
			} else if topic0 == filledTopic {
				// 订单成交事件解析
				if len(vLog.Topics) < 4 {
					continue
				}
				order := dao.Order{
					NFTToken:    vLog.Topics[3].Hex(),
					Seller:      vLog.Topics[1].Hex(),
					Buyer:       vLog.Topics[2].Hex(),
					Status:      dao.OrderStatusCompleted,
					TxHash:      vLog.TxHash.Hex(),
					BlockNumber: vLog.BlockNumber,
					BlockTime:   blockTime,
					CreatedAt:   time.Unix(blockTime, 0),
					UpdatedAt:   time.Unix(blockTime, 0),
				}
				matchOrders = append(matchOrders, order)
			}
		}
	}
	// 批量插入新订单
	if len(createOrders) > 0 {
		err := s.Dao.DB.Transaction(func(tx *gorm.DB) error {
			return s.Dao.CreateOrdersIgnoreConflict(createOrders)
		})
		if err != nil {
			log.Printf("[order_sync] 新订单批量插入失败: %v", err)
		} else {
			log.Printf("[order_sync] 已同步新订单数: %d", len(createOrders))
		}
	}
	// 批量更新取消订单
	if len(cancelOrders) > 0 {
		for _, order := range cancelOrders {
			_ = s.Dao.UpdateOrderStatusByTxHash(order.TxHash, dao.OrderStatusCancelled)
		}
		log.Printf("[order_sync] 已同步取消订单数: %d", len(cancelOrders))
	}
	// 批量更新成交订单
	if len(matchOrders) > 0 {
		for _, order := range matchOrders {
			_ = s.Dao.UpdateOrderStatusByTxHash(order.TxHash, dao.OrderStatusCompleted)
		}
		log.Printf("[order_sync] 已同步成交订单数: %d", len(matchOrders))
	}
	log.Printf("[order_sync] 订单轮询同步完成，已安全同步到区块 %v", safeBlock)
	return safeBlock
}

func parseTokenID(tokenID string) int64 {
	id, err := strconv.ParseInt(tokenID, 0, 64)
	if err != nil {
		return 0
	}
	return id
}

func dedupOrders(orders []dao.Order) []dao.Order {
	seen := make(map[string]struct{})
	var result []dao.Order
	for _, o := range orders {
		key := o.TxHash + ":" + o.NFTToken + ":" + strconv.FormatUint(o.BlockNumber, 10)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, o)
		}
	}
	return result
}
