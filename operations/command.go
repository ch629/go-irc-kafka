package operations

import (
	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"go.uber.org/zap"
	"strconv"
	"time"
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
		channelName := message.Params[0][1:]
		followerSeconds, err := strconv.Atoi(message.Tags["followers-only"])
		if err != nil {
			panic(err)
		}
		slowSeconds, err := strconv.Atoi(message.Tags["slow"])
		if err != nil {
			panic(err)
		}
		state := bot.ChannelState{
			EmoteOnly:      message.Tags["emote-only"] == "1",
			R9k:            message.Tags["r9k"] == "1",
			SubscriberOnly: message.Tags["subs-only"] == "1",
			FollowerOnly:   time.Duration(followerSeconds) * time.Second,
			Slow:           time.Duration(slowSeconds) * time.Second,
		}
		handler.log.Info("Room State", zap.String("channel", channelName), zap.Any("state", state))
	},
	"CAP": func(handler OperationHandler, message parser.Message) {
		handler.log.Info("Capability message received", zap.Any("message", message))
	},
	"CLEARCHAT": func(handler OperationHandler, message parser.Message) {
		handler.log.Info("Chat has been cleared", zap.Any("message", message))
	},
}

var ignoredCommands = map[string]*struct{}{
	// RPL_NAMREPLY
	"353": nil,
	// RPL_ENDOFNAMES
	"366": nil,
	// motd line
	"372": nil,
	// REPL_MOTDSTART
	"375": nil,
	// motd end
	"376": nil,
	"002": nil,
	"003": nil,
	"004": nil,
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
