package api

import (
	"github.com/gavin/nftSync/internal/middleware"
	"github.com/gavin/nftSync/internal/model"
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 请求/响应结构体抽象

type NFTDetailRequest struct {
	Contract string `form:"contract" binding:"required"`
	TokenID  string `form:"token_id" binding:"required"`
}
type NFTDetailResponse struct {
	Data  *model.NFT `json:"data,omitempty"`
	Error string     `json:"error,omitempty"`
}

type NFTListRequest struct {
	Owner string `form:"owner" binding:"required"`
}
type NFTListResponse struct {
	Data  []model.NFT `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// NFTApi 提供 NFT 查询接口

type NFTApi struct {
	Service *service.NFTService
}

// 注册路由
func RegisterNFTApi(r *gin.Engine, svc *service.NFTService) {
	api := &NFTApi{Service: svc}
	// 需要权限的接口单独注册 AuthMiddleware
	r.GET("/api/nft/detail", middleware.AuthMiddleware(), api.GetNFTDetail)
	r.GET("/api/nft/list", middleware.AuthMiddleware(), api.GetNFTListByOwner)
	// 可在此注册无需权限的公开接口
}

// 查询 NFT 详情
func (a *NFTApi) GetNFTDetail(c *gin.Context) {
	var req NFTDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, NFTDetailResponse{Error: "contract and token_id required"})
		return
	}
	nft, err := a.Service.GetNFTDetail(c.Request.Context(), req.Contract, req.TokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, NFTDetailResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, NFTDetailResponse{Data: nft})
}

// 查询 owner 的 NFT 列表
func (a *NFTApi) GetNFTListByOwner(c *gin.Context) {
	var req NFTListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, NFTListResponse{Error: "owner required"})
		return
	}
	nfts, err := a.Service.GetNFTListByOwner(c.Request.Context(), req.Owner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NFTListResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, NFTListResponse{Data: nfts})
}
