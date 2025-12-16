package kafka

import (
	"app/internal/otelx"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

var ErrDLQWriterNil = errors.New("dlq writer is nil")

type dlqEnvelope struct {
	Reason   string      `json:"reason"`
	Retries  int         `json:"retries"`
	FailedAt time.Time   `json:"failed_at"`
	Original dlqOriginal `json:"original"`
}

type dlqOriginal struct {
	Topic     string      `json:"topic"`
	Partition int         `json:"partition"`
	Offset    int64       `json:"offset"`
	Time      time.Time   `json:"time"`
	KeyB64    string      `json:"key_b64,omitempty"`
	ValueB64  string      `json:"value_b64,omitempty"`
	Headers   []dlqHeader `json:"headers,omitempty"`
}

type dlqHeader struct {
	Key   string `json:"key"`
	Value string `json:"value_b64"`
}

func sendToDLQ(ctx context.Context, w *kafka.Writer, orig kafka.Message, cause error, retries int) error {
	if w == nil {
		return ErrDLQWriterNil
	}

	env := dlqEnvelope{
		Reason:   errString(cause),
		Retries:  retries,
		FailedAt: time.Now().UTC(),
		Original: dlqOriginal{
			Topic:     orig.Topic,
			Partition: orig.Partition,
			Offset:    orig.Offset,
			Time:      orig.Time,
			KeyB64:    b64(orig.Key),
			ValueB64:  b64(orig.Value),
			Headers:   toDLQHeaders(orig.Headers),
		},
	}

	b, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal dlq envelope: %w", err)
	}

	dlqMsg := kafka.Message{
		Key:   orig.Key,
		Value: b,
		Headers: []kafka.Header{
			{Key: "x-orig-topic", Value: []byte(orig.Topic)},
			{Key: "x-orig-partition", Value: []byte(fmt.Sprintf("%d", orig.Partition))},
			{Key: "x-orig-offset", Value: []byte(fmt.Sprintf("%d", orig.Offset))},
			{Key: "x-retries", Value: []byte(fmt.Sprintf("%d", retries))},
			{Key: "x-error", Value: []byte(errString(cause))},
			{Key: "x-failed-at", Value: []byte(env.FailedAt.Format(time.RFC3339Nano))},
		},
	}

	otelx.InjectKafka(ctx, &dlqMsg)

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := w.WriteMessages(writeCtx, dlqMsg); err != nil {
		return fmt.Errorf("write to dlq: %w", err)
	}
	return nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func b64(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func toDLQHeaders(hdrs []kafka.Header) []dlqHeader {
	if len(hdrs) == 0 {
		return nil
	}
	out := make([]dlqHeader, 0, len(hdrs))
	for _, h := range hdrs {
		out = append(out, dlqHeader{
			Key:   h.Key,
			Value: b64(h.Value),
		})
	}
	return out
}
