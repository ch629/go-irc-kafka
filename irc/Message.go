package irc

import (
	"encoding/json"
	"fmt"
	"go-irc/kafka"
	"go-irc/parser"
	"strings"
	"time"
)

// TODO: Include some of the tag data here (And in the protobuf object? - Could just be the map/struct)
type ChannelMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
	Channel   string    `json:"channel"`
	Message   string    `json:"message"`
}

func handleMessage(message parser.Message) {
	user := strings.Split(message.Username, "!")[0]
	prefixes := strings.Split(message.Prefix, " ")

	chanMes := ChannelMessage{
		Timestamp: message.Timestamp,
		Sender:    user,
		Channel:   prefixes[0][1:],
		Message:   message.Params,
	}
	// TODO: Remove this logic eventually
	messageJson, _ := json.Marshal(chanMes)
	fmt.Printf("ChannelMessage: %v\n", string(messageJson))

	go kafka.WriteChatMessage(makeProtoMessage(chanMes))
}
