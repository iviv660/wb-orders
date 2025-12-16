package order

import (
	service "app/internal/model"
	"app/internal/repository/converter"
	repo "app/internal/repository/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (o *OrderRepository) GetOrder(ctx context.Context, uuid string) (service.Order, error) {
	oRow, err := o.getOrderRow(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}

	dRow, err := o.getDeliveryRow(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}

	pRow, err := o.getPaymentRow(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}

	itRows, err := o.getItemsRow(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}

	order := converter.ConvertRepoOrderToServiceOrder(oRow)
	order.Delivery = converter.ConvertRepoDeliveryToServiceDelivery(dRow)
	order.Payment = converter.ConvertRepoPaymentToServicePayment(pRow)
	order.Items = converter.RowsToItems(itRows)

	return order, nil
}

func (o *OrderRepository) getOrderRow(ctx context.Context, uuid string) (repo.OrderRow, error) {
	rows, err := o.pool.Query(ctx, `
SELECT order_uid, track_number, entry, locale, internal_signature,
       customer_id, delivery_service, shardkey,
       sm_id, date_created, oof_shard
FROM orders
WHERE order_uid = $1
`, uuid)
	if err != nil {
		return repo.OrderRow{}, err
	}
	defer rows.Close()

	var oRow repo.OrderRow
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return repo.OrderRow{}, err
		}
		return repo.OrderRow{}, pgx.ErrNoRows
	}

	if err := rows.Scan(
		&oRow.OrderUUID,
		&oRow.TrackNumber,
		&oRow.Entry,
		&oRow.Locale,
		&oRow.InternalSignature,
		&oRow.CustomerID,
		&oRow.DeliveryService,
		&oRow.ShardKey,
		&oRow.SmID,
		&oRow.DateCreated,
		&oRow.OffShard,
	); err != nil {
		return repo.OrderRow{}, err
	}

	if rows.Next() {
		return repo.OrderRow{}, errors.New("orders: expected single row, got multiple")
	}

	if err := rows.Err(); err != nil {
		return repo.OrderRow{}, err
	}

	return oRow, nil
}

func (o *OrderRepository) getDeliveryRow(ctx context.Context, uuid string) (repo.DeliveryRow, error) {
	rows, err := o.pool.Query(ctx, `
SELECT order_uid, name, phone, zip, city,
       address, region, email
FROM deliveries
WHERE order_uid = $1
`, uuid)
	if err != nil {
		return repo.DeliveryRow{}, err
	}
	defer rows.Close()

	var dRow repo.DeliveryRow
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return repo.DeliveryRow{}, err
		}
		return repo.DeliveryRow{}, pgx.ErrNoRows
	}

	if err := rows.Scan(
		&dRow.OrderUID,
		&dRow.Name,
		&dRow.Phone,
		&dRow.Zip,
		&dRow.City,
		&dRow.Address,
		&dRow.Region,
		&dRow.Email,
	); err != nil {
		return repo.DeliveryRow{}, err
	}

	if rows.Next() {
		return repo.DeliveryRow{}, errors.New("deliveries: expected single row, got multiple")
	}

	if err := rows.Err(); err != nil {
		return repo.DeliveryRow{}, err
	}

	return dRow, nil
}

func (o *OrderRepository) getPaymentRow(ctx context.Context, uuid string) (repo.PaymentRow, error) {
	rows, err := o.pool.Query(ctx, `
SELECT order_uid, transaction, request_id, currency, provider,
       amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
FROM payments
WHERE order_uid = $1
`, uuid)
	if err != nil {
		return repo.PaymentRow{}, err
	}
	defer rows.Close()

	var pRow repo.PaymentRow
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return repo.PaymentRow{}, err
		}
		return repo.PaymentRow{}, pgx.ErrNoRows
	}

	if err := rows.Scan(
		&pRow.OrderUID,
		&pRow.Transaction,
		&pRow.RequestID,
		&pRow.Currency,
		&pRow.Provider,
		&pRow.Amount,
		&pRow.PaymentDT,
		&pRow.Bank,
		&pRow.DeliveryCost,
		&pRow.GoodsTotal,
		&pRow.CustomFee,
	); err != nil {
		return repo.PaymentRow{}, err
	}

	if rows.Next() {
		return repo.PaymentRow{}, errors.New("payments: expected single row, got multiple")
	}

	if err := rows.Err(); err != nil {
		return repo.PaymentRow{}, err
	}

	return pRow, nil
}

func (o *OrderRepository) getItemsRow(ctx context.Context, uuid string) ([]repo.ItemRow, error) {
	rows, err := o.pool.Query(ctx, `
SELECT order_uid, chrt_id, track_number,
       price, rid, name, sale, size, total_price,
       nm_id, brand, status
FROM items
WHERE order_uid = $1
`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repo.ItemRow, 0, 8)
	for rows.Next() {
		var it repo.ItemRow
		if err := rows.Scan(
			&it.OrderUID,
			&it.ChrtID,
			&it.TrackNumber,
			&it.Price,
			&it.Rid,
			&it.Name,
			&it.Sale,
			&it.Size,
			&it.TotalPrice,
			&it.NmId,
			&it.Brand,
			&it.Status,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
