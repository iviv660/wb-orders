package otelx

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type KafkaHeaderCarrier struct {
	Headers *[]kafka.Header
}

func (c KafkaHeaderCarrier) Get(key string) string {
	for _, h := range *c.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c KafkaHeaderCarrier) Set(key string, value string) {
	hs := *c.Headers
	for i := range hs {
		if hs[i].Key == key {
			hs[i].Value = []byte(value)
			*c.Headers = hs
			return
		}
	}
	*c.Headers = append(hs, kafka.Header{Key: key, Value: []byte(value)})
}

func (c KafkaHeaderCarrier) Keys() []string {
	hs := *c.Headers
	out := make([]string, 0, len(hs))
	for _, h := range hs {
		out = append(out, h.Key)
	}
	return out
}

func ExtractKafka(ctx context.Context, msg *kafka.Message) context.Context {
	h := msg.Headers
	ctx = otel.GetTextMapPropagator().Extract(ctx, KafkaHeaderCarrier{Headers: &h})
	msg.Headers = h
	return ctx
}

func InjectKafka(ctx context.Context, msg *kafka.Message) {
	h := msg.Headers
	otel.GetTextMapPropagator().Inject(ctx, KafkaHeaderCarrier{Headers: &h})
	msg.Headers = h
}

var _ propagation.TextMapCarrier = KafkaHeaderCarrier{}
