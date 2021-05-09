package bot

import (
	"github.com/ch629/go-irc-kafka/twitch"
	"time"
)

type (
	State struct {
		Channels     map[string]*ChannelState
		Capabilities []twitch.Capability
	}
	ChannelState struct {
		EmoteOnly      bool
		FollowerOnly   time.Duration
		R9k            bool
		Slow           time.Duration
		SubscriberOnly bool
	}
	UserState struct {
		Mod        bool
		Subscriber bool
	}
)
