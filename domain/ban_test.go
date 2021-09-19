package domain

import (
	"strconv"
	"testing"
	"time"

	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewBan(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		ts := time.Now().Truncate(time.Millisecond)
		id := uuid.New()
		msg := parser.Message{
			Tags: map[string]string{
				"target-msg-id":  id.String(),
				"ban-duration":   "1",
				"room-id":        "2",
				"tmi-sent-ts":    strconv.FormatInt(ts.UnixNano()/int64(time.Millisecond), 10),
				"target-user-id": "3",
			},
			Command: "CLEARCHAT",
			Params:  []string{"#channel", "user"},
		}
		b, err := NewBan(msg)
		assert.NoError(t, err)
		banDur := 1 * time.Second
		assert.Equal(t, Ban{
			BanDuration:     &banDur,
			Permanent:       false,
			RoomID:          2,
			TargetMessageID: &id,
			Time:            ts,
			TargetUserID:    3,
			ChannelName:     "channel",
			UserName:        "user",
		}, *b)
	})
}
