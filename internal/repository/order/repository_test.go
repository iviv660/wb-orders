package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"app/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_SetOrder_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	r := &OrderRepository{pool: mock}

	order := model.Order{
		OrderUUID:         "uid-1",
		TrackNumber:       "track-1",
		Entry:             "entry",
		Locale:            "ru",
		InternalSignature: "sig",
		CustomerID:        "cust-1",
		DeliveryService:   "dhl",
		ShardKEy:          "shard",
		SmID:              1,
		DateCreated:       time.Now().UTC(),
		OffShard:          "off",
		Delivery:          model.Delivery{Name: "n", Phone: "p", Zip: "z", City: "c", Address: "a", Region: "r", Email: "e"},
		Payment:           model.Payment{Transaction: "t", RequestID: "r", Currency: "RUB", Provider: "p", Amount: 10, PaymentDT: 1, Bank: "b", DeliveryCost: 1, GoodsTotal: 2, CustomFee: 3},
		Items:             []model.Item{{ChrtID: 1, TrackNumber: "track-1", Price: 100, Rid: "rid", Name: "name", Sale: 0, Size: "0", TotalPrice: 100, NmID: 10, Brand: "br", Status: 1}},
	}

	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(
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
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectExec("INSERT INTO deliveries").
		WithArgs(
			order.OrderUUID,
			order.Delivery.Name,
			order.Delivery.Phone,
			order.Delivery.Zip,
			order.Delivery.City,
			order.Delivery.Address,
			order.Delivery.Region,
			order.Delivery.Email,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectExec("INSERT INTO payments").
		WithArgs(
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
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectExec("INSERT INTO items").
		WithArgs(
			order.OrderUUID,
			order.Items[0].ChrtID,
			order.Items[0].TrackNumber,
			order.Items[0].Price,
			order.Items[0].Rid,
			order.Items[0].Name,
			order.Items[0].Sale,
			order.Items[0].Size,
			order.Items[0].TotalPrice,
			order.Items[0].NmID,
			order.Items[0].Brand,
			order.Items[0].Status,
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	err = r.SetOrder(ctx, order)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_SetOrder_OrderInsertError_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	r := &OrderRepository{pool: mock}

	order := model.Order{
		OrderUUID:         "uid-1",
		TrackNumber:       "track-1",
		Entry:             "entry",
		Locale:            "ru",
		InternalSignature: "sig",
		CustomerID:        "cust-1",
		DeliveryService:   "dhl",
		ShardKEy:          "shard",
		SmID:              1,
		DateCreated:       time.Now().UTC(),
		OffShard:          "off",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(
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
		).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err = r.SetOrder(ctx, order)
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrder_NoRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	r := &OrderRepository{pool: mock}

	mock.ExpectQuery("FROM orders").
		WithArgs("uid-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey",
			"sm_id", "date_created", "oof_shard",
		}))

	_, err = r.GetOrder(ctx, "uid-1")
	require.ErrorIs(t, err, pgx.ErrNoRows)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrder_OK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	r := &OrderRepository{pool: mock}

	now := time.Now().UTC()

	mock.ExpectQuery("FROM orders").
		WithArgs("uid-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"order_uid", "track_number", "entry", "locale", "internal_signature",
			"customer_id", "delivery_service", "shardkey",
			"sm_id", "date_created", "oof_shard",
		}).AddRow(
			"uid-1", "track-1", "entry", "ru", "sig",
			"cust-1", "dhl", "shard",
			int32(1), now, "off",
		))

	mock.ExpectQuery("FROM deliveries").
		WithArgs("uid-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"order_uid", "name", "phone", "zip", "city", "address", "region", "email",
		}).AddRow(
			"uid-1", "n", "p", "z", "c", "a", "r", "e",
		))

	mock.ExpectQuery("FROM payments").
		WithArgs("uid-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"order_uid", "transaction", "request_id", "currency", "provider",
			"amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
		}).AddRow(
			"uid-1", "t", "r", "RUB", "prov",
			int32(10), int64(1), "b", int32(1), int32(2), int32(3),
		))

	mock.ExpectQuery("FROM items").
		WithArgs("uid-1").
		WillReturnRows(pgxmock.NewRows([]string{
			"order_uid", "chrt_id", "track_number", "price", "rid", "name", "sale", "size",
			"total_price", "nm_id", "brand", "status",
		}).AddRow(
			"uid-1", int64(1), "track-1", int32(100), "rid", "name", int32(0), "0",
			int32(100), int64(10), "br", int32(1),
		))

	got, err := r.GetOrder(ctx, "uid-1")
	require.NoError(t, err)
	require.Equal(t, "uid-1", got.OrderUUID)

	require.NoError(t, mock.ExpectationsWereMet())
}
