package irc

import (
	"fmt"
	"go-irc/irc/parser"
	pb "go-irc/proto"
	"os"

	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

type (
	// TODO: Better way of handling these (More customizable messages based on RFC 1459)
	channelMessage struct {
		Command string
		Channel string
	}

	stringMessage struct {
		String string
	}
)

func (message *channelMessage) Bytes() []byte {
	return []byte(fmt.Sprintf("%v #%v", message.Command, message.Channel))
}

func (message *stringMessage) Bytes() []byte {
	return []byte(message.String)
}

var output = make(chan IrcMessage)

func OutputStream(client IrcClient) {
	for message := range output {
		client.Output() <- message
	}
}

func Write(message IrcMessage) {
	output <- message
}

func joinChannel(channel string) {
	Write(&channelMessage{
		Command: "JOIN",
		Channel: channel,
	})
}

func leaveChannel(channel string) {
	Write(&channelMessage{
		Command: "PART",
		Channel: channel,
	})
}

func writeCommand(command string, a ...interface{}) {
	Write(&stringMessage{
		String: fmt.Sprintf(command, a...),
	})
}

func makeProtoMessage(message *ChannelMessage) *pb.ChatMessage {
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
	writeCommand("CAP REQ :twitch.tv/membership")
	writeCommand("CAP REQ :twitch.tv/tags")
	writeCommand("CAP REQ :twitch.tv/commands")
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
