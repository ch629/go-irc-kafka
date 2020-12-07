package irc

import (
	"fmt"
	"go-irc/parser"
)

func handlePing(message parser.OldMessage) {
	Write(fmt.Sprintf("PONG :%v", message.Args[0]))
}
