package operations

import (
	"go-irc/irc/parser"
	"go-irc/kafka"
	"strings"
	"time"
)

var producer *kafka.Producer

func initializeProducer() {
	pro, err := kafka.NewDefaultProducer()

	if err != nil {
		panic(err)
	}

	producer = pro
}

type ChannelMessage struct {
	Timestamp time.Time         `json:"timestamp"`
	Sender    string            `json:"sender"`
	Channel   string            `json:"channel"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata"`
}

func handleMessage(message parser.Message) {
	if producer == nil {
		initializeProducer()
	}

	user := strings.Split(message.Prefix, "!")[0]
	mes := &ChannelMessage{
		Timestamp: time.Now(),
		Sender:    user,
		Channel:   message.Params[0][1:],
		Message:   message.Params[1],
		Metadata:  message.Tags,
	}

	producer.WriteChatMessage(makeProtoMessage(mes))
}
