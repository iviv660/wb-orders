package converter

import (
	kafka "app/internal/adapter/model"
	"app/internal/model"
)

// DTO -> model

func OrderDTOToModel(dto kafka.OrderDTO) model.Order {
	items := make([]model.Item, len(dto.Items))
	for i, it := range dto.Items {
		items[i] = ItemDTOToModel(it)
	}

	return model.Order{
		OrderUUID:         dto.OrderUID,
		TrackNumber:       dto.TrackNumber,
		Entry:             dto.Entry,
		Locale:            dto.Locale,
		InternalSignature: dto.InternalSignature,
		CustomerID:        dto.CustomerID,
		DeliveryService:   dto.DeliveryService,
		ShardKEy:          dto.ShardKey,
		SmID:              dto.SmID,
		DateCreated:       dto.DateCreated,
		OffShard:          dto.OffShard,
		Delivery:          DeliveryDTOToModel(dto.Delivery),
		Payment:           PaymentDTOToModel(dto.Payment),
		Items:             items,
	}
}

func DeliveryDTOToModel(dto kafka.DeliveryDTO) model.Delivery {
	return model.Delivery{
		Name:    dto.Name,
		Phone:   dto.Phone,
		Zip:     dto.Zip,
		City:    dto.City,
		Address: dto.Address,
		Region:  dto.Region,
		Email:   dto.Email,
	}
}

func PaymentDTOToModel(dto kafka.PaymentDTO) model.Payment {
	return model.Payment{
		Transaction:  dto.Transaction,
		RequestID:    dto.Request,
		Currency:     dto.Currency,
		Provider:     dto.Provider,
		Amount:       dto.Amount,
		PaymentDT:    dto.PaymentDT,
		Bank:         dto.Bank,
		DeliveryCost: dto.DeliveryCost,
		GoodsTotal:   dto.GoodsTotal,
		CustomFee:    dto.CustomFee,
	}
}

func ItemDTOToModel(dto kafka.ItemDTO) model.Item {
	return model.Item{
		ChrtID:      dto.ChrtID,
		TrackNumber: dto.TrackNumber,
		Price:       dto.Price,
		Rid:         dto.Rid,
		Name:        dto.Name,
		Sale:        dto.Sale,
		Size:        dto.Size,
		TotalPrice:  dto.TotalPrice,
		NmID:        dto.NmID,
		Brand:       dto.Brand,
		Status:      dto.Status,
	}
}

// model -> DTO (если нужно отдавать обратно/в DLQ/в другой топик)

func OrderModelToDTO(o model.Order) kafka.OrderDTO {
	items := make([]kafka.ItemDTO, len(o.Items))
	for i, it := range o.Items {
		items[i] = ItemModelToDTO(it)
	}

	return kafka.OrderDTO{
		OrderUID:          o.OrderUUID,
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
		Delivery:          DeliveryModelToDTO(o.Delivery),
		Payment:           PaymentModelToDTO(o.Payment),
		Items:             items,
	}
}

func DeliveryModelToDTO(d model.Delivery) kafka.DeliveryDTO {
	return kafka.DeliveryDTO{
		Name:    d.Name,
		Phone:   d.Phone,
		Zip:     d.Zip,
		City:    d.City,
		Address: d.Address,
		Region:  d.Region,
		Email:   d.Email,
	}
}

func PaymentModelToDTO(p model.Payment) kafka.PaymentDTO {
	return kafka.PaymentDTO{
		Transaction:  p.Transaction,
		Request:      p.RequestID,
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

func ItemModelToDTO(it model.Item) kafka.ItemDTO {
	return kafka.ItemDTO{
		ChrtID:      it.ChrtID,
		TrackNumber: it.TrackNumber,
		Price:       it.Price,
		Rid:         it.Rid,
		Name:        it.Name,
		Sale:        it.Sale,
		Size:        it.Size,
		TotalPrice:  it.TotalPrice,
		NmID:        it.NmID,
		Brand:       it.Brand,
		Status:      it.Status,
	}
}
