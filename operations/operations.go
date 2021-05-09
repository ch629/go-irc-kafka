package operations

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	pb "github.com/ch629/go-irc-kafka/proto"
	"github.com/ch629/go-irc-kafka/twitch"
	"os"
	"sync"

	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

var output = make(chan client.IrcMessage)
var outputClosed = false
var outputMux sync.Mutex

func OutputStream(client client.IrcClient) {
	for message := range output {
		select {
		case <-client.Done():
			outputMux.Lock()
			outputClosed = true
			close(output)
			outputMux.Unlock()
			return
		default:
		}
		client.Output() <- message
	}
}

func Write(message client.IrcMessage) {
	outputMux.Lock()
	defer outputMux.Unlock()
	if !outputClosed {
		output <- message
	}
}

func joinChannel(channel string) {
	Write(twitch.MakeJoinCommand(channel))
}

func leaveChannel(channel string) {
	Write(twitch.MakePartCommand(channel))
}

func requestCapability(cap twitch.Capability) {
	Write(twitch.MakeCapabilityRequest(cap))
}

func makeProtoMessage(message ChannelMessage) *pb.ChatMessage {
	ts, err := ptypes.TimestampProto(message.Timestamp)
	checkError(err)

	return &pb.ChatMessage{
		Channel:   message.Channel,
		Sender:    message.Sender,
		Message:   message.Message,
		Timestamp: ts,
		Metadata:  makeStruct(message.Metadata),
	}
}

func makeStruct(data map[string]string) *structpb.Struct {
	structMap := make(map[string]*structpb.Value, len(data))
	for k, v := range data {
		structMap[k] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: v,
			},
		}
	}
	return &structpb.Struct{
		Fields: structMap,
	}
}

func (h OperationHandler) Login() {
	Write(twitch.MakePassCommand(h.botConfig.OAuth))
	Write(twitch.MakeNickCommand(h.botConfig.Name))
	requestCapability(twitch.TAGS)
	requestCapability(twitch.COMMANDS)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
		os.Exit(1)
	}
}

func handleWelcome(handler OperationHandler, _ parser.Message) {
	for _, channel := range handler.botConfig.Channels {
		joinChannel(channel)
	}
}
