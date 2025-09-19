package config

import (
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

// Context 只包含基础资源，业务对象由 service 层组合
// 生产级别建议避免 config 包依赖 dao/service

type Context struct {
	Config    *AppConfig
	Db        *gorm.DB
	Redis     *redis.Client
	MultiNode *blockchain.MultiNodeEthClient
}

// NewContext 支持传入配置路径，返回错误，便于上层处理
func NewContext(configPath string) (*Context, error) {
	cfg, err := LoadAppConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 初始化 MySQL
	db, err := gorm.Open(mysql.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	clients, names, err := blockchain.NewEthClientsFromConfig(cfg)
	if err != nil {
		log.Fatalf("节点初始化失败: %v", err)
	}

	multiNode := blockchain.NewMultiNodeEthClient(clients, names)

	ctx := &Context{
		Config:    cfg,
		Db:        db,
		Redis:     redisClient,
		MultiNode: multiNode,
	}
	return ctx, nil
}

// Close 优雅关闭资源（如 Redis），DB 由 gorm 管理
func (c *Context) Close() {
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
	// gorm.DB 无需手动关闭
}
