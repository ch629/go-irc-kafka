package bot

import (
	"context"
	"errors"
	"fmt"

	"github.com/ch629/go-irc-kafka/domain"
	"github.com/ch629/go-irc-kafka/irc"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/twitch"
	"go.uber.org/zap"
)

//go:generate mockery --name=IRCReadWriter
type IRCReadWriter interface {
	// TODO: can we decouple this from parser & client?
	Input() <-chan parser.Message
	Send(messages ...client.IrcMessage) error
}

// TODO: Need some sort of channel to mark when we're ready to send messages to IRC
type Bot struct {
	ircReadWriter  IRCReadWriter
	errors         chan error
	messageHandler MessageHandler
	loginError     chan error
}

var ErrBadPassword = errors.New("")

func New(irc IRCReadWriter, messageHandler MessageHandler) *Bot {
	return &Bot{
		// TODO: Potentially this could only take a PONG function which is needed for this level in the bot?
		//  -> If we decide to have auto joining & logging in, we need to expose those
		//  -> Depends on our boundaries of what a "Bot" is, does it handle sending messages back to IRC, or is this
		//     just a reader bot which can pull messages & handles keeping it alive with something else which abstracts
		//     away from sending messages back to the client
		ircReadWriter:  irc,
		errors:         make(chan error),
		messageHandler: messageHandler,
	}
}

// TODO: Pass some config into this? -> Auto join channels on ready,
func (b *Bot) ProcessMessages(ctx context.Context) {
	log := zap.L()
	// TODO: This is assuming we'll only ever call this in 1 goroutine
	defer close(b.errors)
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
					b.errors <- fmt.Errorf("failed to send PONG: %w", err)
				}
			case irc.PrivateMessage:
				msg, err := domain.MakeChatMessage(message)
				if err != nil {
					b.errors <- fmt.Errorf("failed to map chat message %w", err)
					continue
				}

				if b.messageHandler.onPrivateMessage != nil {
					b.messageHandler.onPrivateMessage(*msg)
				}
			case irc.ClearChat:
				ban, err := domain.NewBan(message)
				if err != nil {
					b.errors <- fmt.Errorf("failed to map ban message %w", err)
					continue
				}
				if b.messageHandler.onBan != nil {
					b.messageHandler.onBan(*ban)
				}
			case irc.EndOfMOTD:
				// Connected & ready to join channels
				b.loginError <- nil
			// ERR_PASSWDMISMATCH
			case irc.ErrPasswordMismatch:
				b.loginError <- ErrBadPassword
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

// Login logs into the IRC server using the name and password, blocking until either the login was successful, fails or the context is cancelled
func (b *Bot) Login(ctx context.Context, name, pass string) error {
	b.loginError = make(chan error)
	defer close(b.loginError)
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

func (b *Bot) RequestCapability(capabilities ...twitch.Capability) error {
	for _, capability := range capabilities {
		if err := b.ircReadWriter.Send(twitch.MakeCapabilityRequest(capability)); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) Errors() <-chan error {
	return b.errors
}
