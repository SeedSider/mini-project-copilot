package services

import (
	"encoding/json"
	"log"

	identityPB "github.com/bankease/bff-service/protogen/identity-service"
	paymentPB "github.com/bankease/bff-service/protogen/payment-service"
	savingPB "github.com/bankease/bff-service/protogen/saving-service"
	profilePB "github.com/bankease/bff-service/protogen/user-profile-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceConnection holds gRPC client connections to downstream services.
type ServiceConnection struct {
	IdentityService    *grpc.ClientConn
	UserProfileService *grpc.ClientConn
	SavingService      *grpc.ClientConn
	PaymentService     *grpc.ClientConn
}

// InitServicesConn initializes all gRPC client connections.
func InitServicesConn(identityAddr, profileAddr, savingAddr, paymentAddr string) *ServiceConnection {
	services := &ServiceConnection{}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(JSONCodec{})),
	}

	var err error

	services.IdentityService, err = grpc.Dial(identityAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to identity-service at %s: %v", identityAddr, err)
	}

	services.UserProfileService, err = grpc.Dial(profileAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to user-profile-service at %s: %v", profileAddr, err)
	}

	services.SavingService, err = grpc.Dial(savingAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to saving-service at %s: %v", savingAddr, err)
	}

	services.PaymentService, err = grpc.Dial(paymentAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to payment-service at %s: %v", paymentAddr, err)
	}

	return services
}

// IdentityClient returns a new identity-service gRPC client.
func (sc *ServiceConnection) IdentityClient() identityPB.IdentityServiceClient {
	return identityPB.NewIdentityServiceClient(sc.IdentityService)
}

// UserProfileClient returns a new user-profile-service gRPC client.
func (sc *ServiceConnection) UserProfileClient() profilePB.UserProfileServiceClient {
	return profilePB.NewUserProfileServiceClient(sc.UserProfileService)
}

// SavingClient returns a new saving-service gRPC client.
func (sc *ServiceConnection) SavingClient() savingPB.SavingServiceClient {
	return savingPB.NewSavingServiceClient(sc.SavingService)
}

// PaymentClient returns a new payment-service gRPC client.
func (sc *ServiceConnection) PaymentClient() paymentPB.PaymentServiceClient {
	return paymentPB.NewPaymentServiceClient(sc.PaymentService)
}

// Close closes all gRPC client connections.
func (sc *ServiceConnection) Close() {
	if sc.IdentityService != nil {
		sc.IdentityService.Close()
	}
	if sc.UserProfileService != nil {
		sc.UserProfileService.Close()
	}
	if sc.SavingService != nil {
		sc.SavingService.Close()
	}
	if sc.PaymentService != nil {
		sc.PaymentService.Close()
	}
}

// JSONCodec is a JSON-based gRPC codec for hand-written protobuf types.
type JSONCodec struct{}

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (JSONCodec) Name() string {
	return "json"
}
