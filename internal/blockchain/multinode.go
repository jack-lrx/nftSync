package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"sync"
)

type MultiNodeEthClient struct {
	Clients   []*EthClient
	NodeNames []string // 节点标识
}

// NewMultiNodeEthClient 创建多节点客户端
func NewMultiNodeEthClient(clients []*EthClient, names []string) *MultiNodeEthClient {
	return &MultiNodeEthClient{Clients: clients, NodeNames: names}
}

// MultiNodeTransferEvent 采集结果结构体
type MultiNodeTransferEvent struct {
	Event       TransferEvent
	SourceNodes []string // 采集到该事件的节点
	Confidence  int      // 置信度（采集到该事件的节点数）
}

// FetchTransferEventsAllNodes 并发采集并交叉验证
func (m *MultiNodeEthClient) FetchTransferEventsAllNodes(contract string, startBlock, endBlock *big.Int, ctx context.Context) []MultiNodeTransferEvent {
	var wg sync.WaitGroup
	results := make([][]TransferEvent, len(m.Clients))
	for i, cli := range m.Clients {
		wg.Add(1)
		go func(idx int, c *EthClient) {
			defer wg.Done()
			events, err := c.FetchTransferEvents(ctx, contract, startBlock, endBlock)
			if err == nil {
				results[idx] = events
			}
		}(i, cli)
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
				v.SourceNodes = append(v.SourceNodes, m.NodeNames[i])
			} else {
				eventMap[key] = &MultiNodeTransferEvent{
					Event:       evt,
					SourceNodes: []string{m.NodeNames[i]},
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
