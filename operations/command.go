package operations

import (
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"go.uber.org/zap"
)

type (
	parserMessageHandler = func(handler OperationHandler, message parser.Message)

	OperationHandler struct {
		log       zap.SugaredLogger
		producer  kafka.Producer
		botConfig config.Bot
	}
)

func MakeOperationHandler(botConfig config.Bot, producer kafka.Producer) OperationHandler {
	return OperationHandler{
		log:       logging.Logger(),
		producer:  producer,
		botConfig: botConfig,
	}
}

var commandMap = map[string]parserMessageHandler{
	"001":     handleWelcome,
	"PING":    handlePing,
	"ERROR":   handleErrorMessage,
	"PRIVMSG": handleMessage,
	"372":     handleMotd,
	"376":     handleMotdEnd,
	"353": func(handler OperationHandler, message parser.Message) {
		// RPL_NAMREPLY
		//  <channel> :[[@|+]<nick> [[@|+]<nick> [...]]]
		handler.log.Infow("Received users", "users", message.Params)
	},
	"366": func(handler OperationHandler, message parser.Message) {
		// RPL_ENDOFNAMES
		// <channel> :End of /NAMES list
		handler.log.Info("End of names list")
	},
	"JOIN": func(handler OperationHandler, message parser.Message) {
		// Joined channel
		handler.log.Infow("Joined channel", "channel", message.Params)
	},
	"421": func(handler OperationHandler, message parser.Message) {
		handler.log.Warnw("Invalid command", "command", message.Params)
	},
	"USERSTATE": func(handler OperationHandler, message parser.Message) {
		handler.log.Infow("Updating user state", "message", message)
	},
	"USERNOTICE": func(handler OperationHandler, message parser.Message) {
		handler.log.Infow("User notice", "message", message)
	},
	"ROOMSTATE": func(handler OperationHandler, message parser.Message) {
		handler.log.Infow("Room state", "message", message)
	},
}

var ignoredCommands = map[string]*struct{}{
	// REPL_MOTDSTART
	"375": nil,
	// capability ack
	"CAP":       nil,
	"002":       nil,
	"003":       nil,
	"004":       nil,
	"CLEARCHAT": nil,
}

func (h OperationHandler) ReadInput(channel <-chan parser.Message) {
	for message := range channel {
		if _, ok := ignoredCommands[message.Command]; ok {
			continue
		}
		if f, ok := commandMap[message.Command]; ok {
			go f(h, message)
		} else {
			// Print out message if not known
			h.log.Warnw("Received unknown message", "command", message.Command, "message", message)
		}
	}
}
