package main

import (
	"context"
	"fmt"
	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/domain"
	"github.com/ch629/go-irc-kafka/irc"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/shutdown"
	"github.com/ch629/go-irc-kafka/twitch"
	"github.com/ch629/go-irc-kafka/twitch/inbound"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"net"
	"strconv"
	"time"
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

	b := bot.NewBot(ctx, ircClient, func(bot *bot.Bot, message parser.Message) error {
		switch message.Command {
		case irc.Ping:
			log.Debug("Received PING")
			bot.Pong(message.Params)
		case irc.EndOfMOTD:
			log.Debug("Ready to join channels")
			bot.RequestCapability(twitch.TAGS, twitch.COMMANDS)
			for _, channel := range conf.Bot.Channels {
				bot.RequestJoinChannel(channel)
			}
		case irc.Capability:
			if message.Params[1] == "ACK" {
				capability := twitch.CapabilityFromParam(message.Params[2])
				bot.AddCapability(capability)
			} else {
				log.Warn("Received non-ACK capability", zap.Any("message", message))
			}
		case irc.Join:
			channel := message.Params.Channel()
			bot.AddChannel(channel)
			log.Info("Joined channel", zap.String("name", channel))
		case irc.Part:
			channel := message.Params.Channel()
			bot.RemoveChannel(channel)
			log.Info("Left channel", zap.String("name", channel))
		case irc.RoomState:
			tags := message.Tags
			channel := message.Params.Channel()
			emoteOnly := tags.GetOrDefault("emote-only", "0") == "1"
			r9k := tags.GetOrDefault("r9k", "0") == "1"
			subscriber := tags.GetOrDefault("subs-only", "0") == "1"
			follower := seconds(atoiOrDefault(tags.GetOrDefault("followers-only", "0"), 0))
			slow := seconds(atoiOrDefault(tags.GetOrDefault("slow", "0"), 0))

			if err = bot.UpdateChannel(channel, emoteOnly, r9k, subscriber, follower, slow); err != nil {
				log.Warn("Failed to update channel", zap.String("channel", channel), zap.Error(err))
			}
		case irc.UserState:
			channel := message.Params.Channel()
			mod := message.Tags.GetOrDefault("mod", "0") == "1"
			subscriber := message.Tags.GetOrDefault("subscriber", "0") == "1"
			if err = bot.UpdateUserState(channel, mod, subscriber); err != nil {
				log.Warn("Failed to update user state", zap.String("channel", channel), zap.Error(err))
			}
		case irc.PrivateMessage:
			msg, err := domain.MakeChatMessage(message)
			if err != nil {
				log.Warn("Failed to make chat message", zap.Error(err))
			}
			producer.SendChatMessage(msg)
		case irc.UserNotice:
			// Sub, Resub etc
			messageId := message.Tags.GetOrDefault("msg-id", "")
			switch messageId {
			case "sub":
				var sub inbound.SubTags
				if err = mapstructure.Decode(message.Tags, &sub); err != nil {
					log.Error("failed to decode tags to sub", zap.Any("tags", message.Tags), zap.Error(err))
				}
			default:
				log.Info("User notice", zap.Any("message", message))
			}
		case irc.ClearChat:
			ban, err := domain.NewBan(message)
			if err != nil {
				log.Warn("Error creating ban message", zap.Any("message", message), zap.Error(err))
			} else {
				producer.SendBan(ban)
				log.Debug("User banned", zap.Any("ban", ban))
			}
		case irc.ClearMessage:
			log.Info("Received", zap.String("command", message.Command), zap.Any("message", message))
		case irc.Notice, irc.HostTarget:
			log.Info("Received", zap.Any("message", message))
		// Ignored messages
		case "001", "002", "003", "004", "375", "372", "353", "366":
		default:
			log.Warn("Received unhandled message", zap.String("command", message.Command), zap.Any("message", message))
		}
		return nil
	})
	b.Login(conf.Bot.Name, conf.Bot.OAuth)
	return b, nil
}

func atoiOrDefault(v string, def int) (i int) {
	var err error
	if i, err = strconv.Atoi(v); err != nil {
		return def
	}
	return
}

func seconds(s int) time.Duration {
	return time.Duration(s) * time.Second
}
