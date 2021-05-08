package main

import (
	"context"
	_ "embed"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/kafka"
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
	log := logging.Logger()
	fs := afero.NewOsFs()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	conf, err := config.LoadConfig(fs)
	if err != nil {
		panic(err)
	}

	// Connect to IRC
	// For some reason bringing this into a method blocks everything...?
	tcpAddr, err := net.ResolveTCPAddr("tcp4", conf.Irc.Address)
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
			log.Info("Received interrupt")
		case <-ircClient.Done():
			return
		}
	}()

	producer, err := kafka.NewDefaultProducer(conf.Kafka)
	checkError(err)

	operationHandler := operations.MakeOperationHandler(conf.Bot, producer)

	// Take output from the irc parser & send to handlers
	go operationHandler.ReadInput(ircClient.Input())

	// Handle errors from irc parsing
	go func() {
		for err := range ircClient.Errors() {
			select {
			case <-ircClient.Done():
				return
			default:
			}
			log.Errorw("error from irc client", "error", err)
		}
	}()

	go func() {
		for err := range producer.Errors() {
			select {
			case <-ircClient.Done():
				return
			default:
			}
			log.Errorw("error from producer", "error", err)
		}
	}()

	// Setup output back to IRC
	go operations.OutputStream(ircClient)

	operationHandler.Login()
	<-ircClient.Done()
	checkError(producer.Close())
	log.Info("Finished shutting down")
}

// TODO: Handle errors propagated through this
func checkError(err error) {
	if err != nil {
		log := logging.Logger()
		log.Fatalw("Fatal error", err)
	}
}
