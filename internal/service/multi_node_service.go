package service

import (
	"fmt"
	"github.com/gavin/nftSync/internal/blockchain"
	"github.com/gavin/nftSync/internal/config"
	"github.com/gavin/nftSync/internal/dao"
	"golang.org/x/net/context"
	"math/big"
	"sync"
)

type MultiNodeSyncService struct {
	MultiNode       *config.MultiNodeEthClient
	lastSyncedBlock *big.Int
	Dao             *dao.Dao
}

func NewMultiNodeSyncService(ctx *config.Context) *MultiNodeSyncService {
	return &MultiNodeSyncService{
		MultiNode:       ctx.MultiNode,
		lastSyncedBlock: big.NewInt(0),
		Dao:             dao.New(ctx.Db),
	}
}

// MultiNodeTransferEvent 采集结果结构体
type MultiNodeTransferEvent struct {
	Event       blockchain.TransferEvent
	SourceNodes []string // 采集到该事件的节点
	Confidence  int      // 置信度（采集到该事件的节点数）
}

// FetchTransferEventsAllNodes 并发采集并交叉验证
func (m *MultiNodeSyncService) FetchTransferEventsAllNodes(contract string, startBlock, endBlock *big.Int, ctx context.Context) []MultiNodeTransferEvent {
	var wg sync.WaitGroup
	results := make([][]blockchain.TransferEvent, len(m.MultiNode.Clients))
	for i, cli := range m.MultiNode.Clients {
		wg.Add(1)
		go func(idx int, c *blockchain.EthClient) {
			defer wg.Done()
			events, err := c.FetchTransferEvents(ctx, contract, startBlock, endBlock)
			if err == nil {
				results[idx] = events
			}
		}(i, blockchain.NewEthClient(cli))
	}
	wg.Wait()
	// 交叉验证：统计各节点采集到的事件
	eventMap := map[string]*MultiNodeTransferEvent{}
	for i, nodeEvents := range results {
		for _, evt := range nodeEvents {
			// 修正：将 BlockNumber (uint64) 转为字符串
			key := evt.Contract + ":" + evt.TokenID + ":" + evt.To + ":" + fmt.Sprintf("%d", evt.BlockNumber)
			if v, ok := eventMap[key]; ok {
				v.Confidence++
				v.SourceNodes = append(v.SourceNodes, m.MultiNode.NodeNames[i])
			} else {
				eventMap[key] = &MultiNodeTransferEvent{
					Event:       evt,
					SourceNodes: []string{m.MultiNode.NodeNames[i]},
					Confidence:  1,
				}
			}
		}
	}
	// 转为切片返回
	finalEvents := []MultiNodeTransferEvent{}
	for _, v := range eventMap {
		finalEvents = append(finalEvents, *v)
	}
	return finalEvents
}
