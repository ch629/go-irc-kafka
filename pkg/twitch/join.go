package twitch

import (
	"fmt"

	"github.com/ch629/go-irc-kafka/pkg/irc"
	"github.com/ch629/go-irc-kafka/pkg/irc/client"
)

type JoinCommand struct {
	Channel string
}

func (command JoinCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("%v #%v", irc.Join, command.Channel))
}

func MakeJoinCommand(channel string) client.IrcMessage {
	return JoinCommand{Channel: channel}
}
