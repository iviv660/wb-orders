package order

import (
	service "app/internal/model"
	"context"
)

const insertOrderQuery = `
INSERT INTO orders (
    order_uid, track_number, entry,
    locale, internal_signature, customer_id,
    delivery_service, shardkey, sm_id,
    date_created, oof_shard
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
`

const insertDeliveryQuery = `
INSERT INTO deliveries (
    order_uid, name, phone, zip, city, address, region, email
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
`

const insertPaymentQuery = `
INSERT INTO payments (
    order_uid, transaction, request_id, currency, provider,
    amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
`

const insertItemQuery = `
INSERT INTO items (
    order_uid, chrt_id, track_number, price,
    rid, name, sale, size, total_price, nm_id, brand, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
`

func (o *OrderRepository) SetOrder(ctx context.Context, order service.Order) error {
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, insertOrderQuery,
		order.OrderUUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKEy,
		order.SmID,
		order.DateCreated,
		order.OffShard,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, insertDeliveryQuery,
		order.OrderUUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, insertPaymentQuery,
		order.OrderUUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	); err != nil {
		return err
	}

	for _, it := range order.Items {
		if _, err := tx.Exec(ctx, insertItemQuery,
			order.OrderUUID,
			it.ChrtID,
			it.TrackNumber,
			it.Price,
			it.Rid,
			it.Name,
			it.Sale,
			it.Size,
			it.TotalPrice,
			it.NmID,
			it.Brand,
			it.Status,
		); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
