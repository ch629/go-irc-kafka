package kafka

import "google.golang.org/protobuf/proto"

type protoEncoder struct {
	proto.Message
}

func (enc protoEncoder) Encode() ([]byte, error) {
	return proto.Marshal(enc)
}

func (enc protoEncoder) Length() int {
	return proto.Size(enc)
}
