package db

import (
	"context"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var menuCols = []string{"id", "index", "type", "title", "icon_url", "is_active"}

// ── GetAllMenus ──

func TestGetAllMenus_Success(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(menuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true).
			AddRow("menu-2", 2, "PREMIUM", "Investasi", "http://icon2.png", true))

	menus, err := p.GetAllMenus(context.Background())
	require.NoError(t, err)
	assert.Len(t, menus, 2)
	assert.Equal(t, "Transfer", menus[0].Title)
	assert.Equal(t, "REGULAR", menus[0].Type)
	assert.Equal(t, "PREMIUM", menus[1].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllMenus_Empty(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(menuCols))

	menus, err := p.GetAllMenus(context.Background())
	require.NoError(t, err)
	assert.Empty(t, menus)
}

func TestGetAllMenus_DBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf("db error"))

	menus, err := p.GetAllMenus(context.Background())
	assert.Error(t, err)
	assert.Nil(t, menus)
}

// ── GetMenusByAccountType ──

func TestGetMenusByAccountType_Regular(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnRows(sqlmock.NewRows(menuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true))

	menus, err := p.GetMenusByAccountType(context.Background(), "REGULAR")
	require.NoError(t, err)
	assert.Len(t, menus, 1)
	assert.Equal(t, "REGULAR", menus[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenusByAccountType_Premium(t *testing.T) {
	p, mock := newTestProvider(t)

	// PREMIUM: returns all menus (no type filter, no args)
	mock.ExpectQuery(`FROM menu`).
		WillReturnRows(sqlmock.NewRows(menuCols).
			AddRow("menu-1", 1, "REGULAR", "Transfer", "http://icon1.png", true).
			AddRow("menu-2", 2, "PREMIUM", "Investasi", "http://icon2.png", true))

	menus, err := p.GetMenusByAccountType(context.Background(), "PREMIUM")
	require.NoError(t, err)
	assert.Len(t, menus, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMenusByAccountType_DBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WillReturnError(fmt.Errorf("db error"))

	// Use PREMIUM to avoid arg mismatch since we don't call WithArgs
	menus, err := p.GetMenusByAccountType(context.Background(), "PREMIUM")
	assert.Error(t, err)
	assert.Nil(t, menus)
}

func TestGetMenusByAccountType_Regular_DBError(t *testing.T) {
	p, mock := newTestProvider(t)

	mock.ExpectQuery(`FROM menu`).
		WithArgs("REGULAR").
		WillReturnError(fmt.Errorf("db error"))

	menus, err := p.GetMenusByAccountType(context.Background(), "REGULAR")
	assert.Error(t, err)
	assert.Nil(t, menus)
}
