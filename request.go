package rest

import (
	"context"
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

func Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func NewRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, reqID{}, uuid.New().String())
}

type reqID struct{}

func RequestID(ctx context.Context) string {
	id, ok := ctx.Value(reqID{}).(string)
	if !ok {
		return "<missing-request-id>"
	}
	return id
}
