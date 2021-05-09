package main

import (
	"context"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/operations"
	"github.com/ch629/go-irc-kafka/shutdown"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"net"
)

// https://tools.ietf.org/html/rfc1459.html

// TODO: Maybe add a rest endpoint to join/leave a channel or use a kafka topic with commands to handle from external sources
func main() {
	ctx := shutdown.InterruptAwareContext(context.Background())
	log := logging.Logger()
	fs := afero.NewOsFs()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		panic(err)
	}

	producer, err := kafka.NewDefaultProducer(conf.Kafka)
	checkError(err)

	go func() {
		for err := range producer.Errors() {
			log.Error("error from producer", zap.Error(err))
		}
	}()
	operationHandler := operations.MakeOperationHandler(conf.Bot, producer)

loop:
	// Retry connection if we randomly close
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}
		// Connect to IRC
		// For some reason bringing this into a method blocks everything...?
		tcpAddr, err := net.ResolveTCPAddr("tcp4", conf.Irc.Address)
		checkError(err)
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		checkError(err)

		ircClient := client.NewDefaultClient(ctx, conn)
		// Close connection on interrupt

		// Take output from the irc parser & send to handlers
		operationHandler.HandleMessages(ircClient.Input())

		// Handle errors from irc parsing
		go func() {
			for err := range ircClient.Errors() {
				log.Error("error from irc client", zap.Error(err))
			}
		}()

		// Setup output back to IRC
		go operations.OutputStream(ircClient)

		operationHandler.Login()
		<-ircClient.Done()
		log.Info("Client stopped")
		conn.Close()
	}
	checkError(producer.Close())
	log.Info("Finished shutting down")
}

// TODO: Handle errors propagated through this
func checkError(err error) {
	if err != nil {
		log := logging.Logger()
		log.Fatal("Fatal error", zap.Error(err))
	}
}
