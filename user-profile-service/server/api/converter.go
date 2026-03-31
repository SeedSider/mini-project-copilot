package api

import (
	pb "github.com/bankease/user-profile-service/protogen/user-profile-service"
	"github.com/bankease/user-profile-service/server/db"
)

func profileToProto(p *db.Profile) *pb.ProfileResponse {
	userId := ""
	if p.UserID != nil {
		userId = *p.UserID
	}
	return &pb.ProfileResponse{
		Id:           p.ID,
		UserId:       userId,
		Bank:         p.Bank,
		Branch:       p.Branch,
		Name:         p.Name,
		CardNumber:   p.CardNumber,
		CardProvider: p.CardProvider,
		Balance:      p.Balance,
		Currency:     p.Currency,
		AccountType:  p.AccountType,
		Image:        p.Image,
	}
}

func menusToProto(menus []db.Menu) []*pb.MenuItem {
	items := make([]*pb.MenuItem, len(menus))
	for i, m := range menus {
		items[i] = &pb.MenuItem{
			Id:       m.ID,
			Index:    int32(m.Index),
			Type:     m.Type,
			Title:    m.Title,
			IconUrl:  m.IconURL,
			IsActive: m.IsActive,
		}
	}
	return items
}

func createReqToModel(req *pb.CreateProfileRequest) db.CreateProfileRequest {
	return db.CreateProfileRequest{
		UserID:       req.GetUserId(),
		Bank:         req.GetBank(),
		Branch:       req.GetBranch(),
		Name:         req.GetName(),
		CardNumber:   req.GetCardNumber(),
		CardProvider: req.GetCardProvider(),
		Balance:      req.GetBalance(),
		Currency:     req.GetCurrency(),
		AccountType:  req.GetAccountType(),
		Image:        req.GetImage(),
	}
}

func updateReqToModel(req *pb.UpdateProfileRequest) (string, db.EditProfileRequest) {
	return req.GetId(), db.EditProfileRequest{
		Bank:       req.GetBank(),
		Branch:     req.GetBranch(),
		Name:       req.GetName(),
		CardNumber: req.GetCardNumber(),
	}
}
