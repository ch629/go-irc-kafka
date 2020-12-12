package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
)

type Producer struct {
	sarama.AsyncProducer
}

// TODO: Pull in viper config
func NewDefaultProducer() (*Producer, error) {
	config := sarama.NewConfig()

	pro, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create async kafka producer due to %w", err)
	}

	return &Producer{
		AsyncProducer: pro,
	}, nil
}
