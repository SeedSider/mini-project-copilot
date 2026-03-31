package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testProfileID   = "profile-uuid-1"
	testUserID      = "user-uuid-1"
	testProfileName = "John Doe"
)

func newTestProvider(t *testing.T) (*Provider, sqlmock.Sqlmock) {
	t.Helper()
	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { dbConn.Close() })
	return New(dbConn), mock
}

var profileCols = []string{
	"id", "user_id", "bank", "branch", "name",
	"card_number", "card_provider", "balance", "currency", "account_type", "image",
}

func profileRow(id, userID string) *sqlmock.Rows {
	uid := userID
	return sqlmock.NewRows(profileCols).AddRow(
		id, &uid, "BRI", "Jakarta", testProfileName,
		"4111111111111111", "VISA", int64(1_000_000), "IDR", "REGULAR", "",
	)
}

// ── GetProfileByID ──

func TestGetProfileByIDSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testProfileID).
		WillReturnRows(profileRow(testProfileID, testUserID))

	profile, err := p.GetProfileByID(context.Background(), testProfileID)
	require.NoError(t, err)
	assert.Equal(t, testProfileID, profile.ID)
	assert.Equal(t, "BRI", profile.Bank)
	assert.Equal(t, "REGULAR", profile.AccountType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByIDNotFound(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	profile, err := p.GetProfileByID(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, profile)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByIDDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`SELECT id, user_id`).
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := p.GetProfileByID(context.Background(), "any")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ── GetProfileByUserID ──

func TestGetProfileByUserIDSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs(testUserID).
		WillReturnRows(profileRow(testProfileID, testUserID))

	profile, err := p.GetProfileByUserID(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, testProfileID, profile.ID)
	require.NotNil(t, profile.UserID)
	assert.Equal(t, testUserID, *profile.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByUserIDNotFound(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`SELECT id, user_id`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	profile, err := p.GetProfileByUserID(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, profile)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileByUserIDDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`SELECT id, user_id`).
		WillReturnError(fmt.Errorf("db error"))

	_, err := p.GetProfileByUserID(context.Background(), "any")
	assert.Error(t, err)
}

// ── UpdateProfile ──

func TestUpdateProfileSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectExec(`UPDATE profile`).
		WithArgs("BRI", "Bandung", "Jane Doe", "4222222222222222", testProfileID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := p.UpdateProfile(context.Background(), testProfileID, EditProfileRequest{
		Bank: "BRI", Branch: "Bandung", Name: "Jane Doe", CardNumber: "4222222222222222",
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProfileNotFound(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectExec(`UPDATE profile`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := p.UpdateProfile(context.Background(), "nonexistent-uuid", EditProfileRequest{
		Bank: "BRI", Branch: "Jakarta", Name: "Test", CardNumber: "1234",
	})
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestUpdateProfileDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectExec(`UPDATE profile`).
		WillReturnError(fmt.Errorf("db error"))

	err := p.UpdateProfile(context.Background(), "any-uuid", EditProfileRequest{})
	assert.Error(t, err)
}

// ── CreateProfile ──

func TestCreateProfileSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	uid := testUserID
	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnRows(sqlmock.NewRows(profileCols).AddRow(
			"new-profile-uuid", &uid, "BRI", "Jakarta", testProfileName,
			"4111111111111111", "VISA", int64(500_000), "IDR", "REGULAR", "",
		))

	profile, err := p.CreateProfile(context.Background(), CreateProfileRequest{
		UserID:       testUserID,
		Bank:         "BRI",
		Branch:       "Jakarta",
		Name:         testProfileName,
		CardNumber:   "4111111111111111",
		CardProvider: "VISA",
		Balance:      500_000,
		Currency:     "IDR",
		AccountType:  "REGULAR",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-profile-uuid", profile.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProfileDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`INSERT INTO profile`).
		WillReturnError(fmt.Errorf("constraint violation"))

	profile, err := p.CreateProfile(context.Background(), CreateProfileRequest{UserID: testUserID})
	assert.Error(t, err)
	assert.Nil(t, profile)
}
