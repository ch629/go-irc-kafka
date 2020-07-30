package irc

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	pb "go-irc/proto"
	"net"
	"os"
)

var output = make(chan []byte)

func OutputStream(conn *net.TCPConn) {
	for message := range output {
		fmt.Print("< ", string(message))
		_, err := conn.Write(message)
		checkError(err)
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

func makeProtoMessage(message ChannelMessage) *pb.ChatMessage {
	ts, err := ptypes.TimestampProto(message.Timestamp)
	checkError(err)

	return &pb.ChatMessage{
		Channel:   message.Channel,
		Sender:    message.Sender,
		Message:   message.Message,
		Timestamp: ts,
	}
}

func Login() {
	writeCommand("PASS oauth:%v", BaseBotConfig.OAuthToken)
	writeCommand("NICK %s", BaseBotConfig.Name)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
		os.Exit(1)
	}
}
