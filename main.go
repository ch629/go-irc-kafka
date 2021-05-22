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
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"net"
	"strconv"
	"time"
)

// https://tools.ietf.org/html/rfc1459.html

func main() {
	err := startBot()
	if err != nil {
		logging.Logger().Fatal("Failed to start bot", zap.Error(err))
	}
}

func startBot() error {
	ctx := shutdown.InterruptAwareContext(context.Background())
	graceful := &shutdown.GracefulShutdown{}
	fs := afero.NewOsFs()
	log := logging.Logger()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", conf.Irc.Address)
	if err != nil {
		return fmt.Errorf("failed to resolve TCP Addr %w", err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("failed to Dial TCP %w", err)
	}

	ircClient := client.NewDefaultClient(ctx, conn)
	graceful.RegisterWait(ircClient)

	producer, err := kafka.NewDefaultProducer(conf.Kafka)
	if err != nil {
		return fmt.Errorf("failed to create producer %w", err)
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
			bot.Send(twitch.MakePongCommand(message.Params[0]))
		// 366 - End of MOTD -> Last message that is received after connecting to IRC
		case irc.EndOfMOTD:
			log.Debug("Ready to join channels")
			bot.Send(twitch.MakeCapabilityRequest(twitch.TAGS), twitch.MakeCapabilityRequest(twitch.COMMANDS))
			for _, channel := range conf.Bot.Channels {
				bot.Send(twitch.MakeJoinCommand(channel))
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
			} else {
				channelData, err := bot.GetChannelData(channel)
				if err != nil {
					log.Warn("Error getting channel data", zap.Error(err))
				} else {
					log.Info("Updated channel state", zap.Any("state", channelData))
				}
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
			log.Debug("PRIVMSG", zap.Any("message", msg))
			producer.Send(msg)
		case irc.UserNotice:
			// Sub, Resub etc
			messageId := message.Tags.GetOrDefault("msg-id", "")
			switch messageId {
			case "sub":
				var sub inbound.SubTags
				if err := mapstructure.Decode(message.Tags, &sub); err != nil {
					log.Error("failed to decode tags to sub", zap.Any("tags", message.Tags), zap.Error(err))
				}
				log.Info("User subscribed", zap.Any("tags", sub))
			//log.Info("User subscribed", zap.String("user", message.Tags.GetOrDefault("display-name", "")), zap.String("channel", message.Params.Channel()), zap.Any("params", message.Params))
			default:
				log.Info("User notice", zap.Any("message", message))
			}
		case irc.ClearChat, irc.ClearMessage:
			log.Info("Received", zap.String("command", message.Command), zap.Any("params", message.Params))
		case irc.Notice, irc.HostTarget:
			log.Info("Received", zap.Any("message", message))
		// Ignored messages
		case "001", "002", "003", "004", "375", "372", "353", "366":
		default:
			log.Warn("Received unhandled message", zap.String("command", message.Command), zap.Any("message", message))
		}
		return nil
	})
	defer b.Close()
	b.Send(twitch.MakePassCommand(conf.Bot.OAuth), twitch.MakeNickCommand(conf.Bot.Name))
	graceful.Wait()
	return nil
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

func makeStructFromMap(data map[string]string) *structpb.Struct {
	structMap := make(map[string]*structpb.Value, len(data))
	for k, v := range data {
		structMap[k] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: v,
			},
		}
	}
	return &structpb.Struct{
		Fields: structMap,
	}
}
