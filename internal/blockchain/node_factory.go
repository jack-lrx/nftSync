package blockchain

import (
	"github.com/gavin/nftSync/internal/config"
)

// NewEthClientsFromConfig 根据配置初始化所有节点
func NewEthClientsFromConfig(cfg *config.AppConfig) ([]*EthClient, []string, error) {
	clients := []*EthClient{}
	names := []string{}
	for _, node := range cfg.EthNodes {
		cli, err := NewEthClient(node.URL)
		if err != nil {
			return nil, nil, err
		}
		clients = append(clients, cli)
		names = append(names, node.Name)
	}
	return clients, names, nil
}
