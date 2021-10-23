package main

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/domain"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/kafka"
	_ "github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/twitch"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	botClient "github.com/ch629/bot-orchestrator/pkg/client"
)

// https://tools.ietf.org/html/rfc1459.html

func main() {
	log := zap.L()
	defer log.Sync()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fs := afero.NewOsFs()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	ircClient, err := makeIrcClient(ctx, conf.Irc.Address)
	if err != nil {
		log.Fatal("failed to make irc client", zap.Error(err))
	}

	client, err := kafka.NewClient(conf.Kafka)
	if err != nil {
		log.Fatal("failed to create kafka client", zap.Error(err))
	}
	defer client.Close()
	if len(client.Brokers()) == 0 {
		log.Fatal("failed to connect to kafka brokers", zap.Error(err))
	}
	producer, err := kafka.NewProducer(client)
	if err != nil {
		log.Fatal("failed to create producer", zap.Error(err))
	}

	messageHandler := bot.MessageHandler{}

	messageHandler.OnPrivateMessage(func(msg domain.ChatMessage) {
		log.Debug("received private message", zap.Any("msg", msg))
		if err := producer.SendChatMessage(msg); err != nil {
			log.Warn("failed to send chat message", zap.Error(err))
		}
	})
	messageHandler.OnBan(func(ban domain.Ban) {
		log.Debug("received ban message", zap.Any("msg", ban))
		if err := producer.SendBan(ban); err != nil {
			log.Warn("failed to send ban message", zap.Error(err))
		}
	})
	messageHandler.OnError(func(err error) {
		log.Error("err from bot", zap.Error(err))
	})

	bot := bot.New(ircClient, messageHandler)
	log.Info("created bot")

	go bot.ProcessMessages(ctx)
	log.Info("processing messages")
	if err := bot.Login(ctx, conf.Bot.Name, conf.Bot.OAuth); err != nil {
		log.Fatal("error when logging in", zap.Error(err))
	}
	log.Info("logged in successfully")

	if err := bot.RequestCapability(twitch.COMMANDS, twitch.MEMBERSHIP, twitch.TAGS); err != nil {
		log.Fatal("failed to request capabilities", zap.Error(err))
	}

	conn, err := grpc.DialContext(ctx, conf.Orchestrator.Address, grpc.WithInsecure())
	if err != nil {
		log.Fatal("failed to dial grpc", zap.Error(err))
	}
	id, err := botClient.Join(ctx, conn, orchestratorClient{
		logger: log,
		bot:    bot,
	})
	if err != nil {
		log.Fatal("failed to join to orchestrator", zap.Error(err))
	}
	log.Info("joined orchestrator", zap.String("bot_id", id.String()))
	<-ctx.Done()
	log.Info("closing")
}

type orchestratorClient struct {
	logger *zap.Logger
	bot    *bot.Bot
}

func (c orchestratorClient) JoinChannel(channel string) {
	c.logger.Info("joining", zap.String("channel", channel))
	c.bot.JoinChannels(channel)
}

func (c orchestratorClient) LeaveChannel(channel string) {
	c.logger.Info("leaving", zap.String("channel", channel))
	c.bot.LeaveChannels(channel)
}

func (c orchestratorClient) Close() {
	c.logger.Info("closing")
	// TODO: Close bot
}

func makeIrcClient(ctx context.Context, address string) (ircClient client.IrcClient, err error) {
	log := zap.L()
	// Sometimes the client closes instantly, retry it 3 times
	// TODO: Do we still need this?
	timer := time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()
	for i := 0; i < 3; i++ {
		conn, err := makeConnection(address)
		if err != nil {
			return nil, err
		}
		ircClient = client.NewClient(ctx, conn)
		go ircClient.ConsumeMessages()
		timer.Stop()
		timer.Reset(10 * time.Millisecond)
		select {
		case <-ircClient.Done():
			err = ircClient.Err()
			log.Warn("IrcClient exited on startup", zap.Error(err))
			// Make sure the connection is closed if we're retrying
			conn.Close()
			continue
		case <-timer.C:
		}
		break
	}

	if err != nil {
		err = fmt.Errorf("failed to create IrcClient after retries: %w", err)
	}
	return
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
