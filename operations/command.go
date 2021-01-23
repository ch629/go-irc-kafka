package operations

import (
	"encoding/json"
	"fmt"
	"go-irc/irc/parser"
	"go-irc/kafka"
)

var producer kafka.Producer

var commandMap = map[string]func(message parser.Message){
	"001":     handleWelcome,
	"PING":    handlePing,
	"ERROR":   handleErrorMessage,
	"PRIVMSG": handleMessage(producer),
	"372":     handleMotd,
	"376":     handleMotdEnd,
	"353": func(message parser.Message) {
		// RPL_NAMREPLY
		//  <channel> :[[@|+]<nick> [[@|+]<nick> [...]]]
		fmt.Println("Got users: ", message.Params)
	},
	"366": func(message parser.Message) {
		// RPL_ENDOFNAMES
		// <channel> :End of /NAMES list
		fmt.Println("End of names list")
	},
	"JOIN": func(message parser.Message) {
		// Joined channel
		fmt.Println("Joined channel: ", message.Params)
	},
	"421": func(message parser.Message) {
		fmt.Println("Invalid command: ", message.Params)
	},
}

var ignoredCommands = map[string]*struct{}{
	// REPL_MOTDSTART
	"375": nil,
}

func ReadInput() {
	pro, err := kafka.NewDefaultProducer()

	if err != nil {
		panic(err)
	}

	producer = *pro

	go func() {
		for message := range parser.Output {
			if _, ok := ignoredCommands[message.Command]; ok {
				continue
			}
			if f, ok := commandMap[message.Command]; ok {
				go f(message)
			} else {
				// Print out message if not known
				bytes, _ := json.Marshal(message)
				fmt.Printf("Message: %v\n", string(bytes))
			}
		}
	}()
}
