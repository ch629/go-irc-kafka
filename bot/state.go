package bot

import "github.com/ch629/go-irc-kafka/twitch"

type (
	State struct {
		Channels     map[string]*ChannelState
		Capabilities []twitch.Capability
	}
	ChannelState struct {
		EmoteOnly      bool
		FollowerOnly   int
		R9k            bool
		Slow           int
		SubscriberOnly bool
	}
	UserState struct {
		Mod        bool
		Subscriber bool
	}
)
