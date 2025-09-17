package main

import (
	"context"
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/service"
	"log"
	"time"
)

func main() {
	cfg, err := config.LoadAppConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
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
