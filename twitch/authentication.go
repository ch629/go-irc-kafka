package twitch

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc"
	"github.com/ch629/go-irc-kafka/irc/client"
)

type (
	PassCommand struct {
		OAuth string
	}
	NickCommand struct {
		Name string
	}
)

func (command PassCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("%v oauth:%v", irc.Password, command.OAuth))
}

func (command NickCommand) Bytes() []byte {
	return []byte(fmt.Sprintf("%v %v", irc.Nickname, command.Name))
}

func MakePassCommand(oauth string) client.IrcMessage {
	return PassCommand{OAuth: oauth}
}

func MakeNickCommand(name string) client.IrcMessage {
	return NickCommand{Name: name}
}
