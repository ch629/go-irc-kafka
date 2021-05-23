package main

import (
	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/domain"
	"github.com/ch629/go-irc-kafka/irc"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/twitch"
	"github.com/ch629/go-irc-kafka/twitch/inbound"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type botMessageHandler struct {
	conf     config.Config
	log      *zap.Logger
	producer kafka.Producer
}

func (h *botMessageHandler) handleMessage(bot *bot.Bot, message parser.Message) (err error) {
	conf := h.conf
	log := h.log
	producer := h.producer
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
