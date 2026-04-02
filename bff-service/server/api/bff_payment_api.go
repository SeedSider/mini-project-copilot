package api

import (
	"context"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	paymentPB "github.com/bankease/bff-service/protogen/payment-service"
	"github.com/bankease/bff-service/server/utils"
	"google.golang.org/grpc/metadata"
)

// forwardAuthCtx converts incoming gRPC metadata (set by contextFromHTTPRequest)
// into outgoing metadata so downstream gRPC services receive the Authorization header.
func forwardAuthCtx(ctx context.Context) context.Context {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

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

	resp, err := s.svcConn.PaymentClient().GetInternetBill(forwardAuthCtx(ctx), &paymentPB.GetInternetBillRequest{})
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

// GetBeneficiaries proxies to payment-service (JWT required).
func (s *Server) GetBeneficiaries(ctx context.Context, req *pb.GetBeneficiariesRequest) (*pb.BeneficiaryListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetBeneficiaries", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().GetBeneficiaries(forwardAuthCtx(ctx), &paymentPB.GetBeneficiariesRequest{
		AccountId: req.GetAccountId(),
	})
	if err != nil {
		log.Error(processId, "GetBeneficiaries", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.BeneficiaryItem, len(resp.Beneficiaries))
	for i, b := range resp.Beneficiaries {
		items[i] = &pb.BeneficiaryItem{
			Id:     b.Id,
			Name:   b.Name,
			Phone:  b.Phone,
			Avatar: b.Avatar,
		}
	}

	return &pb.BeneficiaryListResponse{Beneficiaries: items}, nil
}

// PrepaidPay proxies to payment-service (JWT required).
func (s *Server) PrepaidPay(ctx context.Context, req *pb.PrepaidPayRequest) (*pb.PrepaidPayResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "PrepaidPay", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().PrepaidPay(forwardAuthCtx(ctx), &paymentPB.PrepaidPayRequest{
		CardId:         req.GetCardId(),
		Phone:          req.GetPhone(),
		Amount:         req.GetAmount(),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		log.Error(processId, "PrepaidPay", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	return &pb.PrepaidPayResponse{
		Id:        resp.Id,
		Status:    resp.Status,
		Message:   resp.Message,
		Timestamp: resp.Timestamp,
	}, nil
}

// AddBeneficiary proxies to payment-service (JWT required).
func (s *Server) AddBeneficiary(ctx context.Context, req *pb.AddBeneficiaryRequest) (*pb.BeneficiaryItem, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "AddBeneficiary", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().AddBeneficiary(forwardAuthCtx(ctx), &paymentPB.AddBeneficiaryRequest{
		AccountId: req.GetAccountId(),
		Name:      req.GetName(),
		Phone:     req.GetPhone(),
		Avatar:    req.GetAvatar(),
	})
	if err != nil {
		log.Error(processId, "AddBeneficiary", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	return &pb.BeneficiaryItem{
		Id:     resp.Id,
		Name:   resp.Name,
		Phone:  resp.Phone,
		Avatar: resp.Avatar,
	}, nil
}

// SearchBeneficiaries proxies to payment-service (JWT required).
func (s *Server) SearchBeneficiaries(ctx context.Context, req *pb.SearchBeneficiariesRequest) (*pb.BeneficiaryListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "SearchBeneficiaries", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().SearchBeneficiaries(forwardAuthCtx(ctx), &paymentPB.SearchBeneficiariesRequest{
		AccountId: req.GetAccountId(),
		Query:     req.GetQuery(),
	})
	if err != nil {
		log.Error(processId, "SearchBeneficiaries", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	items := make([]*pb.BeneficiaryItem, len(resp.Beneficiaries))
	for i, b := range resp.Beneficiaries {
		items[i] = &pb.BeneficiaryItem{
			Id:     b.Id,
			Name:   b.Name,
			Phone:  b.Phone,
			Avatar: b.Avatar,
		}
	}

	return &pb.BeneficiaryListResponse{Beneficiaries: items}, nil
}

// GetPaymentCards proxies to payment-service (JWT required).
func (s *Server) GetPaymentCards(ctx context.Context, req *pb.GetPaymentCardsRequest) (*pb.PaymentCardListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetPaymentCards", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().GetPaymentCards(forwardAuthCtx(ctx), &paymentPB.GetPaymentCardsRequest{
		AccountId: req.GetAccountId(),
	})
	if err != nil {
		log.Error(processId, "GetPaymentCards", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	cards := make([]*pb.PaymentCardItem, len(resp.Cards))
	for i, c := range resp.Cards {
		cards[i] = &pb.PaymentCardItem{
			Id:             c.Id,
			AccountId:      c.AccountId,
			HolderName:     c.HolderName,
			CardLabel:      c.CardLabel,
			MaskedNumber:   c.MaskedNumber,
			Balance:        c.Balance,
			Currency:       c.Currency,
			Brand:          c.Brand,
			GradientColors: c.GradientColors,
		}
	}

	return &pb.PaymentCardListResponse{Cards: cards}, nil
}

// CreatePaymentCard proxies to payment-service (JWT required).
func (s *Server) CreatePaymentCard(ctx context.Context, req *pb.CreatePaymentCardRequest) (*pb.PaymentCardItem, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "CreatePaymentCard", "Proxying to payment-service", nil, nil, nil, nil)

	resp, err := s.svcConn.PaymentClient().CreatePaymentCard(forwardAuthCtx(ctx), &paymentPB.CreatePaymentCardRequest{
		AccountId:      req.GetAccountId(),
		HolderName:     req.GetHolderName(),
		CardLabel:      req.CardLabel,
		MaskedNumber:   req.MaskedNumber,
		Balance:        req.Balance,
		Currency:       req.Currency,
		Brand:          req.Brand,
		GradientColors: req.GradientColors,
	})
	if err != nil {
		log.Error(processId, "CreatePaymentCard", err.Error(), nil, nil, nil, err)
		return nil, err
	}

	return &pb.PaymentCardItem{
		Id:             resp.Id,
		AccountId:      resp.AccountId,
		HolderName:     resp.HolderName,
		CardLabel:      resp.CardLabel,
		MaskedNumber:   resp.MaskedNumber,
		Balance:        resp.Balance,
		Currency:       resp.Currency,
		Brand:          resp.Brand,
		GradientColors: resp.GradientColors,
	}, nil
}
