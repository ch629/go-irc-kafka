package kafka

import "google.golang.org/protobuf/proto"

type ProtoEncoder struct {
	proto.Message
}

func (enc ProtoEncoder) Encode() ([]byte, error) {
	return proto.Marshal(enc)
}

func (enc ProtoEncoder) Length() int {
	return proto.Size(enc)
}
