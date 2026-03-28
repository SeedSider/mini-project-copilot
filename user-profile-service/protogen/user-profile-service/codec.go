package user_profile_service

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// JSONCodec is a custom gRPC codec that uses JSON encoding.
// This is required because our protobuf types are hand-written
// and do not implement proto.Message from google.golang.org/protobuf.
// Clients must use grpc.ForceCodec(JSONCodec{}) when dialing.
type JSONCodec struct{}

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (JSONCodec) Name() string {
	return "json"
}

func init() {
	encoding.RegisterCodec(JSONCodec{})
}
