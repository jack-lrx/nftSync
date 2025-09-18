package api

import (
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
func GetOrderHandler(a *OrderApi) gin.HandlerFunc {
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
func ListUserOrdersHandler(a *OrderApi) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ListUserOrdersReq
		if err := c.ShouldBindQuery(&req); err != nil {
			resp := ListUserOrdersResp{Error: "owner参数必填"}
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		orders, err := a.Service.ListUserOrders(req.Owner)
		if err != nil {
			resp := ListUserOrdersResp{Error: err.Error()}
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		resp := ListUserOrdersResp{Orders: orders}
		c.JSON(http.StatusOK, resp)
	}
}

type OrderApi struct {
	Service *service.OrderService
}
