package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type NodeConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
type SyncConfig struct {
	RealtimeInterval int `yaml:"realtime_interval"`
	PollingInterval  int `yaml:"polling_interval"`
	ConfirmBlocks    int `yaml:"confirm_blocks"`
}
type AppConfig struct {
	EthNodes     []NodeConfig `yaml:"eth_nodes"`
	DatabaseDSN  string       `yaml:"database.dsn"`
	APIPort      int          `yaml:"api.port"`
	NFTContracts []string     `yaml:"nft_contracts"`
	Sync         SyncConfig   `yaml:"sync"`
}

var GlobalConfig *AppConfig

func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	GlobalConfig = &cfg
	return GlobalConfig, nil
}
