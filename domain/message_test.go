package domain

import (
	"strconv"
	"testing"
	"time"

	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewBadge(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		b, err := NewBadge("subscriber/6")
		assert.NoError(t, err)
		assert.Equal(t, Badge{
			Name:    "subscriber",
			Version: "6",
		}, b)
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := NewBadge("subscriber")
		assert.ErrorIs(t, err, ErrInvalidBadge)
	})

	t.Run("Empty", func(t *testing.T) {
		b, err := NewBadge("")
		assert.NoError(t, err)
		assert.Empty(t, b)
	})
}

func TestNewBadges(t *testing.T) {
	t.Run("Multiple", func(t *testing.T) {
		bs, err := NewBadges("subscriber/6,turbo/1")
		assert.NoError(t, err)
		assert.Equal(t, []Badge{
			{
				Name:    "subscriber",
				Version: "6",
			},
			{
				Name:    "turbo",
				Version: "1",
			},
		}, bs)
	})

	t.Run("Empty", func(t *testing.T) {
		bs, err := NewBadges("")
		assert.NoError(t, err)
		assert.Empty(t, bs)
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := NewBadge("subscriber")
		assert.ErrorIs(t, err, ErrInvalidBadge)
	})
}

func TestMakeChatMessage(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		ts := time.Now().Truncate(time.Millisecond)
		id := uuid.New()
		msg := parser.Message{
			Tags: map[string]string{
				"display-name": "user",
				"mod":          "1",
				"id":           id.String(),
				"tmi-sent-ts":  strconv.FormatInt(ts.UnixNano()/int64(time.Millisecond), 10),
				"user-id":      "1",
				"room-id":      "2",
				"badges":       "subscriber/3",
			},
			Prefix:  "",
			Command: "PRIVMSG",
			Params: []string{
				"#channel",
				"message",
			},
		}
		chatMessage, err := MakeChatMessage(msg)
		assert.NoError(t, err)
		assert.Equal(t, ChatMessage{
			ID:          id,
			ChannelName: "channel",
			UserName:    "user",
			Message:     "message",
			Time:        ts,
			UserID:      1,
			ChannelID:   2,
			Mod:         true,
			Badges:      []Badge{{"subscriber", "3"}},
		}, chatMessage)
	})
}
