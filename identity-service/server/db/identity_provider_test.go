package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database"
	databasemock "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database/mock"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/logger"
)

func newTestProvider(t *testing.T) (*Provider, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dbSql := &database.DbSql{
		SqlDb: db,
		Dbw:   &databasemock.DatabaseMock{DbPq: db},
	}
	dbSql.Conn = db

	testLogger := logger.New(&logger.LoggerConfig{
		Env:         "DEV",
		ServiceName: "test",
		ProductName: "test",
		LogLevel:    "error",
		LogOutput:   "stdout",
	})

	provider := New(dbSql, testLogger)
	return provider, mock
}

// --- CreateUser ---

func TestCreateUser_Success(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("johndoe", "hashedpw", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-uuid-123"))

	result, err := provider.CreateUser(ctx, "johndoe", "hashedpw", "")
	require.NoError(t, err)
	assert.Equal(t, "user-uuid-123", result.UserID)
	assert.Equal(t, "johndoe", result.Username)
	assert.Nil(t, result.Phone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_WithPhone(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	phone := "08123456789"
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("johndoe", "hashedpw", &phone).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-uuid-456"))

	result, err := provider.CreateUser(ctx, "johndoe", "hashedpw", "08123456789")
	require.NoError(t, err)
	assert.Equal(t, "user-uuid-456", result.UserID)
	assert.NotNil(t, result.Phone)
	assert.Equal(t, "08123456789", *result.Phone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_DBError(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(fmt.Errorf("duplicate key"))

	result, err := provider.CreateUser(ctx, "johndoe", "hashedpw", "")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- GetUserByUsername ---

func TestGetUserByUsername_Success(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE username`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
			AddRow("user-123", "johndoe", "$2a$10$hash", "2026-03-27T00:00:00Z"))

	user, err := provider.GetUserByUsername(ctx, "johndoe")
	require.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "johndoe", user.Username)
	assert.Equal(t, "$2a$10$hash", user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE username`).
		WithArgs("unknown").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at"}))

	user, err := provider.GetUserByUsername(ctx, "unknown")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsername_DBError(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE username`).
		WithArgs("johndoe").
		WillReturnError(fmt.Errorf("connection refused"))

	user, err := provider.GetUserByUsername(ctx, "johndoe")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- CheckUsernameExists ---

func TestCheckUsernameExists_True(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users WHERE username`).
		WithArgs("johndoe").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := provider.CheckUsernameExists(ctx, "johndoe")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUsernameExists_False(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users WHERE username`).
		WithArgs("unknown").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := provider.CheckUsernameExists(ctx, "unknown")
	require.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUsernameExists_DBError(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM users WHERE username`).
		WithArgs("johndoe").
		WillReturnError(fmt.Errorf("connection refused"))

	exists, err := provider.CheckUsernameExists(ctx, "johndoe")
	assert.Error(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Provider ---

func TestNew(t *testing.T) {
	testLogger := logger.New(&logger.LoggerConfig{
		Env:         "DEV",
		ServiceName: "test",
		ProductName: "test",
		LogLevel:    "error",
		LogOutput:   "stdout",
	})
	dbSql := database.InitConnectionDB("postgres", database.Config{}, &databasemock.DatabaseMock{})
	provider := New(dbSql, testLogger)
	assert.NotNil(t, provider)
	assert.Equal(t, dbSql, provider.GetDbSql())
}

// --- NotFoundErr ---

func TestNotFoundErr(t *testing.T) {
	err := NotFound("User", "123")
	assert.Equal(t, "User with id '123' not found", err.Error())
	assert.Equal(t, "User", err.ResourceType)
	assert.Equal(t, "123", err.ID)
}

// --- UpdatePasswordByUsername ---

func TestUpdatePasswordByUsername_Success(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WithArgs("new-hashed-pw", "johndoe").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := provider.UpdatePasswordByUsername(ctx, "johndoe", "new-hashed-pw")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePasswordByUsername_UserNotFound(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WithArgs("new-hashed-pw", "unknown").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := provider.UpdatePasswordByUsername(ctx, "unknown", "new-hashed-pw")
	assert.Error(t, err)
	_, isNotFound := err.(NotFoundErr)
	assert.True(t, isNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePasswordByUsername_DBError(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WillReturnError(fmt.Errorf("connection refused"))

	err := provider.UpdatePasswordByUsername(ctx, "johndoe", "new-hashed-pw")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePasswordByUsername_RowsAffectedError(t *testing.T) {
	provider, mock := newTestProvider(t)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WithArgs("new-hashed-pw", "johndoe").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err := provider.UpdatePasswordByUsername(ctx, "johndoe", "new-hashed-pw")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
