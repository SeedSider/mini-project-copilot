package api

import (
	pb "github.com/bankease/saving-service/protogen/saving-service"
	"github.com/bankease/saving-service/server/db"
)

func exchangeRatesToProto(items []db.ExchangeRate) []*pb.ExchangeRateItem {
	result := make([]*pb.ExchangeRateItem, len(items))
	for i, e := range items {
		result[i] = &pb.ExchangeRateItem{
			Id:          e.ID,
			Country:     e.Country,
			Currency:    e.Currency,
			CountryCode: e.CountryCode,
			Buy:         e.Buy,
			Sell:        e.Sell,
		}
	}
	return result
}

func interestRatesToProto(items []db.InterestRate) []*pb.InterestRateItem {
	result := make([]*pb.InterestRateItem, len(items))
	for i, ir := range items {
		result[i] = &pb.InterestRateItem{
			Id:      ir.ID,
			Kind:    ir.Kind,
			Deposit: ir.Deposit,
			Rate:    ir.Rate,
		}
	}
	return result
}

func branchesToProto(items []db.Branch) []*pb.BranchItem {
	result := make([]*pb.BranchItem, len(items))
	for i, b := range items {
		result[i] = &pb.BranchItem{
			Id:        b.ID,
			Name:      b.Name,
			Distance:  b.Distance,
			Latitude:  b.Latitude,
			Longitude: b.Longitude,
		}
	}
	return result
}
