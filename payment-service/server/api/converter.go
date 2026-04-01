package api

import (
	pb "github.com/bankease/payment-service/protogen/payment-service"
	"github.com/bankease/payment-service/server/db"
)

func providersToProto(providers []db.ServiceProvider) []*pb.ProviderItem {
	items := make([]*pb.ProviderItem, len(providers))
	for i, p := range providers {
		items[i] = &pb.ProviderItem{
			Id:   p.ID,
			Name: p.Name,
		}
	}
	return items
}

func internetBillToProto(bill *db.InternetBill) *pb.InternetBillDetail {
	return &pb.InternetBillDetail{
		CustomerId:  bill.CustomerID,
		Name:        bill.Name,
		Address:     bill.Address,
		PhoneNumber: bill.PhoneNumber,
		Code:        bill.Code,
		From:        bill.From,
		To:          bill.To,
		InternetFee: bill.InternetFee,
		Tax:         bill.Tax,
		Total:       bill.Total,
	}
}

func currenciesToProto(currencies []db.Currency) []*pb.CurrencyEntry {
	items := make([]*pb.CurrencyEntry, len(currencies))
	for i, c := range currencies {
		items[i] = &pb.CurrencyEntry{
			Code:  c.Code,
			Label: c.Label,
			Rate:  c.Rate,
		}
	}
	return items
}
