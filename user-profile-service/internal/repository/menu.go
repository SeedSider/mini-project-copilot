package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/bankease/user-profile-service/internal/models"
)

// MenuRepository handles database operations for menu.
type MenuRepository struct {
	DB *sql.DB
}

// GetAllMenus retrieves all active menus ordered by index.
func (r *MenuRepository) GetAllMenus(ctx context.Context) ([]models.Menu, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, "index", type, title, icon_url, is_active
		FROM menu WHERE is_active = TRUE ORDER BY "index" ASC`

	return r.scanMenus(ctx, query)
}

// GetMenusByAccountType retrieves menus filtered by account type.
// PREMIUM → all menus; REGULAR → only REGULAR type menus.
func (r *MenuRepository) GetMenusByAccountType(ctx context.Context, accountType string) ([]models.Menu, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if accountType == "PREMIUM" {
		query := `SELECT id, "index", type, title, icon_url, is_active
			FROM menu WHERE is_active = TRUE ORDER BY "index" ASC`
		return r.scanMenus(ctx, query)
	}

	query := `SELECT id, "index", type, title, icon_url, is_active
		FROM menu WHERE type = $1 AND is_active = TRUE ORDER BY "index" ASC`
	return r.scanMenusWithParam(ctx, query, accountType)
}

func (r *MenuRepository) scanMenus(ctx context.Context, query string) ([]models.Menu, error) {
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		if err := rows.Scan(&m.ID, &m.Index, &m.Type, &m.Title, &m.IconURL, &m.IsActive); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if menus == nil {
		menus = []models.Menu{}
	}

	return menus, nil
}

func (r *MenuRepository) scanMenusWithParam(ctx context.Context, query string, param string) ([]models.Menu, error) {
	rows, err := r.DB.QueryContext(ctx, query, param)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		if err := rows.Scan(&m.ID, &m.Index, &m.Type, &m.Title, &m.IconURL, &m.IsActive); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if menus == nil {
		menus = []models.Menu{}
	}

	return menus, nil
}
