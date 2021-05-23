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
	"time"
)

type (
	Producer interface {
		SendChatMessage(message domain.ChatMessage)
		SendBan(ban domain.Ban)
		Close() error
		Errors() <-chan *sarama.ProducerError
	}

	producer struct {
		sarama.AsyncProducer
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
	}, nil
}

func (producer *producer) SendChatMessage(message domain.ChatMessage) {
	ts, err := ptypes.TimestampProto(message.Time)
	if err != nil {
		logging.Logger().Warn("Failed to convert time to proto timestamp", zap.Error(err))
		return
	}
	producer.Input() <- &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.chat", message.ChannelName),
		Key:   sarama.StringEncoder(message.UserName),
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

func (producer *producer) SendBan(ban domain.Ban) {
	ts, err := ptypes.TimestampProto(ban.Time)
	if err != nil {
		logging.Logger().Warn("Failed to convert time to proto timestamp", zap.Error(err))
		return
	}
	var banDur uint32
	if ban.BanDuration != nil {
		banDur = uint32(ban.BanDuration.Truncate(time.Second).Seconds())
	}
	var messageId string
	if ban.TargetMessageID != nil {
		messageId = ban.TargetMessageID.String()
	}
	producer.Input() <- &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.bans", ban.ChannelName),
		Key:   sarama.StringEncoder(ban.UserName),
		Value: protoEncoder{
			&pb.Ban{
				ChannelId:       uint32(ban.RoomID),
				TargetUserId:    uint32(ban.TargetUserID),
				ChannelName:     ban.ChannelName,
				TargetUserName:  ban.UserName,
				Timestamp:       ts,
				DurationSeconds: banDur,
				Permanent:       ban.Permanent,
				TargetMessageId: messageId,
			},
		},
	}
}

func mapBadges(badges []domain.Badge) []*pb.Badge {
	b := make([]*pb.Badge, len(badges))
	for i, badge := range badges {
		b[i] = &pb.Badge{
			Name:    badge.Name,
			Version: badge.Version,
		}
	}
	return b
}
