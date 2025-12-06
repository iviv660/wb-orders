package order

import (
	service "app/internal/model"
	"app/internal/repository/converter"
	repo "app/internal/repository/model"
	"context"
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

	itRow, err := o.getItemsRow(ctx, uuid)
	if err != nil {
		return service.Order{}, err
	}
	order := converter.ConvertRepoOrderToServiceOrder(oRow)
	order.Delivery = converter.ConvertRepoDeliveryToServiceDelivery(dRow)
	order.Payment = converter.ConvertRepoPaymentToServicePayment(pRow)
	order.Items = converter.RowsToItems(itRow)

	return order, nil

}

func (o *OrderRepository) getOrderRow(ctx context.Context, uuid string) (repo.OrderRow, error) {
	row, err := o.pool.Query(ctx,
		`SELECT order_uid, track_number, entry, locale, internal_signature, 
       customer_id, delivery_service, shard_key, 
       sm_id, date_created, off_shard FROM orders WHERE order_uuid = $1`,
		uuid,
	)
	if err != nil {
		return repo.OrderRow{}, err
	}
	defer row.Close()

	var oRow repo.OrderRow

	for row.Next() {
		err = row.Scan(&oRow.OrderUUID, &oRow.TrackNumber, &oRow.Entry,
			&oRow.Locale, &oRow.InternalSignature, &oRow.CustomerID,
			&oRow.DeliveryService, &oRow.ShardKey, &oRow.SmID,
			&oRow.DateCreated, &oRow.OffShard)
		if err != nil {
			return repo.OrderRow{}, err
		}
	}
	return oRow, nil
}

func (o *OrderRepository) getDeliveryRow(ctx context.Context, uuid string) (repo.DeliveryRow, error) {
	row, err := o.pool.Query(ctx,
		`SELECT order_uid, name, phone, zip, city, 
			address, region, email FROM delivery WHERE order_uuid = $1`,
		uuid,
	)
	if err != nil {
		return repo.DeliveryRow{}, err
	}

	defer row.Close()

	var dRow repo.DeliveryRow

	for row.Next() {
		err = row.Scan(&dRow.OrderUID, &dRow.Name, &dRow.Phone,
			&dRow.Zip, &dRow.City, &dRow.Address, &dRow.Region,
			&dRow.Email)
		if err != nil {
			return repo.DeliveryRow{}, err
		}
	}
	return dRow, nil
}

func (o *OrderRepository) getPaymentRow(ctx context.Context, uuid string) (repo.PaymentRow, error) {
	row, err := o.pool.Query(ctx,
		`SELECT order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
			FROM payments WHERE order_uuid = $1`,
		uuid,
	)
	if err != nil {
		return repo.PaymentRow{}, err
	}

	var pRow repo.PaymentRow

	for row.Next() {
		err = row.Scan(&pRow.OrderUID, &pRow.Transaction,
			&pRow.RequestID, &pRow.Currency, &pRow.Provider,
			pRow.Amount, pRow.PaymentDT, &pRow.Bank,
			&pRow.DeliveryCost, &pRow.GoodsTotal, &pRow.CustomFee)
		if err != nil {
			return repo.PaymentRow{}, err
		}
	}
	return pRow, nil
}

func (o *OrderRepository) getItemsRow(ctx context.Context, uuid string) ([]repo.ItemRow, error) {
	row, err := o.pool.Query(ctx,
		`SELECT order_uid, chrt_id, track_number, 
		price, rid, name, sale, size, total_price, 
		nm_id, brand, status FROM items WHERE order_uid = $1`,
		uuid,
	)
	if err != nil {
		return []repo.ItemRow{}, err
	}
	defer row.Close()
	items := make([]repo.ItemRow, 0)
	for row.Next() {
		var it repo.ItemRow
		if err = row.Scan(
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
			return []repo.ItemRow{}, err
		}
		items = append(items, it)
	}

	if err = row.Err(); err != nil {
		return []repo.ItemRow{}, err
	}

	return items, nil
}
