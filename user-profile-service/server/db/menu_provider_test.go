package db

import (
	"context"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMenuID1  = "menu-1"
	testMenuID2  = "menu-2"
	testIconURL1 = "http://icon1.png"
	testIconURL2 = "http://icon2.png"
	testDBError  = "db error"
)

var menuCols = []string{"id", "index", "type", "title", "icon_url", "is_active"}

// ── GetAllMenus ──

func TestGetAllMenusSuccess(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(menuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true).
			AddRow(testMenuID2, 2, "PREMIUM", "Investasi", testIconURL2, true))

	menus, err := p.GetAllMenus(context.Background())
	require.NoError(t, err)
	assert.Len(t, menus, 2)
	assert.Equal(t, "Transfer", menus[0].Title)
	assert.Equal(t, "REGULAR", menus[0].Type)
	assert.Equal(t, "PREMIUM", menus[1].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllMenusEmpty(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(menuCols))

	menus, err := p.GetAllMenus(context.Background())
	require.NoError(t, err)
	assert.Empty(t, menus)
}

func TestGetAllMenusDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf(testDBError))

	menus, err := p.GetAllMenus(context.Background())
	assert.Error(t, err)
	assert.Nil(t, menus)
}

// ── GetMenusByAccountType ──

func TestGetMenusByAccountTypeRegular(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnRows(sqlmock.NewRows(menuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true))

	menus, err := p.GetMenusByAccountType(context.Background(), "REGULAR")
	require.NoError(t, err)
	assert.Len(t, menus, 1)
	assert.Equal(t, "REGULAR", menus[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenusByAccountTypePremium(t *testing.T) {
	p, mock := newTestProvider(t)

	// PREMIUM: returns all menus (no type filter, no args)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(menuCols).
			AddRow(testMenuID1, 1, "REGULAR", "Transfer", testIconURL1, true).
			AddRow(testMenuID2, 2, "PREMIUM", "Investasi", testIconURL2, true))

	menus, err := p.GetMenusByAccountType(context.Background(), "PREMIUM")
	require.NoError(t, err)
	assert.Len(t, menus, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenusByAccountTypeDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf(testDBError))

	// Use PREMIUM to avoid arg mismatch since we don't call WithArgs
	menus, err := p.GetMenusByAccountType(context.Background(), "PREMIUM")
	assert.Error(t, err)
	assert.Nil(t, menus)
}

func TestGetMenusByAccountTypeRegularDBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnError(fmt.Errorf(testDBError))

	menus, err := p.GetMenusByAccountType(context.Background(), "REGULAR")
	assert.Error(t, err)
	assert.Nil(t, menus)
}
