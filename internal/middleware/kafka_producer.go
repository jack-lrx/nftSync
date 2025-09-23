package middleware

import (
	"github.com/IBM/sarama"
	"log"
)

// KafkaProducer 封装 Kafka 生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaProducer 创建 Kafka 生产者
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{producer: producer, topic: topic}, nil
}

// SendFloorPriceUpdateMsg 发送地板价更新消息
func (kp *KafkaProducer) SendFloorPriceUpdateMsg(collection string) error {
	msg := &sarama.ProducerMessage{
		Topic: kp.topic,
		Value: sarama.StringEncoder(collection),
	}
	_, _, err := kp.producer.SendMessage(msg)
	if err != nil {
		log.Printf("[kafka] 地板价消息发送失败: %v", err)
	}
	return err
}

// Close 关闭生产者
func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}
