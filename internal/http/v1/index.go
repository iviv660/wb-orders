package v1

import (
	gen "app/internal/api/v1"
	"bytes"
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func (h *Handler) Index(ctx context.Context) (gen.IndexOK, error) {
	_, span := otel.Tracer("app/http/v1").Start(ctx, "v1.Index")
	defer span.End()

	b, err := os.ReadFile("api/index.html")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "read file error")
		return gen.IndexOK{}, err
	}

	span.SetStatus(codes.Ok, "ok")
	return gen.IndexOK{Data: bytes.NewBuffer(b)}, nil
}
