package irc

import (
	"fmt"
	"go-irc/parser"
)

func handlePing(message parser.NewMessage) {
	Write(fmt.Sprintf("PONG :%v", message.Args[0]))
}
