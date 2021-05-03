package operations

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/twitch"
)

func handlePing(message parser.Message) {
	Write(twitch.MakePongCommand(message.Params[0]))
}
