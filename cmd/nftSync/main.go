package main

import (
	"context"
	"github.com/gavin/nftSync/internal/api"
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"github.com/gavin/nftSync/internal/middleware"
	"github.com/gavin/nftSync/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

func main() {
	cfg, err := config.LoadAppConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 初始化 MySQL
	db, err := gorm.Open(mysql.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	if err := db.AutoMigrate(&dao.NFT{}, &dao.Item{}); err != nil {
		log.Fatalf("表结构迁移失败: %v", err)
	}
	db.AutoMigrate(&dao.NFT{}, &dao.Item{})

	// 初始化 Redis
	service.InitRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB) // 可从 cfg 读取

	// 创建 NFTService 实例
	nftService := &service.NFTService{Repo: &dao.NFTRepository{DB: db}}

	// 启动 Gin Web 服务，集成业务中间件
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
		apiObj := &api.NFTApi{Service: nftService}
		// 需要权限的接口单独注册 AuthMiddleware
		nftGroup.GET("/detail", middleware.AuthMiddleware(), apiObj.GetNFTDetail)
		nftGroup.GET("/list", middleware.AuthMiddleware(), apiObj.GetNFTListByOwner)

		// 注册订单相关接口，添加权限校验
		orderGroup := apiGroup.Group("/order")
		orderGroup.Use(middleware.AuthMiddleware())
		orderGroup.GET(":id", api.GetOrderHandler)
		orderGroup.GET("/list", api.ListUserOrdersHandler)

		// 注册用户相关接口，无需权限校验
		userGroup := apiGroup.Group("/user")
		userGroup.POST("/register", api.RegisterUserHandler)
		userGroup.POST("/login", api.LoginUserHandler)

		if err := r.Run(":8080"); err != nil {
			log.Fatalf("API服务启动失败: %v", err)
		}
	}()

	clients, names, err := blockchain.NewEthClientsFromConfig(cfg)
	if err != nil {
		log.Fatalf("节点初始化失败: %v", err)
	}
	multiNode := blockchain.NewMultiNodeEthClient(clients, names)
	syncService := service.NewMultiNodeSyncService(multiNode)

	// 实时监听任务（每新区块触发）
	realtimeTicker := time.NewTicker(time.Duration(config.GlobalConfig.Sync.RealtimeInterval) * time.Second)
	defer realtimeTicker.Stop()
	pollingTicker := time.NewTicker(time.Duration(config.GlobalConfig.Sync.PollingInterval) * time.Second)
	defer pollingTicker.Stop()
	done := make(chan struct{})

	for {
		select {
		case <-realtimeTicker.C:
			ctx := context.Background()
			syncService.SyncMintEventsRealtime(ctx)
		case <-pollingTicker.C:
			ctx := context.Background()
			syncService.SyncMintEventsPolling(ctx)
		case <-done:
			return
		}
	}
}
