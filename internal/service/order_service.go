package service

import (
	"github.com/gavin/nftSync/internal/dao"
	"time"
)

// GetOrder 查询订单，直接返回 DTO
func (s *Service) GetOrder(orderID int64) (*OrderDTO, error) {
	order, err := s.Dao.GetOrder(orderID)
	if err != nil {
		return nil, err
	}
	return s.ToOrderDTO(order), nil
}

// ListUserOrders 用户订单列表查询，按 owner 查询
func (s *Service) ListUserOrders(owner string) ([]OrderDTO, error) {
	orders, err := s.Dao.ListUserOrders(owner)
	if err != nil {
		return nil, err
	}
	res := make([]OrderDTO, 0, len(orders))
	for i := range orders {
		res = append(res, *s.ToOrderDTO(&orders[i]))
	}
	return res, nil
}

// OrderDTO 用于安全输出订单信息
// 可根据实际业务裁剪字段
type OrderDTO struct {
	ID        int64     `json:"id"`
	NFTID     int64     `json:"nft_id"`
	NFTToken  string    `json:"nft_token"`
	Seller    string    `json:"seller"`
	Buyer     string    `json:"buyer"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Fee       float64   `json:"fee"`
}

// ToOrderDTO 将 dao.Order 转换为 OrderDTO
func (s *Service) ToOrderDTO(order *dao.Order) *OrderDTO {
	if order == nil {
		return nil
	}
	return &OrderDTO{
		ID:        order.ID,
		NFTID:     order.NFTID,
		NFTToken:  order.NFTToken,
		Seller:    order.Seller,
		Buyer:     order.Buyer,
		Price:     order.Price,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}

// 订单统计 DTO
// 可根据实际业务裁剪字段
type OrderStatsDTO struct {
	Total       int64   `json:"total"`
	TotalAmount float64 `json:"total_amount"`
}

// GetOrderStats 查询订单统计（生产级）
func (s *Service) GetOrderStats() (*OrderStatsDTO, error) {
	stats, err := s.Dao.GetOrderStats()
	if err != nil {
		return nil, err
	}
	return &OrderStatsDTO{
		Total:       stats.Total,
		TotalAmount: stats.TotalAmount,
	}, nil
}
