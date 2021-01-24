package kafka

import (
	"fmt"
	pb "go-irc/proto"

	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

type Producer interface {
	sarama.AsyncProducer
	WriteChatMessage(message *pb.ChatMessage)
	Close() error
}

type producer struct {
	sarama.AsyncProducer
	topic string
}

func NewDefaultProducer() (Producer, error) {
	config := sarama.NewConfig()
	brokers := viper.GetStringSlice("kafka.brokers")

	pro, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create async kafka producer due to %w", err)
	}

	return &producer{
		AsyncProducer: pro,
		topic:         viper.GetString("kafka.topic"),
	}, nil
}

func (producer *producer) WriteChatMessage(message *pb.ChatMessage) {
	producer.Input() <- &sarama.ProducerMessage{
		Topic: producer.topic,
		Key:   sarama.StringEncoder(message.Channel),
		Value: ProtoEncoder{message},
	}
}
