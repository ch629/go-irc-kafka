package nats

import (
	"time"

	"github.com/ch629/go-irc-kafka/pkg/domain"
	"github.com/google/uuid"
)

type (
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
