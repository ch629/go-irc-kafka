package main

import (
	"context"
	"fmt"
	"github.com/ch629/go-irc-kafka/bot"
	"github.com/ch629/go-irc-kafka/config"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/kafka"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/middleware"
	"github.com/ch629/go-irc-kafka/shutdown"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
	"net"
	"net/http"
)

// https://tools.ietf.org/html/rfc1459.html

func main() {
	log := logging.Logger()
	ctx := shutdown.InterruptAwareContext(context.Background())
	graceful := &shutdown.GracefulShutdown{}
	b, err := startBot(ctx, graceful)
	if err != nil {
		log.Fatal("Failed to start bot", zap.Error(err))
	}
	defer b.Close()

	server := startHttpServer(b, log)
	defer server.Close()

	graceful.Wait()
}

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
		})
	// TODO: Zap recovery
	neg := negroni.New(middleware.NewLogger(log), negroni.NewRecovery())
	neg.UseHandler(r)

	server := &http.Server{
		Addr:    ":8080",
		Handler: neg,
	}
	go server.ListenAndServe()
	return server
}

func startBot(ctx context.Context, graceful *shutdown.GracefulShutdown) (*bot.Bot, error) {
	fs := afero.NewOsFs()
	log := logging.Logger()

	conf, err := config.LoadConfig(fs)
	if err != nil {
		return nil, err
	}

	conn, err := makeConnection(conf.Irc.Address)
	if err != nil {
		return nil, err
	}

	ircClient := client.NewClient(ctx, conn)
	graceful.RegisterWait(ircClient)

	producer, err := kafka.NewProducer(conf.Kafka)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer %w", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Error("error from producer", zap.Error(err))
		}
	}()

	handler := botMessageHandler{
		conf:     conf,
		log:      logging.Logger(),
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
