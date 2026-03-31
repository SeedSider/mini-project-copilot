package api

import (
	"context"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	savingPB "github.com/bankease/bff-service/protogen/saving-service"
	"github.com/bankease/bff-service/server/utils"
)

// GetExchangeRates proxies to saving-service.
func (s *Server) GetExchangeRates(ctx context.Context, _ *pb.GetExchangeRatesRequest) (*pb.ExchangeRateListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetExchangeRates", "Proxying to saving-service", nil, nil, nil, nil)

	resp, err := s.svcConn.SavingClient().GetExchangeRates(ctx, &savingPB.GetExchangeRatesRequest{})
	if err != nil {
		log.Error(processId, "GetExchangeRates", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.ExchangeRateItem, len(resp.ExchangeRates))
	for i, e := range resp.ExchangeRates {
		items[i] = &pb.ExchangeRateItem{
			Id:          e.Id,
			Country:     e.Country,
			Currency:    e.Currency,
			CountryCode: e.CountryCode,
			Buy:         e.Buy,
			Sell:        e.Sell,
		}
	}

	return &pb.ExchangeRateListResponse{ExchangeRates: items}, nil
}

// GetInterestRates proxies to saving-service.
func (s *Server) GetInterestRates(ctx context.Context, _ *pb.GetInterestRatesRequest) (*pb.InterestRateListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetInterestRates", "Proxying to saving-service", nil, nil, nil, nil)

	resp, err := s.svcConn.SavingClient().GetInterestRates(ctx, &savingPB.GetInterestRatesRequest{})
	if err != nil {
		log.Error(processId, "GetInterestRates", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.InterestRateItem, len(resp.InterestRates))
	for i, ir := range resp.InterestRates {
		items[i] = &pb.InterestRateItem{
			Id:      ir.Id,
			Kind:    ir.Kind,
			Deposit: ir.Deposit,
			Rate:    ir.Rate,
		}
	}

	return &pb.InterestRateListResponse{InterestRates: items}, nil
}

// GetBranches proxies to saving-service.
func (s *Server) GetBranches(ctx context.Context, req *pb.GetBranchesRequest) (*pb.BranchListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetBranches", "Proxying to saving-service", nil, nil, nil, nil)

	resp, err := s.svcConn.SavingClient().GetBranches(ctx, &savingPB.GetBranchesRequest{
		Query: req.GetQuery(),
	})
	if err != nil {
		log.Error(processId, "GetBranches", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.BranchItem, len(resp.Branches))
	for i, b := range resp.Branches {
		items[i] = &pb.BranchItem{
			Id:        b.Id,
			Name:      b.Name,
			Distance:  b.Distance,
			Latitude:  b.Latitude,
			Longitude: b.Longitude,
		}
	}

	return &pb.BranchListResponse{Branches: items}, nil
}
