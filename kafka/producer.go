package kafka

import (
	"fmt"
	pb "go-irc/proto"

	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

type Producer struct {
	sarama.AsyncProducer
}

func NewDefaultProducer() (*Producer, error) {
	config := sarama.NewConfig()
	brokers := viper.GetStringSlice("kafka.brokers")

	pro, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create async kafka producer due to %w", err)
	}

	return &Producer{
		AsyncProducer: pro,
	}, nil
}

func (producer *Producer) WriteChatMessage(message *pb.ChatMessage) {
	topic := viper.GetString("kafka.topic")
	producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Value: ProtoEncoder{message},
	}
}
