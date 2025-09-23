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
	OrderInterval    int `yaml:"order_interval"`
}
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
type AppConfig struct {
	EthNodes        []NodeConfig          `yaml:"eth_nodes"`
	DatabaseDSN     string                `yaml:"database.dsn"`
	APIPort         int                   `yaml:"api.port"`
	NFTContracts    []string              `yaml:"nft_contracts"`
	OrderContracts  []string              `yaml:"order_contracts"`
	Sync            SyncConfig            `yaml:"sync"`
	Redis           RedisConfig           `yaml:"redis"`
	FloorPriceKafka FloorPriceKafkaConfig `yaml:"floor_price_kafka"`
}

type NotifyConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	MQTopic    string `yaml:"mq_topic"`
}

type FloorPriceKafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
