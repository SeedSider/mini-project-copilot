package api

import (
	"context"
	"fmt"

	pb "github.com/bankease/bff-service/protogen/bff-service"
	profilePB "github.com/bankease/bff-service/protogen/user-profile-service"
	"github.com/bankease/bff-service/server/utils"
)

// GetAllMenus proxies to user-profile-service.
func (s *Server) GetAllMenus(ctx context.Context, _ *pb.GetAllMenusRequest) (*pb.MenuListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetAllMenus", "Processing get all menus request", nil, nil, nil, nil)

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	resp, err := s.svcConn.UserProfileClient().GetAllMenus(outCtx, &profilePB.GetAllMenusRequest{})
	if err != nil {
		log.Error(processId, "GetAllMenus", fmt.Sprintf("profile.GetAllMenus failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return toMenuListResponse(resp), nil
}

// GetMenusByAccountType proxies to user-profile-service.
func (s *Server) GetMenusByAccountType(ctx context.Context, req *pb.GetMenusByAccountTypeRequest) (*pb.MenuListResponse, error) {
	processId := utils.GetProcessIdFromCtx(ctx)
	log.Info(processId, "GetMenusByAccountType", "Processing get menus by account type request", nil, nil, nil, nil)

	outCtx := utils.CreateNewContextWithProcessId(ctx, nil)
	resp, err := s.svcConn.UserProfileClient().GetMenusByAccountType(outCtx, &profilePB.GetMenusByAccountTypeRequest{
		AccountType: req.GetAccountType(),
	})
	if err != nil {
		log.Error(processId, "GetMenusByAccountType", fmt.Sprintf("profile.GetMenusByAccountType failed: %v", err), nil, nil, nil, err)
		return nil, err
	}

	return toMenuListResponse(resp), nil
}

// toMenuListResponse converts user-profile-service MenuListResponse to BFF MenuListResponse.
func toMenuListResponse(resp *profilePB.MenuListResponse) *pb.MenuListResponse {
	menus := make([]*pb.MenuItem, len(resp.Menus))
	for i, m := range resp.Menus {
		menus[i] = &pb.MenuItem{
			Id:       m.Id,
			Index:    m.Index,
			Type:     m.Type,
			Title:    m.Title,
			IconUrl:  m.IconUrl,
			IsActive: m.IsActive,
		}
	}
	return &pb.MenuListResponse{Menus: menus}
}
