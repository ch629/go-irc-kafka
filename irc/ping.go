package irc

import "go-irc/irc/parser"

func handlePing(message parser.Message) {
	writeCommand("PONG :%v", message.Params[0])
}
