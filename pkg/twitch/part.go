package twitch

import (
	"fmt"

	"github.com/ch629/go-irc-kafka/pkg/irc/client"
)

type PartCommand struct {
	Channel string
}

func (command PartCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("PART #%v", command.Channel))
}

func MakePartCommand(channel string) client.IrcMessage {
	return PartCommand{Channel: channel}
}
