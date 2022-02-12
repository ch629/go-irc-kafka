package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
)

func NewJSONEncoder(value interface{}) (sarama.ByteEncoder, error) {
	bs, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}
	return sarama.ByteEncoder(bs), nil
}
