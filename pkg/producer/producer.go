package producer

import "github.com/ch629/go-irc-kafka/pkg/domain"

//go:generate mockery --name=Producer --disable-version-string
type Producer interface {
	SendChatMessage(message domain.ChatMessage) error
	SendBan(ban domain.Ban) error
	Close() error
}
