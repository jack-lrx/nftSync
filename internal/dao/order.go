package dao

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderStatus 订单状态枚举
const (
	OrderStatusListed    = "listed"    // 挂单中
	OrderStatusMatched   = "matched"   // 已撮合
	OrderStatusCompleted = "completed" // 已成交
	OrderStatusCancelled = "cancelled" // 已取消
)

// 订单类型常量
const (
	OrderTypeListing       = "listing"       // 挂单（卖家出售）
	OrderTypeItemBid       = "itembid"       // 单品出价（买家对某个NFT出价）
	OrderTypeCollectionBid = "collectionbid" // 合集出价（买家对整个合集出价）
)

// Order 订单结构体
// 一个订单代表一次NFT挂单或成交
type Order struct {
	ID       int64  `gorm:"primaryKey;column:id" json:"id"`
	OrderID  string `gorm:"uniqueIndex;column:order_id" json:"order_id"` // 订单唯一键
	NFTID    int64  `gorm:"column:nft_id" json:"nft_id"`
	NFTToken string `gorm:"column:nft_token" json:"nft_token"`
	Seller   string `gorm:"column:seller" json:"seller"`
	Buyer    string `gorm:"column:buyer" json:"buyer"`
	// OrderType 字段已加入，所有方法已同步
	Price       decimal.Decimal `gorm:"type:decimal(38,18);column:price" json:"price"`
	Fee         decimal.Decimal `gorm:"type:decimal(38,18);column:fee" json:"fee"`
	Status      string          `gorm:"column:status" json:"status"`
	CreatedAt   time.Time       `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	TxHash      string          `gorm:"column:tx_hash" json:"tx_hash"`
	BlockNumber uint64          `gorm:"column:block_number" json:"block_number"`
	BlockTime   int64           `gorm:"column:block_time" json:"block_time"`
	OrderType   string          `gorm:"column:order_type" json:"order_type"` // 订单类型
}

// 创建订单
func (r *Dao) CreateOrder(order *Order) error {
	return r.DB.Create(order).Error
}

// 批量插入订单，已存在数据自动跳过
func (r *Dao) CreateOrderIgnoreConflict(order *Order) error {
	return r.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(order).Error
}

// 查询订单详情
func (r *Dao) GetOrder(id int64) (*Order, error) {
	var order Order
	if err := r.DB.Where("id = ?", id).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// 更新订单状态
func (r *Dao) UpdateOrderStatus(id int64, status string) error {
	return r.DB.Model(&Order{}).Where("id = ?", id).Update("status", status).Error
}

// 批量查询订单（按NFT和状态）
func (r *Dao) ListOrders(nftID int64, status string) ([]Order, error) {
	var orders []Order
	if err := r.DB.Where("nft_id = ? AND status = ?", nftID, status).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// 原子撮合订单，更新状态为matched，记录买家和时间
func (r *Dao) UpdateOrderMatched(id int64, buyer string) error {
	return r.DB.Model(&Order{}).Where("id = ? AND status = ?", id, OrderStatusListed).
		Updates(map[string]interface{}{
			"status":     OrderStatusMatched,
			"buyer":      buyer,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

// 用户订单列表查询，按 owner 查询（卖家或买家）
func (r *Dao) ListUserOrders(owner string) ([]Order, error) {
	var orders []Order
	if err := r.DB.Where("seller = ? OR buyer = ?", owner, owner).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// 订单统计（生产级，统计已成交订单数和总金额）
type OrderStats struct {
	Total       int64
	TotalAmount float64
}

func (r *Dao) GetOrderStats() (*OrderStats, error) {
	var stats OrderStats
	err := r.DB.Model(&Order{}).
		Select("COUNT(*) as total, COALESCE(SUM(price),0) as total_amount").
		Where("status = ?", OrderStatusCompleted).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// 根据OrderID查询订单
func (r *Dao) GetOrderByOrderID(orderId string) (*Order, error) {
	var order Order
	if err := r.DB.Where("order_id = ?", orderId).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// 根据OrderID更新订单状态
func (r *Dao) UpdateOrderStatusByOrderID(orderId string, status string) error {
	return r.DB.Model(&Order{}).Where("order_id = ?", orderId).Update("status", status).Error
}
