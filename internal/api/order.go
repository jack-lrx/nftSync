package api

import (
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// 查询响应结构体
type GetOrderResp struct {
	Order *service.OrderDTO `json:"order,omitempty"`
	Error string            `json:"error,omitempty"`
}

// 查询接口
func GetOrderHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDStr := c.Param("id")
		orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
		if err != nil {
			resp := GetOrderResp{Error: "invalid order id"}
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		order, err := a.Service.GetOrder(orderID)
		if err != nil {
			resp := GetOrderResp{Error: err.Error()}
			c.JSON(http.StatusNotFound, resp)
			return
		}
		resp := GetOrderResp{Order: order}
		c.JSON(http.StatusOK, resp)
	}
}

// 用户订单列表查询请求结构体
// 支持按 owner 查询
// GET /api/order/list?owner=xxx

type ListUserOrdersReq struct {
	Owner string `form:"owner" binding:"required"`
}

type ListUserOrdersResp struct {
	Orders []service.OrderDTO `json:"orders,omitempty"`
	Error  string             `json:"error,omitempty"`
}

// 用户订单列表查询接口
func ListUserOrdersHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListUserOrdersReq
		if err := c.ShouldBindQuery(&req); err != nil {
			resp := ListUserOrdersResp{Error: "owner参数必填"}
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		orders, err := service.NewService(ctx).ListUserOrders(req.Owner)
		if err != nil {
			resp := ListUserOrdersResp{Error: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		resp := ListUserOrdersResp{Orders: orders}
		c.JSON(http.StatusOK, resp)
	}
}

// 订单同步请求结构体
// 可扩展为支持区块高度、时间范围等参数
// POST /api/orders/sync

type SyncOrdersReq struct {
	Orders []service.OrderDTO `json:"orders" binding:"required"`
}

type SyncOrdersResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// 订单同步接口
func SyncOrdersHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SyncOrdersReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, SyncOrdersResp{Success: false, Error: "invalid request"})
			return
		}
		// DTO 转 dao.Order
		orders := make([]dao.Order, 0, len(req.Orders))
		for _, dto := range req.Orders {
			orders = append(orders, dao.Order{
				ID:        dto.ID,
				NFTID:     dto.NFTID,
				NFTToken:  dto.NFTToken,
				Seller:    dto.Seller,
				Buyer:     dto.Buyer,
				Price:     dto.Price,
				Status:    dto.Status,
				CreatedAt: dto.CreatedAt,
				UpdatedAt: dto.UpdatedAt,
			})
		}
		if err := service.NewService(ctx).SyncOrders(orders); err != nil {
			c.JSON(http.StatusInternalServerError, SyncOrdersResp{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, SyncOrdersResp{Success: true})
	}
}

// 订单指标统计接口
// GET /api/orders/stats

type GetOrderStatsResp struct {
	Total       int64   `json:"total"`
	TotalAmount float64 `json:"total_amount"`
	Error       string  `json:"error,omitempty"`
}

func GetOrderStatsHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		total, totalAmount, err := service.NewService(ctx).GetOrderStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, GetOrderStatsResp{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, GetOrderStatsResp{Total: total, TotalAmount: totalAmount})
	}
}
