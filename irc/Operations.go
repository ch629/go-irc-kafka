package irc

import (
	"fmt"
	"go-irc/parser"
	pb "go-irc/proto"
	"os"

	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

var output = make(chan []byte)

func OutputStream(client IrcClient) {
	for message := range output {
		client.Output() <- message
	}
}

func Write(message string) {
	output <- []byte(fmt.Sprintf("%v\r\n", message))
}

func joinChannel(channel string) {
	writeCommand("JOIN #%v", channel)
}

func leaveChannel(channel string) {
	writeCommand("PART #%v", channel)
}

func writeCommand(command string, a ...interface{}) {
	Write(fmt.Sprintf(command, a...))
}

func makeProtoMessage(message *ChannelMessage) *pb.ChatMessage {
	ts, err := ptypes.TimestampProto(message.Timestamp)
	checkError(err)

	return &pb.ChatMessage{
		Channel:   message.Channel,
		Sender:    message.Sender,
		Message:   message.Message,
		Timestamp: ts,
		Metadata:  mapMapToStruct(message.Metadata),
	}
}

func mapMapToStruct(data map[string]string) *structpb.Struct {
	var structMap = make(map[string]*structpb.Value)
	for k, v := range data {
		if len(v) > 0 {
			structMap[k] = &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: v,
				},
			}
		}
	}
	return &structpb.Struct{
		Fields: structMap,
	}
}

func Login() {
	writeCommand("PASS oauth:%v", BaseBotConfig.OAuthToken)
	writeCommand("NICK %s", BaseBotConfig.Name)
	writeCommand("CAP REQ :twitch.tv/tags")
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
		os.Exit(1)
	}
}

func handleWelcome(_ parser.Message) {
	for _, channel := range BaseBotConfig.Channels {
		joinChannel(channel)
	}
}
