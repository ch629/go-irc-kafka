package domain

import (
	"errors"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
)

type (
	ChatMessage struct {
		// ID is the unique ID of the message sent
		ID uuid.UUID
		// ChannelName is the name of the channel which the message was sent in
		ChannelName string
		// UserName is the name of the user who sent the message
		UserName string
		// Message is the actual message text
		Message string
		// Time is the time that the IRC server received the message
		Time time.Time
		// UserID is the string ID of the user
		UserID int
		// ChannelID is the ID of the Channel
		ChannelID int
		// Mod is whether the user is a Moderator
		Mod bool
		// Badges is the badges the user has assigned
		// TODO: Should this be a map instead?
		Badges []Badge
	}

	Badge struct {
		Name    string
		Version string
	}
)

var (
	ErrInvalidBadge = errors.New("badge provided was invalid")
)

func NewBadge(name string) (b Badge, err error) {
	if len(name) == 0 {
		return
	}
	split := strings.SplitN(name, "/", 2)
	if len(split) < 2 {
		return b, ErrInvalidBadge
	}
	b.Name = split[0]
	b.Version = split[1]
	return
}

func NewBadges(name string) (b []Badge, err error) {
	if len(name) == 0 {
		return
	}
	split := strings.Split(name, ",")
	b = make([]Badge, len(split))
	for i := range b {
		if b[i], err = NewBadge(split[i]); err != nil {
			return
		}
	}
	return
}

func MakeChatMessage(message parser.Message) (c ChatMessage, err error) {
	tags := message.Tags
	c = ChatMessage{
		ChannelName: message.Params.Channel(),
		UserName:    tags["display-name"],
		Message:     message.Params[1],
		Mod:         tags["mod"] == "1",
	}
	if c.ID, err = uuid.Parse(tags["id"]); err != nil {
		return
	}
	if c.Time, err = timeFromTmiSentTs(tags); err != nil {
		return
	}
	if c.UserID, err = strconv.Atoi(tags["user-id"]); err != nil {
		return
	}
	if c.ChannelID, err = strconv.Atoi(tags["room-id"]); err != nil {
		return
	}
	c.Badges, err = NewBadges(tags["badges"])
	return
}
