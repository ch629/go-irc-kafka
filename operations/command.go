package operations

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/logging"
)

var log = logging.Logger()

var commandMap = map[string]func(message parser.Message){
	"001":     handleWelcome,
	"PING":    handlePing,
	"ERROR":   handleErrorMessage,
	"PRIVMSG": handleMessage,
	"372":     handleMotd,
	"376":     handleMotdEnd,
	"353": func(message parser.Message) {
		// RPL_NAMREPLY
		//  <channel> :[[@|+]<nick> [[@|+]<nick> [...]]]
		log.Infow("Received users", "users", message.Params)
	},
	"366": func(message parser.Message) {
		// RPL_ENDOFNAMES
		// <channel> :End of /NAMES list
		log.Info("End of names list")
	},
	"JOIN": func(message parser.Message) {
		// Joined channel
		log.Infow("Joined channel", "channel", message.Params)
	},
	"421": func(message parser.Message) {
		log.Warnw("Invalid command", "command", message.Params)
	},
}

var ignoredCommands = map[string]*struct{}{
	// REPL_MOTDSTART
	"375": nil,
}

func ReadInput(channel <-chan parser.Message) {
	for message := range channel {
		if _, ok := ignoredCommands[message.Command]; ok {
			continue
		}
		if f, ok := commandMap[message.Command]; ok {
			go f(message)
		} else {
			// Print out message if not known
			log.Warnw("Received unknown message", "command", message.Command, "message", message)
		}
	}
}
