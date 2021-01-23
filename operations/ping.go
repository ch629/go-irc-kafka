package operations

import (
	"go-irc/irc/parser"
	"go-irc/twitch"
)

func handlePing(message parser.Message) {
	Write(twitch.MakePongCommand(message.Params[0]))
}
