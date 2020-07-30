package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"time"
)

var (
	batchSize    = 50
	batchTimeout = time.Second * 10
	writer       *kafka.Writer
)

func push(messages []kafka.Message) {
	fmt.Printf("Pushing batch with %v messages\n", len(messages))

	err := writer.WriteMessages(context.Background(), messages...)
	if err != nil {
		panic(err)
	}
}

func initializeWriter() {
	if writer == nil {
		writer = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  viper.GetStringSlice("kafka.brokers"),
			Topic:    viper.GetString("kafka.topic"),
			Balancer: &kafka.LeastBytes{},
		})
	}
}

func BatchPoll() {
	initializeWriter()

	go func() {
		batch := make([]kafka.Message, 0, batchSize)
		expire := time.After(batchTimeout)
		for {
			select {
			case message, ok := <-kafkaOutput:
				if !ok {
					// TODO: Check this, think it'll only happen if there is a problem with the channel somehow, so I guess we could just skip that message?
					fmt.Println("!ok??")
					continue
				}

				// TODO: Should we be handling the mapping here, or keep as []byte and map when it's pushed
				// TODO: Configurable keys???
				batch = append(batch, kafka.Message{Key: []byte("KEY"), Value: message})
				if len(batch) == batchSize {
					go push(batch)
					batch = batch[:0]
					expire = time.After(batchTimeout)
				}
			case <-expire:
				// TODO: Do we push from here, or do we add it to another channel which takes batches & another goroutine worker processes these as they come in?
				//  Then we can reformat some of this to work in more of a "worker" form
				go push(batch)
				batch = batch[:0]
				expire = time.After(batchTimeout)
			}
		}
	}()
}
