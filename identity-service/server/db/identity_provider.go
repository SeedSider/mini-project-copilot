package db

import (
	"context"
	"fmt"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/utils"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    string
}

type Profile struct {
	ID        string
	UserID    string
	FullName  string
	Phone     *string
	CreatedAt string
}

type UserWithProfile struct {
	UserID   string
	Email    string
	FullName string
	Phone    *string
}

func (p *Provider) CreateUserWithProfile(ctx context.Context, email, passwordHash, fullName, phone string) (*UserWithProfile, error) {
	const functionName = "CreateUserWithProfile"
	processId := utils.GetProcessIdFromCtx(ctx)

	tx, err := p.dbSql.GetPmConnection().Begin()
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] begin tx: %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	var userID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&userID)
	if err != nil {
		tx.Rollback()
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] insert user: %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	var phoneVal *string
	if phone != "" {
		phoneVal = &phone
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO profiles (user_id, full_name, phone) VALUES ($1, $2, $3)`,
		userID, fullName, phoneVal,
	)
	if err != nil {
		tx.Rollback()
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] insert profile: %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] commit tx: %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	log.Info(processId, functionName, fmt.Sprintf("User created: %s", userID), nil, nil, nil, nil)
	return &UserWithProfile{
		UserID:   userID,
		Email:    email,
		FullName: fullName,
		Phone:    phoneVal,
	}, nil
}

func (p *Provider) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const functionName = "GetUserByEmail"
	processId := utils.GetProcessIdFromCtx(ctx)

	var user User
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	return &user, nil
}

func (p *Provider) GetProfileByUserID(ctx context.Context, userID string) (*Profile, error) {
	const functionName = "GetProfileByUserID"
	processId := utils.GetProcessIdFromCtx(ctx)

	var profile Profile
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`SELECT id, user_id, full_name, phone, created_at FROM profiles WHERE user_id = $1`,
		userID,
	).Scan(&profile.ID, &profile.UserID, &profile.FullName, &profile.Phone, &profile.CreatedAt)
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] %v", functionName, err), nil, nil, nil, err)
		return nil, err
	}

	return &profile, nil
}

func (p *Provider) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	const functionName = "CheckEmailExists"
	processId := utils.GetProcessIdFromCtx(ctx)

	var count int
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`SELECT COUNT(1) FROM users WHERE email = $1`,
		email,
	).Scan(&count)
	if err != nil {
		log.Error(processId, functionName, fmt.Sprintf("[error][db][func: %s] %v", functionName, err), nil, nil, nil, err)
		return false, err
	}

	return count > 0, nil
}
