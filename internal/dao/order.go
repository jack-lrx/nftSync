package dao

import (
	"time"
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
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DAO 方法定义
func CreateOrder(order *Order) error {
	// TODO: 实现数据库插入逻辑
	return nil
}

func GetOrder(id int64) (*Order, error) {
	// TODO: 实现数据库查询逻辑
	return nil, nil
}

func UpdateOrderStatus(id int64, status string) error {
	// TODO: 实现数据库更新逻辑
	return nil
}

func ListOrders(nftID int64, status string) ([]*Order, error) {
	// TODO: 实现数据库批量查询逻辑
	return nil, nil
}

// 原子撮合订单，更新状态为matched，记录买家和时间
func UpdateOrderMatched(id int64, buyer string) error {
	// TODO: 实现数据库原子更新逻辑
	// UPDATE orders SET status='matched', buyer=?, updated_at=? WHERE id=? AND status='listed'
	return nil
}

// 用户订单列表查询，按 owner 查询
func ListUserOrders(owner string) ([]*Order, error) {
	// TODO: 实现数据库查询逻辑
	// SELECT * FROM orders WHERE seller=? OR buyer=?
	return nil, nil
}
