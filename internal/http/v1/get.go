package v1

import (
	gen "app/internal/api/v1"
	"app/internal/converter"
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var httpTracer = otel.Tracer("app/http/v1")

func (h *Handler) GetOrder(ctx context.Context, params gen.GetOrderParams) (gen.GetOrderRes, error) {
	ctx, span := httpTracer.Start(ctx, "v1.GetOrder",
		trace.WithAttributes(attribute.String("order.uid", params.OrderUID)),
	)
	defer span.End()

	order, err := h.orderService.Get(ctx, params.OrderUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service error")
		return nil, err
	}

	res := converter.ModelOrderToGen(order)
	span.SetStatus(codes.Ok, "ok")
	return &res, nil
}
