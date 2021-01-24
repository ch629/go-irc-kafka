package main

import (
	"fmt"
	"go-irc/config"
	"go-irc/irc/client"
	"go-irc/irc/parser"
	"go-irc/logging"
	"go-irc/operations"
	"net"
	"os"
	"sync"
)

// https://tools.ietf.org/html/rfc1459.html

// TODO: Maybe add a rest endpoint to join/leave a channel or use a kafka topic with commands to handle from external sources
func main() {
	log := logging.Logger()
	// TODO: Handle this WaitGroup better
	wg := sync.WaitGroup{}
	wg.Add(1)
	config.LoadConfig()

	client.InitializeConfig()

	// Reads entire message objects created by the parser
	operations.ReadInput()

	// Connect to IRC
	// For some reason bringing this into a method blocks everything...?
	tcpAddr, err := net.ResolveTCPAddr("tcp4", client.BaseBotConfig.Address)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	defer conn.Close()

	ircClient := client.NewDefaultClient(conn)

	// Take output from the irc parser & send to handlers
	go func() {
		for input := range ircClient.Input() {
			parser.Output <- input
		}
	}()

	// Handle errors from irc parsing
	go func() {
		for err := range ircClient.Errors() {
			log.Errorw("error from irc client", "error", err)
		}
	}()

	// Setup output back to IRC
	go operations.OutputStream(ircClient)

	operations.Login()

	wg.Wait()
}

// TODO: Handle errors propagated through this
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
		os.Exit(1)
	}
}
