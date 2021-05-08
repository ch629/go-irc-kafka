package kafka

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/config"
	pb "github.com/ch629/go-irc-kafka/proto"

	"github.com/Shopify/sarama"
)

type (
	Producer interface {
		Send(message *pb.ChatMessage)
		Close() error
		Errors() <-chan *sarama.ProducerError
	}

	producer struct {
		sarama.AsyncProducer
		topic string
	}
)

func NewDefaultProducer(kafkaConfig config.Kafka) (Producer, error) {
	saramaConfig := sarama.NewConfig()
	brokers := kafkaConfig.Brokers
	saramaConfig.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	saramaConfig.Producer.Compression = sarama.CompressionSnappy

	pro, err := sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create async kafka producer due to %w", err)
	}

	return &producer{
		AsyncProducer: pro,
		topic:         kafkaConfig.Topic,
	}, nil
}

func (producer *producer) Send(message *pb.ChatMessage) {
	producer.Input() <- &sarama.ProducerMessage{
		Topic: producer.topic,
		Key:   sarama.StringEncoder(message.Channel),
		Value: protoEncoder{message},
	}
}
