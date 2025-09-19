package dao

import (
	"time"

	"gorm.io/gorm"
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
	ID        int64     `json:"id" db:"id"`
	NFTID     int64     `json:"nft_id" db:"nft_id"`
	NFTToken  string    `json:"nft_token" db:"nft_token"`
	Seller    string    `json:"seller" db:"seller"`
	Buyer     string    `json:"buyer" db:"buyer"`
	Price     float64   `json:"price" db:"price"`
	Fee       float64   `json:"fee" db:"fee"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Dao 订单数据访问对象
// 用于分层管理订单持久化逻辑，风格参考NFTRepository

// 创建订单
func (r *Dao) CreateOrder(order *Order) error {
	return r.DB.Create(order).Error
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
