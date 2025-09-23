package service

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"log"
)

// FloorPriceService 负责消费地板价更新消息并更新地板价
type FloorPriceService struct {
	Dao                *dao.Dao
	FloorPriceConsumer sarama.PartitionConsumer
}

func NewFloorPriceService(bizCtx *config.Context) *FloorPriceService {
	return &FloorPriceService{
		Dao:                dao.New(bizCtx.Db),
		FloorPriceConsumer: bizCtx.FloorPriceConsumer,
	}
}

func (fps *FloorPriceService) StartKafkaConsumer(ctx context.Context) {
	for {
		select {
		case msg := <-fps.FloorPriceConsumer.Messages():
			collection := string(msg.Value)
			log.Printf("[floor_price] 收到地板价更新消息: %s", collection)
			fps.UpdateFloorPrice(collection)
		case err := <-fps.FloorPriceConsumer.Errors():
			log.Printf("[floor_price] Kafka消费错误: %v", err)
		case <-ctx.Done():
			fps.FloorPriceConsumer.Close()
			return
		}
	}
}

// UpdateFloorPrice 计算并更新地板价
func (fps *FloorPriceService) UpdateFloorPrice(collection string) {
	// 查询该合集所有挂单，取最低价
	orders, err := fps.Dao.ListOrdersByCollection(collection)
	if err != nil {
		log.Printf("[floor_price] 查询订单失败: %v", err)
		return
	}
	var minPrice string
	for i, order := range orders {
		if i == 0 || order.Price.LessThan(orders[i-1].Price) {
			minPrice = order.Price.String()
		}
	}
	if minPrice == "" {
		log.Printf("[floor_price] 合集无挂单: %s", collection)
		return
	}
	// 更新地板价
	if err := fps.Dao.UpdateFloorPrice(collection, minPrice); err != nil {
		log.Printf("[floor_price] 地板价更新失败: %v", err)
	} else {
		log.Printf("[floor_price] 地板价已更新: %s -> %s", collection, minPrice)
	}
}
