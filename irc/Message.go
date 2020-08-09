package irc

import (
	"go-irc/kafka"
	"go-irc/parser"
	"strings"
	"time"
)

// TODO: Do we want to split out the metadata a bit? Rather than storing in a map
type ChannelMessage struct {
	Timestamp time.Time         `json:"timestamp"`
	Sender    string            `json:"sender"`
	Channel   string            `json:"channel"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata"`
}

func handleMessage(message parser.Message) {
	user := strings.Split(message.Username, "!")[0]
	prefixes := strings.Split(message.Prefix, " ")

	go kafka.WriteChatMessage(makeProtoMessage(&ChannelMessage{
		Timestamp: message.Timestamp,
		Sender:    user,
		Channel:   prefixes[0][1:],
		Message:   message.Params,
		Metadata:  message.Metadata,
	}))
}
