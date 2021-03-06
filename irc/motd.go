package irc

import (
	"fmt"
	"go-irc/irc/parser"
	"strings"
)

var motd strings.Builder

func handleMotd(message parser.Message) {
	// Trim out the username from the beginning of each MOTD message
	motd.WriteString(fmt.Sprintf("%v\n", strings.TrimPrefix(message.Params[1], BaseBotConfig.Name)))
}

func handleMotdEnd(_ parser.Message) {
	fmt.Println(motd.String())
	motd.Reset()
}
