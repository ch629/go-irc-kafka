package kafka

import (
	pb "go-irc/proto"

	"github.com/Shopify/sarama"
)

// TODO: Configurable topic
func (producer *Producer) WriteChatMessage(message *pb.ChatMessage) {
	producer.Input() <- &sarama.ProducerMessage{
		Topic: "topic",
		Value: ProtoEncoder{message},
	}
}
