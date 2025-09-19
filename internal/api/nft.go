package api

import (
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// 请求/响应结构体抽象

type NFTDetailRequest struct {
	Contract string `form:"contract" binding:"required"`
	TokenID  string `form:"token_id" binding:"required"`
}
type NFTDetailResponse struct {
	Data  *service.NFTDetailDTO `json:"data,omitempty"`
	Error string                `json:"error,omitempty"`
}

type NFTListRequest struct {
	Owner string `form:"owner" binding:"required"`
}
type NFTListResponse struct {
	Data  []service.NFTDetailDTO `json:"data,omitempty"`
	Error string                 `json:"error,omitempty"`
}

// Api 提供 NFT 查询接口

// 查询 NFT 详情
func GetNFTDetail(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		var req NFTDetailRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, NFTDetailResponse{Error: "contract and token_id required"})
			return
		}
		nft, err := service.NewService(ctx).GetNFTDetail(c.Request.Context(), req.Contract, req.TokenID)
		if err != nil {
			c.JSON(http.StatusNotFound, NFTDetailResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, NFTDetailResponse{Data: nft})
		log.Printf("[TRACE] GetNFTDetail contract=%s tokenID=%s cost=%v", req.Contract, req.TokenID, time.Since(start))
	}
}

// 查询 owner 的 NFT 列表
func GetNFTListByOwner(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req NFTListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, NFTListResponse{Error: "owner required"})
			return
		}
		nfts, err := service.NewService(ctx).GetNFTListByOwner(c.Request.Context(), req.Owner)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NFTListResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, NFTListResponse{Data: nfts})
	}
}
