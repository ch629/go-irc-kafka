package twitch

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc/client"
)

type JoinCommand struct {
	Channel string
}

func (command *JoinCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("JOIN #%v", command.Channel))
}

func MakeJoinCommand(channel string) client.IrcMessage {
	return &JoinCommand{Channel: channel}
}
