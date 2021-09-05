package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/middleware"
	"github.com/ch629/go-irc-kafka/twitch"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

// https://tools.ietf.org/html/rfc1459.html

func main() {
	log := zap.L()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	b, err := startBot(ctx)
	if err != nil {
		log.Fatal("Failed to start bot", zap.Error(err))
	}
	defer b.Close()

	server := startHttpServer(b, log)
	defer server.Close()
	<-ctx.Done()
}

// TODO: Move this somewhere else
func startHttpServer(b *bot.Bot, log *zap.Logger) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.
		GET("/channel/:channel", func(c *gin.Context) {
			chanData, err := b.GetChannelData(c.Param("channel"))
			if err != nil {
				c.AbortWithError(http.StatusNotFound, err)
			}
			c.JSON(http.StatusOK, chanData)
		}).
		GET("/channel", func(c *gin.Context) {
			c.JSON(http.StatusOK, b.Channels())
		}).
		GET("/capability", func(c *gin.Context) {
			c.JSON(http.StatusOK, b.Capabilities())
		}).
		DELETE("/channel/:channel", func(c *gin.Context) {
			channel := c.Param("channel")
			if b.InChannel(channel) {
				b.RequestLeaveChannel(channel)
				c.Status(http.StatusOK)
				return
			}
			c.Status(http.StatusBadRequest)
		}).
		POST("/channel/:channel", func(c *gin.Context) {
			channel := c.Param("channel")
			if b.InChannel(channel) {
				// TODO: Respond saying already in channel?
				c.Status(http.StatusBadRequest)
				return
			}
			b.RequestJoinChannel(channel)
			c.Status(http.StatusOK)
		}).
		GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, struct {
				Channels     []string            `json:"channels"`
				Capabilities []twitch.Capability `json:"capabilities"`
			}{
				Channels:     b.Channels(),
				Capabilities: b.Capabilities(),
			})
		})
	// TODO: Zap recovery
	neg := negroni.New(middleware.NewLogger(log), negroni.NewRecovery())
	neg.UseHandler(r)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      neg,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go server.ListenAndServe()
	return server
}

func makeIrcClient(ctx context.Context, address string) (ircClient client.IrcClient, err error) {
	log := zap.L()
	// Sometimes the client closes instantly, retry it 3 times
	// TODO: Does this need to happen after attempting to login, or can we just base it from here?
	for i := 0; i < 3; i++ {
		conn, err := makeConnection(address)
		if err != nil {
			return nil, err
		}
		ircClient = client.NewClient(ctx, conn)
		go ircClient.ConsumeMessages()
		select {
		case <-ircClient.Done():
			err = ircClient.Err()
			log.Warn("IrcClient exited on startup", zap.Error(err))
			// Make sure the connection is closed if we're retrying
			conn.Close()
			continue
		// TODO: Can we just default?
		case <-time.Tick(10 * time.Millisecond):
		}
		break
	}

	if err != nil {
		err = fmt.Errorf("failed to create IrcClient after retries: %w", err)
	}
	return
}

func startBot(ctx context.Context) (*bot.Bot, error) {
	fs := afero.NewOsFs()
	log := zap.L()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		return nil, err
	}

	ircClient, err := makeIrcClient(ctx, conf.Irc.Address)
	if err != nil {
		return nil, err
	}

	producer, err := kafka.NewProducer(conf.Kafka)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer %w", err)
	}

	handler := botMessageHandler{
		conf:     conf,
		log:      log,
		producer: producer,
	}
	b := bot.NewBot(ctx, ircClient, handler.handleMessage)
	b.Login(conf.Bot.Name, conf.Bot.OAuth)
	return b, nil
}

func makeConnection(address string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP Addr %w", err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to Dial TCP %w", err)
	}
	return conn, nil
}
