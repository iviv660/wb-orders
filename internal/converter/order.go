package converter

import (
	gen "app/internal/api/v1"
	"app/internal/model"
)

//
// model -> gen
//

func ModelOrderToGen(o model.Order) gen.Order {
	items := make([]gen.Item, len(o.Items))
	for i, it := range o.Items {
		items[i] = ModelItemToGen(it)
	}

	return gen.Order{
		OrderUID:          o.OrderUUID,
		TrackNumber:       o.TrackNumber,
		Entry:             o.Entry,
		Locale:            o.Locale,
		InternalSignature: o.InternalSignature,
		CustomerID:        o.CustomerID,
		DeliveryService:   o.DeliveryService,
		ShardKey:          o.ShardKEy,
		SmID:              int32(o.SmID),
		DateCreated:       o.DateCreated,
		OffShard:          o.OffShard,
		Delivery:          ModelDeliveryToGen(o.Delivery),
		Payment:           ModelPaymentToGen(o.Payment),
		Items:             items,
	}
}

func ModelDeliveryToGen(d model.Delivery) gen.Delivery {
	return gen.Delivery{
		Name:    d.Name,
		Phone:   d.Phone,
		Zip:     d.Zip,
		City:    d.City,
		Address: d.Address,
		Region:  d.Region,
		Email:   d.Email,
	}
}

func ModelPaymentToGen(p model.Payment) gen.Payment {
	return gen.Payment{
		Transaction:  p.Transaction,
		Request:      p.RequestID,
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       int32(p.Amount),
		PaymentDt:    int32(p.PaymentDT),
		Bank:         p.Bank,
		DeliveryCost: int32(p.DeliveryCost),
		GoodsTotal:   int32(p.GoodsTotal),
		CustomFee:    int32(p.CustomFee),
	}
}

func ModelItemToGen(it model.Item) gen.Item {
	return gen.Item{
		ChrtID:      it.ChrtID,
		TrackNumber: it.TrackNumber,
		Price:       int32(it.Price),
		Rid:         it.Rid,
		Name:        it.Name,
		Sale:        int32(it.Sale),
		Size:        it.Size,
		TotalPrice:  int32(it.TotalPrice),
		NmID:        int32(it.NmID),
		Brand:       it.Brand,
		Status:      int32(it.Status),
	}
}

//
// gen -> model
//

func GenOrderToModel(o gen.Order) model.Order {
	items := make([]model.Item, len(o.Items))
	for i, it := range o.Items {
		items[i] = GenItemToModel(it)
	}

	return model.Order{
		OrderUUID:         o.OrderUID,
		TrackNumber:       o.TrackNumber,
		Entry:             o.Entry,
		Locale:            o.Locale,
		InternalSignature: o.InternalSignature,
		CustomerID:        o.CustomerID,
		DeliveryService:   o.DeliveryService,
		ShardKEy:          o.ShardKey,
		SmID:              int(o.SmID),
		DateCreated:       o.DateCreated,
		OffShard:          o.OffShard,
		Delivery:          GenDeliveryToModel(o.Delivery),
		Payment:           GenPaymentToModel(o.Payment),
		Items:             items,
	}
}

func GenDeliveryToModel(d gen.Delivery) model.Delivery {
	return model.Delivery{
		Name:    d.Name,
		Phone:   d.Phone,
		Zip:     d.Zip,
		City:    d.City,
		Address: d.Address,
		Region:  d.Region,
		Email:   d.Email,
	}
}

func GenPaymentToModel(p gen.Payment) model.Payment {
	return model.Payment{
		Transaction:  p.Transaction,
		RequestID:    p.Request,
		Currency:     p.Currency,
		Provider:     p.Provider,
		Amount:       int(p.Amount),
		PaymentDT:    int(p.PaymentDt),
		Bank:         p.Bank,
		DeliveryCost: int(p.DeliveryCost),
		GoodsTotal:   int(p.GoodsTotal),
		CustomFee:    int(p.CustomFee),
	}
}

func GenItemToModel(it gen.Item) model.Item {
	return model.Item{
		ChrtID:      it.ChrtID,
		TrackNumber: it.TrackNumber,
		Price:       int(it.Price),
		Rid:         it.Rid,
		Name:        it.Name,
		Sale:        int(it.Sale),
		Size:        it.Size,
		TotalPrice:  int(it.TotalPrice),
		NmID:        int(it.NmID),
		Brand:       it.Brand,
		Status:      int(it.Status),
	}
}
