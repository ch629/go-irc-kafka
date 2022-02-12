package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ch629/go-irc-kafka/pkg/irc/parser"
	"github.com/google/uuid"
)

type Ban struct {
	BanDuration     *time.Duration `json:"ban_duration,omitempty"`
	Permanent       bool           `json:"permanent"`
	RoomID          int            `json:"room_id,omitempty"`
	TargetMessageID *uuid.UUID     `json:"target_message_id,omitempty"`
	Time            time.Time      `json:"time"`
	TargetUserID    int            `json:"target_user_id,omitempty"`
	ChannelName     string         `json:"channel_name,omitempty"`
	UserName        string         `json:"user_name,omitempty"`
}

// TODO: Should we be wrapping the lower level errors in this?
func NewBan(message parser.Message) (*Ban, error) {
	tags := message.Tags
	var err error
	b := &Ban{
		ChannelName: message.Params.Channel(),
		UserName:    message.Params[1],
	}
	// Target message ID is optional
	if msgId, hasMsgId := tags["target-msg-id"]; hasMsgId {
		var id uuid.UUID
		if id, err = uuid.Parse(msgId); err != nil {
			return nil, fmt.Errorf("failed to parse message id as uuid: %w", err)
		}
		b.TargetMessageID = &id
	}
	// Ban duration is optional, if not provided it's a permanent ban
	if durString, hasDuration := tags["ban-duration"]; !hasDuration {
		b.Permanent = true
	} else {
		var durSec int
		durSec, err = strconv.Atoi(durString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse duration as seconds int: %w", err)
		}
		dur := time.Duration(durSec) * time.Second
		b.BanDuration = &dur
	}
	if b.RoomID, err = strconv.Atoi(tags["room-id"]); err != nil {
		return nil, fmt.Errorf("failed to parse room-id as int: %w", err)
	}
	if b.Time, err = timeFromTmiSentTs(tags); err != nil {
		return nil, fmt.Errorf("failed to parse timestamp as time: %w", err)
	}
	if b.TargetUserID, err = strconv.Atoi(tags["target-user-id"]); err != nil {
		return nil, fmt.Errorf("failed to parse target user id as int: %w", err)
	}
	return b, nil
}
