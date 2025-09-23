package dao

import (
	"gorm.io/gorm"
)

// FloorPrice 地板价表
// collection: NFT合集合约地址
// price: 当前地板价
// 可根据实际业务扩展字段

// FloorPrice 结构体
type FloorPrice struct {
	ID         int64  `gorm:"primaryKey;column:id" json:"id"`
	Collection string `gorm:"uniqueIndex;column:collection" json:"collection"`
	Price      string `gorm:"column:price" json:"price"`
}

// UpdateFloorPrice 更新地板价
func (r *Dao) UpdateFloorPrice(collection string, price string) error {
	fp := FloorPrice{Collection: collection, Price: price}
	return r.DB.Clauses(gorm.Clauses{gorm.OnConflict{UpdateAll: true}}).Create(&fp).Error
}

// ListOrdersByCollection 查询某合集所有挂单订单
func (r *Dao) ListOrdersByCollection(collection string) ([]Order, error) {
	var orders []Order
	if err := r.DB.Where("nft_token = ? AND status = ?", collection, OrderStatusListed).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetFloorPrice 查询地板价
func (r *Dao) GetFloorPrice(collection string) (string, error) {
	var fp FloorPrice
	if err := r.DB.Where("collection = ?", collection).First(&fp).Error; err != nil {
		return "", err
	}
	return fp.Price, nil
}
