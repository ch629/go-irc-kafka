package bot

import (
	"context"
	"errors"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/twitch"
	"go.uber.org/zap"
	"sync"
	"time"
)

type (
	Bot struct {
		ctx       context.Context
		State     *State
		client    client.IrcClient
		onMessage func(*Bot, parser.Message) error
	}
	State struct {
		mux          sync.RWMutex
		Channels     map[string]*Channel
		Capabilities []twitch.Capability
	}
	Channel struct {
		mux   sync.RWMutex
		State *ChannelState
		User  *UserState
	}
	ChannelState struct {
		mux            sync.RWMutex
		EmoteOnly      bool
		FollowerOnly   time.Duration
		R9k            bool
		Slow           time.Duration
		SubscriberOnly bool
	}
	UserState struct {
		mux        sync.RWMutex
		Mod        bool
		Subscriber bool
	}

	// TODO: Is this snapshot idea good? It means we don't have to hold onto locks if we just want to look at what
	//  the channel data looks like at any given time
	ChannelStateSnapshot struct {
		EmoteOnly      bool
		FollowerOnly   time.Duration
		R9k            bool
		Slow           time.Duration
		SubscriberOnly bool
	}
	UserStateSnapshot struct {
		Mod        bool
		Subscriber bool
	}
)

var (
	ErrNotInChannel = errors.New("not in channel")
)

func NewBot(ctx context.Context, client client.IrcClient, onMessage func(bot *Bot, message parser.Message) error) *Bot {
	b := &Bot{
		ctx: ctx,
		State: &State{
			Channels:     make(map[string]*Channel),
			Capabilities: make([]twitch.Capability, 0),
		},
		client:    client,
		onMessage: onMessage,
	}
	go b.handleMessages()
	return b
}

func (b *Bot) handleMessages() {
	for {
		select {
		case <-b.ctx.Done():
			return
		case msg := <-b.client.Input():
			// TODO: Error channel
			// TODO: Try to run these in a goroutine, might need a channel of updates otherwise we get nil
			//  references with joining channels & then updating them
			if err := b.onMessage(b, msg); err != nil {
				logging.Logger().Error("Failed to handle message", zap.String("command", msg.Command), zap.Error(err))
			}
		}
	}
}

func (b Bot) Send(messages ...client.IrcMessage) {
	for _, message := range messages {
		b.client.Output() <- message
	}
}

func (s *ChannelState) Update(emoteOnly, r9k, subscriber bool, follower, slow time.Duration) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.EmoteOnly = emoteOnly
	s.R9k = r9k
	s.SubscriberOnly = subscriber
	s.FollowerOnly = follower
	s.Slow = slow
}

// TODO: Make these funcs run from a Bot instead of state?
func (s *State) UpdateChannel(channel string, emoteOnly, r9k, subscriber bool, follower, slow time.Duration) {
	logging.Logger().Info("Updating channel", zap.String("channel", channel))
	s.mux.RLock()
	defer s.mux.RUnlock()
	s.Channels[channel].State.Update(emoteOnly, r9k, subscriber, follower, slow)
}

func (s *ChannelState) Snapshot() *ChannelStateSnapshot {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return &ChannelStateSnapshot{
		EmoteOnly:      s.EmoteOnly,
		FollowerOnly:   s.FollowerOnly,
		R9k:            s.R9k,
		Slow:           s.Slow,
		SubscriberOnly: s.SubscriberOnly,
	}
}

func (s *UserState) Snapshot() UserStateSnapshot {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return UserStateSnapshot{
		Mod:        s.Mod,
		Subscriber: s.Subscriber,
	}
}

func (s *State) ReadChannelState(channel string) (*ChannelStateSnapshot, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	ch, ok := s.Channels[channel]
	if !ok {
		return nil, ErrNotInChannel
	}
	return ch.State.Snapshot(), nil
}

func (s *State) AddCapability(capability twitch.Capability) {
	s.mux.Lock()
	defer s.mux.Unlock()
	// TODO: Check duplicate
	s.Capabilities = append(s.Capabilities, capability)
}

func (s *UserState) Update(mod, subscriber bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Mod = mod
	s.Subscriber = subscriber
}

func (s *State) UpdateUserState(channel string, mod, subscriber bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	s.Channels[channel].User.Update(mod, subscriber)
}

func (s *State) AddChannel(channel string) *Channel {
	s.mux.Lock()
	defer s.mux.Unlock()
	c := &Channel{
		State: &ChannelState{},
		User:  &UserState{},
	}
	s.Channels[channel] = c
	return c
}

func (s *State) RemoveChannel(channel string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.Channels, channel)
}

func (b Bot) Close() error {
	return b.client.Close()
}
