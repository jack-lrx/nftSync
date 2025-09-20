package service

import (
	"context"
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
	latestBlock, err := mainClient.GetBlockNumber(ctx)
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
	var orders []dao.Order

	for _, contract := range orderContracts {
		orderEvents, err := mainClient.FetchOrderFilledEvents(ctx, contract, startBlock, safeBlock)
		if err != nil {
			log.Printf("[order_sync] 主节点订单事件拉取失败: %v，尝试其他节点补全", err)
			for i := 1; i < len(s.MultiNode.Clients); i++ {
				orderEvents, err = s.MultiNode.Clients[i].FetchOrderFilledEvents(ctx, contract, startBlock, safeBlock)
				if err == nil && len(orderEvents) > 0 {
					break
				}
			}
			if err != nil {
				log.Printf("[order_sync] 所有节点订单事件拉取失败: %v", err)
				continue
			}
		}
		for _, evt := range orderEvents {
			nftID := parseTokenID(evt.TokenID)
			order := dao.Order{
				NFTID:       nftID,
				NFTToken:    evt.TokenID,
				Seller:      evt.Seller,
				Buyer:       evt.Buyer,
				Price:       evt.Price,
				Fee:         evt.Fee,
				Status:      dao.OrderStatusCompleted,
				CreatedAt:   time.Unix(evt.BlockTime, 0),
				UpdatedAt:   time.Unix(evt.BlockTime, 0),
				TxHash:      evt.TxHash,
				BlockNumber: evt.BlockNumber,
				BlockTime:   evt.BlockTime,
			}
			orders = append(orders, order)
		}
	}
	orders = dedupOrders(orders)
	if len(orders) > 0 {
		err := s.Dao.DB.Transaction(func(tx *gorm.DB) error {
			return s.Dao.CreateOrdersIgnoreConflict(orders)
		})
		if err != nil {
			log.Printf("[order_sync] 订单批量插入失败: %v", err)
		} else {
			log.Printf("[order_sync] 已同步订单数: %d", len(orders))
		}
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
