package twitch

import (
	"fmt"
	"go-irc/irc/client"
)

type (
	PassCommand struct {
		OAuth string
	}
	NickCommand struct {
		Name string
	}
)

func (command *PassCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("PASS oauth:%v", command.OAuth))
}

func (command *NickCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("NICK %v", command.Name))
}

func MakePassCommand(oauth string) client.IrcMessage {
	return &PassCommand{OAuth: oauth}
}

func MakeNickCommand(name string) client.IrcMessage {
	return &NickCommand{Name: name}
}
