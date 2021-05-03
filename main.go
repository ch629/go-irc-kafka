package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/operations"
	"github.com/dimiro1/banner"
	"github.com/spf13/afero"
	"net"
	"os"
	"os/signal"
	"strings"
)

//go:embed banner.tmpl
var bannerTmpl string

// https://tools.ietf.org/html/rfc1459.html

// TODO: Maybe add a rest endpoint to join/leave a channel or use a kafka topic with commands to handle from external sources
func main() {
	banner.Init(os.Stderr, true, false, strings.NewReader(bannerTmpl))
	fs := afero.NewOsFs()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	con, err := config.LoadConfig(fs)
	if err != nil {
		panic(err)
	}

	log := logging.Logger()
	client.InitializeConfig(con.Bot)
	operations.InitConfig(con.Kafka)

	// Reads entire message objects created by the parser
	operations.ReadInput()

	// Connect to IRC
	// For some reason bringing this into a method blocks everything...?
	tcpAddr, err := net.ResolveTCPAddr("tcp4", client.BaseBotConfig.Address)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	defer conn.Close()

	ircClient := client.NewDefaultClient(context.Background(), conn)
	// Close connection on interrupt
	go func() {
		select {
		case <-signals:
			ircClient.Close()
		case <-ircClient.Done():
			return
		}
	}()

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
	<-ircClient.Done()
}

// TODO: Handle errors propagated through this
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
		os.Exit(1)
	}
}
