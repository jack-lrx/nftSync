package api

import (
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserApi struct {
	Service *service.Service
}

// 注册请求结构体
// POST /api/user/register
// {"email":"xxx", "password":"xxx", "wallet_addr":"xxx"}
type RegisterUserReq struct {
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	WalletAddr string `json:"wallet_addr"`
}

type RegisterUserResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func RegisterUserHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, RegisterUserResp{Success: false, Error: "参数错误"})
			return
		}
		if err := service.NewService(ctx).RegisterUser(req.Email, req.Password, req.WalletAddr); err != nil {
			c.JSON(http.StatusBadRequest, RegisterUserResp{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, RegisterUserResp{Success: true})
	}
}

// 登录请求结构体
type LoginUserReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUserResp struct {
	Success bool   `json:"success"`
	UserID  int64  `json:"user_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

func LoginUserHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, LoginUserResp{Success: false, Error: "参数错误"})
			return
		}
		user, err := service.NewService(ctx).LoginUser(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, LoginUserResp{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, LoginUserResp{Success: true, UserID: user.ID})
	}
}

// 用户存在性查询请求结构体
// GET /api/user/exists?email=xxx 或 /api/user/exists?wallet_addr=xxx

type UserExistsReq struct {
	Email      string `form:"email"`
	WalletAddr string `form:"wallet_addr"`
}

type UserExistsResp struct {
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

func UserExistsHandler(ctx *config.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserExistsReq
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, UserExistsResp{Exists: false, Error: "参数错误"})
			return
		}
		exists, err := service.NewService(ctx).UserExists(req.Email, req.WalletAddr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, UserExistsResp{Exists: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusOK, UserExistsResp{Exists: exists})
	}
}
