package api

import (
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

func RegisterUserHandler(c *gin.Context) {
	var req RegisterUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterUserResp{Success: false, Error: "参数错误"})
		return
	}
	if err := service.RegisterUser(req.Email, req.Password, req.WalletAddr); err != nil {
		c.JSON(http.StatusBadRequest, RegisterUserResp{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RegisterUserResp{Success: true})
}

// 登录请求结构体
// POST /api/user/login
// {"email":"xxx", "password":"xxx"}
type LoginUserReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUserResp struct {
	Success bool   `json:"success"`
	UserID  int64  `json:"user_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

func LoginUserHandler(c *gin.Context) {
	var req LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginUserResp{Success: false, Error: "参数错误"})
		return
	}
	user, err := service.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginUserResp{Success: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, LoginUserResp{Success: true, UserID: user.ID})
}
