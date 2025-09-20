package dao

import (
	"time"

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

// Order 订单结构体
// 一个订单代表一次NFT挂单或成交
type Order struct {
	ID          int64     `json:"id" db:"id"`
	NFTID       int64     `json:"nft_id" db:"nft_id"`
	NFTToken    string    `json:"nft_token" db:"nft_token"`
	Seller      string    `json:"seller" db:"seller"`
	Buyer       string    `json:"buyer" db:"buyer"`
	Price       float64   `json:"price" db:"price"`
	Fee         float64   `json:"fee" db:"fee"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	TxHash      string    `json:"tx_hash" db:"tx_hash"`
	BlockNumber uint64    `json:"block_number" db:"block_number"`
	BlockTime   int64     `json:"block_time" db:"block_time"`
}

// 创建订单
func (r *Dao) CreateOrder(order *Order) error {
	return r.DB.Create(order).Error
}

// 批量插入订单，已存在数据自动跳过
func (r *Dao) CreateOrdersIgnoreConflict(orders []Order) error {
	return r.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&orders).Error
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
