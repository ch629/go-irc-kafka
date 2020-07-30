package irc

import (
	"encoding/json"
	"fmt"
	"go-irc/kafka"
	"go-irc/parser"
	"strings"
	"time"
)

type ChannelMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
	Channel   string    `json:"channel"`
	Message   string    `json:"message"`
}

func handleMessage(message parser.Message) {
	user := strings.Split(message.Prefix, "!")[0]
	text := message.Params
	split := strings.Split(text, " ")

	channel := strings.TrimPrefix(split[0], "#")
	mes := strings.TrimPrefix(strings.Join(split[1:], " "), channel+" ")

	chanMes := ChannelMessage{
		Timestamp: message.Timestamp,
		Sender:    user,
		Channel:   channel,
		Message:   mes,
	}
	// TODO: Remove this logic eventually
	messageJson, _ := json.Marshal(chanMes)
	fmt.Printf("ChannelMessage: %v\n", string(messageJson))

	go kafka.WriteChatMessage(makeProtoMessage(chanMes))
}
