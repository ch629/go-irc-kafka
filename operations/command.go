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
		log       *zap.Logger
		producer  kafka.Producer
		botConfig config.Bot
	}
)

func MakeOperationHandler(botConfig config.Bot, producer kafka.Producer) *OperationHandler {
	return &OperationHandler{
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
		handler.log.Info("Received users", zap.Strings("users", message.Params))
	},
	"366": func(handler OperationHandler, message parser.Message) {
		// RPL_ENDOFNAMES
		// <channel> :End of /NAMES list
		handler.log.Info("End of names list")
	},
	"JOIN": func(handler OperationHandler, message parser.Message) {
		// Joined channel
		handler.log.Info("Joined channel", zap.Strings("channel", message.Params))
	},
	"421": func(handler OperationHandler, message parser.Message) {
		handler.log.Warn("Invalid command", zap.Strings("command", message.Params))
	},
	"USERSTATE": func(handler OperationHandler, message parser.Message) {
		handler.log.Info("Updating user state", zap.Any("message", message))
	},
	"USERNOTICE": func(handler OperationHandler, message parser.Message) {
		handler.log.Info("User notice", zap.Any("message", message))
	},
	"ROOMSTATE": func(handler OperationHandler, message parser.Message) {
		handler.log.Info("Room state", zap.Any("message", message))
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

// TODO: Should this be called only internally, or do we expose it so somewhere else orchestrates it?
func (h OperationHandler) HandleMessages(channel <-chan parser.Message) {
	go func() {
		for message := range channel {
			if _, ok := ignoredCommands[message.Command]; ok {
				continue
			}
			if f, ok := commandMap[message.Command]; ok {
				go f(h, message)
			} else {
				// Print out message if not known
				h.log.Warn("Received unknown message", zap.String("command", message.Command), zap.Any("message", message))
			}
		}
	}()
}
