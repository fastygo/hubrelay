package grpcapi

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

const codecName = "json"

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

type jsonCodec struct{}

func (jsonCodec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (jsonCodec) Unmarshal(data []byte, value any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, value)
}

func (jsonCodec) Name() string {
	return codecName
}
