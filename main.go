package main

import (
	"context"
	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/proto"
	"github.com/ch629/go-irc-kafka/shutdown"
	"github.com/ch629/go-irc-kafka/twitch"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"net"
	"strconv"
	"strings"
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
	fs := afero.NewOsFs()
	log := logging.Logger()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", conf.Irc.Address)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	ircClient := client.NewDefaultClient(ctx, conn)

	producer, err := kafka.NewDefaultProducer(conf.Kafka)
	if err != nil {
		return err
	}

	go func() {
		for err := range producer.Errors() {
			log.Error("error from producer", zap.Error(err))
		}
	}()

	b := bot.NewBot(ctx, ircClient, func(bot *bot.Bot, message parser.Message) error {
		// TODO: Constants for command types
		switch message.Command {
		case "PING":
			log.Info("PING")
			bot.Send(twitch.MakePongCommand(message.Params[0]))
		// 366 - End of MOTD -> Last message that is received after connecting to IRC
		case "376":
			log.Info("Ready to join channels")
			bot.Send(twitch.MakeCapabilityRequest(twitch.TAGS), twitch.MakeCapabilityRequest(twitch.COMMANDS))
			for _, channel := range conf.Bot.Channels {
				bot.Send(twitch.MakeJoinCommand(channel))
			}
		case "CAP":
			if message.Params[1] == "ACK" {
				capability := twitch.CapabilityFromParam(message.Params[2])
				bot.State.AddCapability(capability)
			} else {
				log.Warn("Received non-ACK capability", zap.Any("message", message))
			}
		case "JOIN":
			channel := message.Params[0][1:]
			bot.State.AddChannel(channel)
			log.Info("Joined channel", zap.String("name", channel))
		case "ROOMSTATE":
			channel := message.Params[0][1:]
			emoteOnly := getOrDefault(message.Tags, "emote-only", "0") == "1"
			r9k := getOrDefault(message.Tags, "r9k", "0") == "1"
			subscriber := getOrDefault(message.Tags, "subs-only", "0") == "1"
			follower := seconds(atoiOrDefault(getOrDefault(message.Tags, "followers-only", "0"), 0))
			slow := seconds(atoiOrDefault(getOrDefault(message.Tags, "slow", "0"), 0))

			bot.State.UpdateChannel(channel, emoteOnly, r9k, subscriber, follower, slow)
		case "USERSTATE":
			channel := message.Params[0][1:]
			mod := getOrDefault(message.Tags, "mod", "0") == "1"
			subscriber := getOrDefault(message.Tags, "subscriber", "0") == "1"
			bot.State.UpdateUserState(channel, mod, subscriber)
		case "PRIVMSG":
			channel := message.Params[0][1:]
			msg := message.Params[1]
			user := strings.SplitN(message.Prefix, "!", 2)[0]
			log.Info("PRIVMSG", zap.String("channel", channel), zap.Any("user", user), zap.String("message", msg))
			producer.Send(&proto.ChatMessage{
				Channel:   channel,
				Sender:    user,
				Message:   msg,
				Timestamp: ptypes.TimestampNow(),
				Metadata:  makeStructFromMap(message.Tags),
			})
		// Ignored messages
		case "001", "002", "003", "004", "375", "372", "353", "366":
		default:
			log.Info("Received", zap.Any("message", message))
		}
		return nil
	})
	defer b.Close()
	b.Send(twitch.MakePassCommand(conf.Bot.OAuth), twitch.MakeNickCommand(conf.Bot.Name))
	<-ircClient.Done()
	return nil
}

func getOrDefault(m map[string]string, key string, def string) (v string) {
	var ok bool
	if v, ok = m[key]; !ok {
		return def
	}
	return
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
