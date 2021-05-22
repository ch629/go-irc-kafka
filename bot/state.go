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
	// TODO: Interface
	Bot struct {
		ctx       context.Context
		State     *State
		client    client.IrcClient
		onMessage func(*Bot, parser.Message) error
	}
	State struct {
		mux sync.RWMutex
		// TODO: Should these be exported?
		Channels     map[string]*Channel
		Capabilities []twitch.Capability
	}
	Channel struct {
		mux   sync.RWMutex
		State ChannelState
		User  UserState
	}
	ChannelState struct {
		ChannelStateData
		mux sync.RWMutex
	}
	UserState struct {
		UserStateData
		mux sync.RWMutex
	}

	ChannelStateData struct {
		EmoteOnly      bool
		FollowerOnly   time.Duration
		R9k            bool
		Slow           time.Duration
		SubscriberOnly bool
	}
	UserStateData struct {
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

// Login sends the required messages to IRC to login
func (b *Bot) Login(name, pass string) {
	b.Send(twitch.MakePassCommand(pass), twitch.MakeNickCommand(name))
}

// Send sends messages to IRC
func (b *Bot) Send(messages ...client.IrcMessage) {
	for _, message := range messages {
		b.client.Output() <- message
	}
}

func (s *ChannelState) Update(emoteOnly, r9k, subscriber bool, follower, slow time.Duration) {
	if s == nil {
		return
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.EmoteOnly = emoteOnly
	s.R9k = r9k
	s.SubscriberOnly = subscriber
	s.FollowerOnly = follower
	s.Slow = slow
}

func (s *ChannelState) Data() ChannelStateData {
	if s == nil {
		return ChannelStateData{}
	}
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.ChannelStateData
}

func (s *UserState) Data() UserStateData {
	if s == nil {
		return UserStateData{}
	}
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.UserStateData
}

func (b *Bot) GetChannelData(channel string) (*ChannelStateData, error) {
	s := b.State
	s.mux.RLock()
	defer s.mux.RUnlock()
	ch, ok := s.Channels[channel]
	if !ok {
		return nil, ErrNotInChannel
	}
	snap := ch.State.Data()
	return &snap, nil
}

func (s *UserState) Update(mod, subscriber bool) {
	if s == nil {
		return
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Mod = mod
	s.Subscriber = subscriber
}

// AddChannel adds the channel to State
func (b *Bot) AddChannel(channel string) {
	s := b.State
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Channels[channel] = &Channel{
		State: ChannelState{},
		User:  UserState{},
	}
}

// AddCapability adds the Capability to state
func (b *Bot) AddCapability(capability twitch.Capability) {
	s := b.State
	if !b.HasCapability(capability) {
		s.mux.Lock()
		defer s.mux.Unlock()
		s.Capabilities = append(s.Capabilities, capability)
	}
}

// HasCapability returns whether the capability exists within State
func (b *Bot) HasCapability(capability twitch.Capability) bool {
	b.State.mux.RLock()
	defer b.State.mux.RUnlock()
	for _, c := range b.State.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// UpdateUserState updates UserState for the given channel
func (b *Bot) UpdateUserState(channel string, mod, subscriber bool) error {
	s := b.State
	s.mux.RLock()
	defer s.mux.RUnlock()
	c, ok := s.Channels[channel]
	if !ok {
		return ErrNotInChannel
	}
	c.User.Update(mod, subscriber)
	return nil
}

// UpdateChannel updates ChannelState for the provided channel
func (b *Bot) UpdateChannel(channel string, emoteOnly, r9k, subscriber bool, follower, slow time.Duration) error {
	s := b.State
	s.mux.RLock()
	defer s.mux.RUnlock()
	c, ok := s.Channels[channel]
	if !ok {
		return ErrNotInChannel
	}
	c.State.Update(emoteOnly, r9k, subscriber, follower, slow)
	return nil
}

// RemoveChannel removes the channel from State
func (b *Bot) RemoveChannel(channel string) {
	s := b.State
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.Channels, channel)
}

// Close closes the connection to IRC
func (b *Bot) Close() error {
	return b.client.Close()
}

// JoinChannel sends a Join message to IRC
func (b *Bot) JoinChannel(channel string) {
	b.Send(twitch.MakeJoinCommand(channel))
}

// RequestCapability sends Capability requests to IRC
func (b *Bot) RequestCapability(capabilities ...twitch.Capability) {
	for _, capability := range capabilities {
		b.Send(twitch.MakeCapabilityRequest(capability))
	}
}

// LeaveChannel sends a Part message to IRC & removes the channel from State
func (b *Bot) LeaveChannel(channel string) {
	b.Send(twitch.MakePartCommand(channel))
	b.RemoveChannel(channel)
}
