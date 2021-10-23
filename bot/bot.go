package bot

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ch629/go-irc-kafka/domain"
	"github.com/ch629/go-irc-kafka/irc"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/twitch"
	"go.uber.org/zap"
)

//go:generate mockery --name=IRCReadWriter --disable-version-string
type IRCReadWriter interface {
	Input() <-chan parser.Message
	Send(messages ...client.IrcMessage) error
}

type Bot struct {
	ircReadWriter  IRCReadWriter
	messageHandler MessageHandler
	loginError     chan error
	logger         *zap.Logger

	loginMux  sync.Mutex
	loggingIn bool
}

var ErrBadPassword = errors.New("bad password")

func New(irc IRCReadWriter, messageHandler MessageHandler) *Bot {
	return &Bot{
		ircReadWriter:  irc,
		messageHandler: messageHandler,
		logger:         zap.L(),
	}
}

func (b *Bot) ProcessMessages(ctx context.Context) {
	log := b.logger
	for {
		select {
		case message, ok := <-b.ircReadWriter.Input():
			// Channel has closed
			if !ok {
				return
			}
			switch message.Command {
			case irc.Ping:
				if err := b.ircReadWriter.Send(twitch.MakePongCommand(message.Params[0])); err != nil {
					b.error(fmt.Errorf("failed to send PONG: %w", err))
				}
			case irc.PrivateMessage:
				if b.messageHandler.onPrivateMessage == nil {
					continue
				}
				msg, err := domain.MakeChatMessage(message)
				if err != nil {
					b.error(fmt.Errorf("failed to map chat message %w", err))
					continue
				}

				b.messageHandler.onPrivateMessage(*msg)
			case irc.ClearChat:
				if b.messageHandler.onBan == nil {
					continue
				}
				ban, err := domain.NewBan(message)
				if err != nil {
					b.error(fmt.Errorf("failed to map ban message %w", err))
					continue
				}
				b.messageHandler.onBan(*ban)
			case irc.EndOfMOTD:
				// Connected & ready to join channels
				if b.loggingIn {
					b.loginError <- nil
				}
			// ERR_PASSWDMISMATCH
			case irc.ErrPasswordMismatch:
				if b.loggingIn {
					b.loginError <- ErrBadPassword
				}
			// Ignored messages
			case "001", "002", "003", "004", "375", "372", "353", "366":
			default:
				log.Info("received unhandled command", zap.String("command", message.Command), zap.String("message", fmt.Sprintf("%+v", message)))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bot) error(err error) {
	if b.messageHandler.onError != nil && err != nil {
		b.messageHandler.onError(err)
	}
}

// Login logs into the IRC server using the name and password, blocking until either the login was successful, fails or the context is cancelled
func (b *Bot) Login(ctx context.Context, name, pass string) error {
	// TODO: Write some tests around getting login errors after we're done logging in etc
	b.loginMux.Lock()
	b.loggingIn = true
	b.loginError = make(chan error)
	defer func() {
		b.loggingIn = false
		close(b.loginError)
		b.loginMux.Unlock()
	}()
	if err := b.ircReadWriter.Send(twitch.MakePassCommand(pass), twitch.MakeNickCommand(name)); err != nil {
		return err
	}
	select {
	case err := <-b.loginError:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (b *Bot) JoinChannels(channels ...string) error {
	for _, ch := range channels {
		if err := b.ircReadWriter.Send(twitch.MakeJoinCommand(ch)); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) LeaveChannels(channels ...string) error {
	for _, ch := range channels {
		if err := b.ircReadWriter.Send(twitch.MakePartCommand(ch)); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) RequestCapability(capabilities ...twitch.Capability) error {
	for _, capability := range capabilities {
		if err := b.ircReadWriter.Send(twitch.MakeCapabilityRequest(capability)); err != nil {
			return err
		}
	}
	return nil
}
