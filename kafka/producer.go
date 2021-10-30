package kafka

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockery --name=Producer --disable-version-string
type (
	Producer interface {
		SendChatMessage(message domain.ChatMessage) error
		SendBan(ban domain.Ban) error
		Close() error
	}

	producer struct {
		logger *zap.Logger
		sarama.SyncProducer
	}

	// TODO: Pull these out into some model package?
	chatMessage struct {
		ID          uuid.UUID `json:"id"`
		ChannelName string    `json:"channel_name"`
		UserName    string    `json:"user_name"`
		Message     string    `json:"message"`
		Timestamp   time.Time `json:"timestamp"`
		UserID      int       `json:"user_id"`
		ChannelID   int       `json:"channel_id"`
		Badges      []badge   `json:"badges"`
	}

	badge struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	banMessage struct {
		ChannelID       int            `json:"channel_id"`
		TargetUserID    int            `json:"target_user_id"`
		ChannelName     string         `json:"channel_name"`
		TargetUserName  string         `json:"target_user_name"`
		Timestamp       time.Time      `json:"timestamp"`
		Duration        *time.Duration `json:"duration,omitempty"`
		Permanent       bool           `json:"permanent"`
		TargetMessageID *uuid.UUID     `json:"target_message_id,omitempty"`
	}
)

func NewClient(kafkaConfig config.Kafka) (sarama.Client, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	saramaConfig.Producer.Compression = sarama.CompressionSnappy
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = true
	return sarama.NewClient(kafkaConfig.Brokers, saramaConfig)
}

func NewProducer(client sarama.Client) (Producer, error) {
	pro, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create async kafka producer due to %w", err)
	}

	return &producer{
		SyncProducer: pro,
		logger:       zap.L(),
	}, nil
}

func (producer *producer) SendChatMessage(message domain.ChatMessage) error {
	chatMessage := mapChatMessage(message)
	enc, err := NewJSONEncoder(chatMessage)
	if err != nil {
		return err
	}
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.chat", message.ChannelName),
		Key:   sarama.StringEncoder(message.UserName),
		Value: enc,
	})
	return err
}

func (producer *producer) SendBan(ban domain.Ban) error {
	banMessage := mapBan(ban)
	enc, err := NewJSONEncoder(banMessage)
	if err != nil {
		return err
	}
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.bans", ban.ChannelName),
		Key:   sarama.StringEncoder(ban.UserName),
		Value: enc,
	})
	return err
}

func mapChatMessage(message domain.ChatMessage) chatMessage {
	return chatMessage{
		ID:          message.ID,
		ChannelName: message.ChannelName,
		UserName:    message.UserName,
		Message:     message.Message,
		Timestamp:   message.Time,
		UserID:      message.UserID,
		ChannelID:   message.ChannelID,
		Badges:      mapBadges(message.Badges),
	}
}

func mapBadges(badges []domain.Badge) []badge {
	b := make([]badge, len(badges))
	for i, domainBadge := range badges {
		b[i] = badge{
			Name:    domainBadge.Name,
			Version: domainBadge.Version,
		}
	}
	return b
}

func mapBan(ban domain.Ban) banMessage {
	return banMessage{
		ChannelID:       ban.RoomID,
		TargetUserID:    ban.TargetUserID,
		ChannelName:     ban.ChannelName,
		TargetUserName:  ban.UserName,
		Timestamp:       ban.Time,
		Duration:        ban.BanDuration,
		Permanent:       ban.Permanent,
		TargetMessageID: ban.TargetMessageID,
	}
}
