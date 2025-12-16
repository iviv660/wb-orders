package kafka

import (
	"context"
	"errors"
	"sync"
	"testing"

	"app/internal/otelx"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type fakeReader struct {
	mu sync.Mutex

	msgs      []kafka.Message
	fetchErrs []error

	commitErr error

	fetchCalls  int
	commitCalls int
	committed   []kafka.Message
}

func (r *fakeReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	i := r.fetchCalls
	r.fetchCalls++

	if i < len(r.fetchErrs) && r.fetchErrs[i] != nil {
		return kafka.Message{}, r.fetchErrs[i]
	}
	if i < len(r.msgs) {
		return r.msgs[i], nil
	}
	return kafka.Message{}, context.Canceled
}

func (r *fakeReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commitCalls++
	r.committed = append(r.committed, msgs...)
	return r.commitErr
}

func TestConsumer_Read_OK_CommitsAndPassesSpanCtx(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	parentCtx, parentSpan := otel.Tracer("test").Start(context.Background(), "parent")
	parentSpan.End()

	msg := kafka.Message{
		Topic:     "orders",
		Partition: 1,
		Offset:    10,
		Headers:   nil,
	}

	carrier := otelx.KafkaHeaderCarrier{Headers: &msg.Headers}
	otel.GetTextMapPropagator().Inject(parentCtx, carrier)

	fr := &fakeReader{
		msgs:      []kafka.Message{msg},
		fetchErrs: []error{nil, context.Canceled},
	}

	c := New(fr)

	var gotTraceID trace.TraceID

	err := c.Read(context.Background(), func(ctx context.Context, m kafka.Message) error {
		sc := trace.SpanFromContext(ctx).SpanContext()
		require.True(t, sc.IsValid())
		gotTraceID = sc.TraceID()
		return nil
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, fr.commitCalls)
	require.Len(t, fr.committed, 1)
	require.Equal(t, msg.Topic, fr.committed[0].Topic)
	require.NotEqual(t, trace.TraceID{}, gotTraceID)
}

func TestConsumer_Read_HandlerError_NoCommit(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	msg := kafka.Message{Topic: "orders", Partition: 1, Offset: 10}
	fr := &fakeReader{msgs: []kafka.Message{msg}}
	c := New(fr)

	wantErr := errors.New("handler failed")

	err := c.Read(context.Background(), func(ctx context.Context, m kafka.Message) error {
		return wantErr
	})

	require.ErrorIs(t, err, wantErr)
	require.Equal(t, 0, fr.commitCalls)
}

func TestConsumer_Read_FetchError(t *testing.T) {
	fr := &fakeReader{fetchErrs: []error{errors.New("fetch failed")}}
	c := New(fr)

	err := c.Read(context.Background(), func(ctx context.Context, m kafka.Message) error {
		return nil
	})

	require.Error(t, err)
	require.Equal(t, 0, fr.commitCalls)
}

func TestConsumer_Read_CommitError(t *testing.T) {
	msg := kafka.Message{Topic: "orders", Partition: 1, Offset: 10}
	fr := &fakeReader{
		msgs:      []kafka.Message{msg},
		commitErr: errors.New("commit failed"),
	}
	c := New(fr)

	err := c.Read(context.Background(), func(ctx context.Context, m kafka.Message) error {
		return nil
	})

	require.Error(t, err)
	require.Equal(t, 1, fr.commitCalls)
}
