package config

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Context 只包含基础资源，业务对象由 service 层组合
// 生产级别建议避免 config 包依赖 dao/service

type Context struct {
	Config    *AppConfig
	Db        *gorm.DB
	Redis     *redis.Client
	MultiNode *MultiNodeEthClient
}

type MultiNodeEthClient struct {
	Clients   []*ethclient.Client
	NodeNames []string // 节点标识
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

	multiNode, err := newMultiNodeEthClient(cfg)
	if err != nil {
		return nil, err
	}

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

// newMultiNodeEthClient 根据配置初始化所有节点
func newMultiNodeEthClient(cfg *AppConfig) (*MultiNodeEthClient, error) {
	clients := []*ethclient.Client{}
	names := []string{}
	for _, node := range cfg.EthNodes {
		cli, err := ethclient.Dial(node.URL)
		if err != nil {
			return nil, err
		}
		clients = append(clients, cli)
		names = append(names, node.Name)
	}
	return &MultiNodeEthClient{
		Clients:   clients,
		NodeNames: names,
	}, nil
}
