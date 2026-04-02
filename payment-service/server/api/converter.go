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

func beneficiariesToProto(beneficiaries []db.Beneficiary) []*pb.BeneficiaryItem {
	items := make([]*pb.BeneficiaryItem, len(beneficiaries))
	for i, b := range beneficiaries {
		items[i] = &pb.BeneficiaryItem{
			Id:     b.ID,
			Name:   b.Name,
			Phone:  b.Phone,
			Avatar: b.Avatar,
		}
	}
	return items
}

func transactionToProto(txn *db.PrepaidTransaction) *pb.PrepaidPayResponse {
	return &pb.PrepaidPayResponse{
		Id:        txn.ID,
		Status:    txn.Status,
		Message:   txn.Message,
		Timestamp: txn.Timestamp,
	}
}

func cardsToProto(cards []db.PaymentCard) []*pb.PaymentCardItem {
	items := make([]*pb.PaymentCardItem, len(cards))
	for i, c := range cards {
		items[i] = cardToProto(&c)
	}
	return items
}

func cardToProto(c *db.PaymentCard) *pb.PaymentCardItem {
	return &pb.PaymentCardItem{
		Id:             c.ID,
		AccountId:      c.AccountID,
		HolderName:     c.HolderName,
		CardLabel:      c.CardLabel,
		MaskedNumber:   c.MaskedNumber,
		Balance:        c.Balance,
		Currency:       c.Currency,
		Brand:          c.Brand,
		GradientColors: c.GradientColors,
	}
}
