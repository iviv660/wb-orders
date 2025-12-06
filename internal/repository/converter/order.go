package converter

import (
	service "app/internal/model"
	repo "app/internal/repository/model"
)

//Конвертации из service model в repository model

func ConvertServiceOrderToRepoOrder(o service.Order) repo.OrderRow {
	return repo.OrderRow{
		OrderUUID:         o.OrderUUID,
		TrackNumber:       o.TrackNumber,
		Entry:             o.Entry,
		Locale:            o.Locale,
		InternalSignature: o.InternalSignature,
		CustomerID:        o.CustomerID,
		DeliveryService:   o.DeliveryService,
		ShardKey:          o.ShardKEy,
		SmID:              o.SmID,
		DateCreated:       o.DateCreated,
		OffShard:          o.OffShard,
	}
}

func ConvertServiceDeliveryToRepoDelivery(orderUUID string, d service.Delivery) repo.DeliveryRow {
	return repo.DeliveryRow{
		OrderUID: orderUUID,
		Name:     d.Name,
		Phone:    d.Phone,
		Zip:      d.Zip,
		City:     d.City,
		Address:  d.Address,
		Region:   d.Region,
		Email:    d.Email,
	}
}

func ConvertServicePaymentToRepoPayment(orderUUID string, p service.Payment) repo.PaymentRow {
	return repo.PaymentRow{
		OrderUID:     orderUUID,
		Transaction:  p.Transaction,
		RequestID:    p.RequestID,
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       p.Amount,
		PaymentDT:    p.PaymentDT,
		Bank:         p.Bank,
		DeliveryCost: p.DeliveryCost,
		GoodsTotal:   p.GoodsTotal,
		CustomFee:    p.CustomFee,
	}
}

func ConvertServiceItemToRepoItem(orderUID string, it service.Item) repo.ItemRow {
	return repo.ItemRow{
		OrderUID:    orderUID,
		ChrtID:      it.ChrtID,
		TrackNumber: it.TrackNumber,
		Price:       it.Price,
		Rid:         it.Rid,
		Name:        it.Name,
		Sale:        it.Sale,
		Size:        it.Size,
		TotalPrice:  it.TotalPrice,
		NmId:        it.NmID,
		Brand:       it.Brand,
		Status:      it.Status,
	}
}

func ItemsToRows(orderUID string, items []service.Item) []repo.ItemRow {
	rows := make([]repo.ItemRow, len(items))
	for i, it := range items {
		rows[i] = ConvertServiceItemToRepoItem(orderUID, it)
	}
	return rows
}

//Конвертации из repository model в service model

func ConvertRepoOrderToServiceOrder(o repo.OrderRow) service.Order {
	return service.Order{
		OrderUUID:         o.OrderUUID,
		TrackNumber:       o.TrackNumber,
		Entry:             o.Entry,
		Locale:            o.Locale,
		InternalSignature: o.InternalSignature,
		CustomerID:        o.CustomerID,
		DeliveryService:   o.DeliveryService,
		ShardKEy:          o.ShardKey,
		SmID:              o.SmID,
		DateCreated:       o.DateCreated,
		OffShard:          o.OffShard,
	}
}

func ConvertRepoDeliveryToServiceDelivery(d repo.DeliveryRow) service.Delivery {
	return service.Delivery{
		Name:    d.Name,
		Phone:   d.Phone,
		Zip:     d.Zip,
		City:    d.City,
		Address: d.Address,
		Region:  d.Region,
		Email:   d.Email,
	}
}

func ConvertRepoPaymentToServicePayment(p repo.PaymentRow) service.Payment {
	return service.Payment{
		Transaction:  p.Transaction,
		RequestID:    p.RequestID,
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       p.Amount,
		PaymentDT:    p.PaymentDT,
		Bank:         p.Bank,
		DeliveryCost: p.DeliveryCost,
		GoodsTotal:   p.GoodsTotal,
		CustomFee:    p.CustomFee,
	}
}

func ConvertRepoItemToServiceItem(it repo.ItemRow) service.Item {
	return service.Item{
		ChrtID:      it.ChrtID,
		TrackNumber: it.TrackNumber,
		Price:       it.Price,
		Rid:         it.Rid,
		Name:        it.Name,
		Sale:        it.Sale,
		Size:        it.Size,
		TotalPrice:  it.TotalPrice,
		NmID:        it.NmId,
		Brand:       it.Brand,
		Status:      it.Status,
	}
}

func RowsToItems(rows []repo.ItemRow) []service.Item {
	items := make([]service.Item, len(rows))
	for i, r := range rows {
		items[i] = ConvertRepoItemToServiceItem(r)
	}
	return items
}
