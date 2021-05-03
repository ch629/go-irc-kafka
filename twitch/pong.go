package twitch

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc/client"
)

type PongCommand struct {
	Server string
}

func (command *PongCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("PONG :%v", command.Server))
}

func MakePongCommand(server string) client.IrcMessage {
	return &PongCommand{Server: server}
}
