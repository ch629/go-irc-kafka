package irc

import (
	"fmt"
	"go-irc/parser"
)

func handlePing(message parser.Message) {
	Write(fmt.Sprintf("PONG :%v", message.Args[0]))
}
