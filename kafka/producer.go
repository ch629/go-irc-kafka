package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/domain"
	"github.com/ch629/go-irc-kafka/logging"
	pb "github.com/ch629/go-irc-kafka/proto"
	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
)

type (
	Producer interface {
		Send(message domain.ChatMessage)
		Close() error
		Errors() <-chan *sarama.ProducerError
	}

	producer struct {
		sarama.AsyncProducer
		topic string
	}
)

func NewProducer(kafkaConfig config.Kafka) (Producer, error) {
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

func (producer *producer) Send(message domain.ChatMessage) {
	ts, err := ptypes.TimestampProto(message.Time)
	if err != nil {
		logging.Logger().Warn("Failed to convert time to proto timestamp", zap.Error(err))
		return
	}
	producer.Input() <- &sarama.ProducerMessage{
		Topic: producer.topic,
		Key:   sarama.StringEncoder(message.ChannelName),
		Value: protoEncoder{
			&pb.ChatMessage{
				Id:          message.ID.String(),
				ChannelName: message.ChannelName,
				UserName:    message.UserName,
				Message:     message.Message,
				Timestamp:   ts,
				UserId:      uint32(message.UserID),
				ChannelId:   uint32(message.ChannelID),
				Badges:      mapBadges(message.Badges),
			},
		},
	}
}

func mapBadges(badges []domain.Badge) []*pb.Badge {
	b := make([]*pb.Badge, len(badges))
	for i := range b {
		b[i] = &pb.Badge{
			Name:    badges[i].Name,
			Version: uint32(badges[i].Version),
		}
	}
	return b
}
