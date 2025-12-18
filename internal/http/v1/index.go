package v1

import (
	gen "app/internal/api/v1"
	"bytes"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func (h *Handler) Index(ctx context.Context) (gen.IndexOK, error) {
	_, span := otel.Tracer("app/http/v1").Start(ctx, "v1.Index")
	defer span.End()

	span.SetStatus(codes.Ok, "ok")
	return gen.IndexOK{Data: bytes.NewBufferString(indexHTML)}, nil
}
