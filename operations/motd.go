package operations

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/logging"
	"strings"
)

var motd strings.Builder

func handleMotd(message parser.Message) {
	// Trim out the username from the beginning of each MOTD message
	motd.WriteString(fmt.Sprintf("%v\n", strings.TrimPrefix(message.Params[1], botConfig.Name)))
}

func handleMotdEnd(_ parser.Message) {
	log := logging.Logger()
	log.Infow("MOTD", "message", motd.String())
	motd.Reset()
}
