package operations

import (
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/kafka"
	"strings"
	"sync"
	"time"
)

var producer kafka.Producer
var producerInit sync.Once

func initializeProducer(kafkaConfig config.Kafka) {
	pro, err := kafka.NewDefaultProducer(kafkaConfig)

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

// TODO: Temp until message handling rewrite
var con config.Kafka

func InitConfig(kafkaConfig config.Kafka) {
	con = kafkaConfig
}

func handleMessage(message parser.Message) {
	producerInit.Do(func() {
		initializeProducer(con)
	})

	user := strings.Split(message.Prefix, "!")[0]
	mes := &ChannelMessage{
		Timestamp: time.Now(),
		Sender:    user,
		Channel:   message.Params[0][1:],
		Message:   message.Params[1],
		Metadata:  message.Tags,
	}

	producer.Send(makeProtoMessage(mes))
}
