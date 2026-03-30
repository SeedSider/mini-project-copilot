package services

import (
	"encoding/json"
	"log"

	identityPB "github.com/bankease/bff-service/protogen/identity-service"
	profilePB "github.com/bankease/bff-service/protogen/user-profile-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceConnection holds gRPC client connections to downstream services.
type ServiceConnection struct {
	IdentityService    *grpc.ClientConn
	UserProfileService *grpc.ClientConn
}

// InitServicesConn initializes all gRPC client connections.
func InitServicesConn(identityAddr, profileAddr string) *ServiceConnection {
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

// Close closes all gRPC client connections.
func (sc *ServiceConnection) Close() {
	if sc.IdentityService != nil {
		sc.IdentityService.Close()
	}
	if sc.UserProfileService != nil {
		sc.UserProfileService.Close()
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
