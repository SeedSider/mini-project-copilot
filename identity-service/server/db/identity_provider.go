package db

import (
	"context"
	"fmt"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/utils"
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    string
}

type UserWithProfile struct {
	UserID   string
	Username string
	Phone    *string
}

func (p *Provider) CreateUser(ctx context.Context, username, passwordHash, phone string) (*UserWithProfile, error) {
	const functionName = "CreateUser"
	processId := utils.GetProcessIdFromCtx(ctx)

	var phoneVal *string
	if phone != "" {
		phoneVal = &phone
	}

	var userID string
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash, phone) VALUES ($1, $2, $3) RETURNING id`,
		username, passwordHash, phoneVal,
	).Scan(&userID)
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] insert user: %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	log.Info(processId, functionName, fmt.Sprintf("User created: %s", userID), nil, nil, nil, nil)
	return &UserWithProfile{
		UserID:   userID,
		Username: username,
		Phone:    phoneVal,
	}, nil
}

func (p *Provider) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	const functionName = "GetUserByUsername"
	processId := utils.GetProcessIdFromCtx(ctx)

	var user User
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	return &user, nil
}

func (p *Provider) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	const functionName = "CheckUsernameExists"
	processId := utils.GetProcessIdFromCtx(ctx)

	var count int
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`SELECT COUNT(1) FROM users WHERE username = $1`,
		username,
	).Scan(&count)
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] %v", functionName, err), nil, nil, nil, err)
		return false, err
	}

	return count > 0, nil
}
