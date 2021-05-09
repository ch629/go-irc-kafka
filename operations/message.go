package operations

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
	"strings"
	"time"
)

type ChannelMessage struct {
	Timestamp time.Time         `json:"timestamp"`
	Sender    string            `json:"sender"`
	Channel   string            `json:"channel"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func handleMessage(handler OperationHandler, message parser.Message) {
	user := strings.Split(message.Prefix, "!")[0]
	mes := ChannelMessage{
		Timestamp: time.Now(),
		Sender:    user,
		Channel:   message.Params[0][1:],
		Message:   message.Params[1],
		Metadata:  message.Tags,
	}

	handler.producer.Send(makeProtoMessage(mes))
}
