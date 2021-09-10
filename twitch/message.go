package twitch

import (
	"fmt"

	"github.com/ch629/go-irc-kafka/irc/client"
)

type MessageCommand struct {
	Channel string
	Message string
}

func (command MessageCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("PRIVMSG #%v :%v", command.Channel, command.Message))
}

func MakeMessageCommand(channel string, message string) client.IrcMessage {
	return MessageCommand{
		Channel: channel,
		Message: message,
	}
}
