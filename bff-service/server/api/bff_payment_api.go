package api

import (
	"context"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	paymentPB "github.com/bankease/bff-service/protogen/payment-service"
	"github.com/bankease/bff-service/server/utils"
)

// GetProviders proxies to payment-service.
func (s *Server) GetProviders(ctx context.Context, _ *pb.GetProvidersRequest) (*pb.ProviderListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetProviders", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().GetProviders(ctx, &paymentPB.GetProvidersRequest{})
	if err != nil {
		log.Error(processId, "GetProviders", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.ProviderItem, len(resp.Providers))
	for i, p := range resp.Providers {
		items[i] = &pb.ProviderItem{
			Id:   p.Id,
			Name: p.Name,
		}
	}

	return &pb.ProviderListResponse{Providers: items}, nil
}

// GetInternetBill proxies to payment-service (JWT required — user_claims forwarded via context).
func (s *Server) GetInternetBill(ctx context.Context, _ *pb.GetInternetBillRequest) (*pb.InternetBillResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetInternetBill", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().GetInternetBill(ctx, &paymentPB.GetInternetBillRequest{})
	if err != nil {
		log.Error(processId, "GetInternetBill", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	var bill *pb.InternetBillDetail
	if resp.Bill != nil {
		bill = &pb.InternetBillDetail{
			CustomerId:  resp.Bill.CustomerId,
			Name:        resp.Bill.Name,
			Address:     resp.Bill.Address,
			PhoneNumber: resp.Bill.PhoneNumber,
			Code:        resp.Bill.Code,
			From:        resp.Bill.From,
			To:          resp.Bill.To,
			InternetFee: resp.Bill.InternetFee,
			Tax:         resp.Bill.Tax,
			Total:       resp.Bill.Total,
		}
	}

	return &pb.InternetBillResponse{Bill: bill}, nil
}

// GetCurrencyList proxies to payment-service.
func (s *Server) GetCurrencyList(ctx context.Context, _ *pb.GetCurrencyListRequest) (*pb.CurrencyListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetCurrencyList", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().GetCurrencyList(ctx, &paymentPB.GetCurrencyListRequest{})
	if err != nil {
		log.Error(processId, "GetCurrencyList", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.CurrencyEntry, len(resp.Currencies))
	for i, c := range resp.Currencies {
		items[i] = &pb.CurrencyEntry{
			Code:  c.Code,
			Label: c.Label,
			Rate:  c.Rate,
		}
	}

	return &pb.CurrencyListResponse{Currencies: items}, nil
}
