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
	user := strings.Split(message.Prefix, "!")[0]
	mes := &ChannelMessage{
		Timestamp: time.Now(),
		Sender:    user,
		Channel:   message.Params[0][1:],
		Message:   message.Params[1],
		Metadata:  message.Tags,
	}

	go kafka.WriteChatMessage(makeProtoMessage(mes))
}
