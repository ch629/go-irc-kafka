package kafka

import (
	"github.com/golang/protobuf/proto"
	pb "go-irc/proto"
)

var kafkaOutput = make(chan []byte)

// TODO: Make sure to handle errors correctly
func WriteChatMessage(message *pb.ChatMessage) {
	bytes, err := proto.Marshal(message)
	checkError(err)
	kafkaOutput <- bytes
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
