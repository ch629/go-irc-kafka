package nats

import (
	"encoding/json"
	"fmt"

	"github.com/ch629/go-irc-kafka/pkg/domain"
	"github.com/nats-io/nats.go"
)

type Producer struct {
	con *nats.Conn
}

func NewProducer(url string) (*Producer, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &Producer{
		con: nc,
	}, nil
}

func (p *Producer) SendChatMessage(message domain.ChatMessage) error {
	chatMessage := mapChatMessage(message)
	payload, err := json.Marshal(chatMessage)
	if err != nil {
		return fmt.Errorf("marshalling to json: %w", err)
	}
	return p.con.Publish(
		fmt.Sprintf("%s.chat", message.ChannelName),
		payload,
	)
}

func (p *Producer) SendBan(ban domain.Ban) error {
	banMessage := mapBan(ban)
	payload, err := json.Marshal(banMessage)
	if err != nil {
		return fmt.Errorf("marshalling to json: %w", err)
	}
	return p.con.Publish(
		fmt.Sprintf("%s.bans", ban.ChannelName),
		payload,
	)
}
