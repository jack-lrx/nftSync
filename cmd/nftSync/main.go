package main

import (
	"context"
	"github.com/gavin/nftSync/internal/api"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/middleware"
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	// 环境变量传递配置路径
	configPath := "configs/config.yaml"
	bizCtx, err := config.NewContext(configPath)
	if err != nil {
		log.Fatalf("Context初始化失败: %v", err)
	}
	defer bizCtx.Close()

	go func() {
		// 推荐使用 gin.New()，避免重复注册默认中间件
		middleware.InitLogger() // 初始化 zap 日志
		r := gin.New()
		// 注册业务中间件
		r.Use(middleware.ZapLogger())
		r.Use(middleware.ZapRecovery())

		apiGroup := r.Group("/api")
		// 注册nft相关接口，添加权限校验
		nftGroup := apiGroup.Group("/nft")
		// 需要权限的接口单独注册 AuthMiddleware
		nftGroup.GET("/detail", middleware.AuthMiddleware(), api.GetNFTDetail(bizCtx))
		nftGroup.GET("/list", middleware.AuthMiddleware(), api.GetNFTListByOwner(bizCtx))

		// 注册订单相关接口，添加权限校验
		orderGroup := apiGroup.Group("/order")
		orderGroup.Use(middleware.AuthMiddleware())
		orderGroup.GET(":id", api.GetOrderHandler(bizCtx))
		orderGroup.GET("/list", api.ListUserOrdersHandler(bizCtx))

		// 注册用户相关接口，无需权限校验
		userGroup := apiGroup.Group("/user")
		userGroup.POST("/register", api.RegisterUserHandler(bizCtx))
		userGroup.POST("/login", api.LoginUserHandler(bizCtx))
		userGroup.GET("/exists", api.UserExistsHandler(bizCtx))

		if err := r.Run(":8080"); err != nil {
			log.Fatalf("API服务启动失败: %v", err)
		}
	}()

	multiNodeSyncService := service.NewMultiNodeSyncService(bizCtx)
	// 启动nft实时同步 goroutine
	go func() {
		realtimeTicker := time.NewTicker(time.Duration(bizCtx.Config.Sync.RealtimeInterval) * time.Second)
		defer realtimeTicker.Stop()
		ctx := context.Background()
		for {
			<-realtimeTicker.C
			multiNodeSyncService.SyncMintEventsRealtime(ctx, bizCtx)
		}
	}()

	// 启动nft补全同步 goroutine
	go func() {
		pollingTicker := time.NewTicker(time.Duration(bizCtx.Config.Sync.PollingInterval) * time.Second)
		defer pollingTicker.Stop()
		ctx := context.Background()
		for {
			<-pollingTicker.C
			multiNodeSyncService.SyncMintEventsPolling(ctx, bizCtx)
		}
	}()

	// 启动订单同步 goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(bizCtx.Config.Sync.OrderInterval) * time.Second)
		defer ticker.Stop()
		ctx := context.Background()
		for {
			<-ticker.C
			multiNodeSyncService.SyncOrderEventsPolling(ctx, bizCtx)
		}
	}()

	//地板价消息消费
	err = service.NewFloorPriceService(bizCtx).StartKafkaConsumer(context.Background())
	if err != nil {
		log.Fatalf("地板价消息消费启动失败: %v", err)
	}

	select {} // 阻塞主 goroutine，防止退出
}
