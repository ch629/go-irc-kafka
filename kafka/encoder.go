package kafka

import (
	"encoding/json"

	"github.com/Shopify/sarama"
)

func NewJsonEncoder(value interface{}) (sarama.ByteEncoder, error) {
	bs, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return sarama.ByteEncoder(bs), nil
}
