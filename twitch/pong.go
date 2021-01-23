package twitch

import (
	"fmt"
	"go-irc/irc/client"
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
