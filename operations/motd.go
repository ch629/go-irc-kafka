package operations

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"go.uber.org/zap"
	"strings"
)

var motd strings.Builder

func handleMotd(handler OperationHandler, message parser.Message) {
	// Trim out the username from the beginning of each MOTD message
	motd.WriteString(fmt.Sprintf("%v\n", strings.TrimPrefix(message.Params[1], handler.botConfig.Name)))
}

func handleMotdEnd(handler OperationHandler, _ parser.Message) {
	handler.log.Info("MOTD", zap.String("message", motd.String()))
	motd.Reset()
}
