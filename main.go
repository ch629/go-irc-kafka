package main

import (
	"context"
	"fmt"
	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/shutdown"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"net"
)

// https://tools.ietf.org/html/rfc1459.html

func main() {
	ctx := shutdown.InterruptAwareContext(context.Background())
	graceful := &shutdown.GracefulShutdown{}
	b, err := startBot(ctx, graceful)
	if err != nil {
		logging.Logger().Fatal("Failed to start bot", zap.Error(err))
	}
	defer b.Close()
	graceful.Wait()
}

func startBot(ctx context.Context, graceful *shutdown.GracefulShutdown) (*bot.Bot, error) {
	fs := afero.NewOsFs()
	log := logging.Logger()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		return nil, err
	}

	conn, err := makeConnection(conf.Irc.Address)
	if err != nil {
		return nil, err
	}

	ircClient := client.NewClient(ctx, conn)
	graceful.RegisterWait(ircClient)

	producer, err := kafka.NewProducer(conf.Kafka)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer %w", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Error("error from producer", zap.Error(err))
		}
	}()

	handler := botMessageHandler{
		conf:     conf,
		log:      logging.Logger(),
		producer: producer,
	}
	b := bot.NewBot(ctx, ircClient, handler.handleMessage)
	b.Login(conf.Bot.Name, conf.Bot.OAuth)
	return b, nil
}

func makeConnection(address string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP Addr %w", err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to Dial TCP %w", err)
	}
	return conn, nil
}
