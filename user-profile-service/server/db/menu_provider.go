package db

import (
	"context"
	"database/sql"
	"time"
)

// Menu represents a homepage menu item record in the database.
type Menu struct {
	ID       string `json:"id"`
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	IconURL  string `json:"icon_url"`
	IsActive bool   `json:"is_active"`
}

// MenuResponse wraps a list of menus for the API response.
type MenuResponse struct {
	Menus []Menu `json:"menus"`
}

// GetAllMenus retrieves all active menus ordered by index.
func (p *Provider) GetAllMenus(ctx context.Context) ([]Menu, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, "index", type, title, icon_url, is_active
		FROM menu WHERE is_active = TRUE ORDER BY "index" ASC`

	return p.scanMenus(ctx, query)
}

// GetMenusByAccountType retrieves menus filtered by account type.
// PREMIUM → all menus; REGULAR → only REGULAR type menus.
func (p *Provider) GetMenusByAccountType(ctx context.Context, accountType string) ([]Menu, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if accountType == "PREMIUM" {
		query := `SELECT id, "index", type, title, icon_url, is_active
			FROM menu WHERE is_active = TRUE ORDER BY "index" ASC`
		return p.scanMenus(ctx, query)
	}

	query := `SELECT id, "index", type, title, icon_url, is_active
		FROM menu WHERE type = $1 AND is_active = TRUE ORDER BY "index" ASC`
	return p.scanMenusWithParam(ctx, query, accountType)
}

func (p *Provider) scanMenus(ctx context.Context, query string) ([]Menu, error) {
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMenuRows(rows)
}

func (p *Provider) scanMenusWithParam(ctx context.Context, query string, param string) ([]Menu, error) {
	rows, err := p.DB.QueryContext(ctx, query, param)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMenuRows(rows)
}

func scanMenuRows(rows *sql.Rows) ([]Menu, error) {
	var menus []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(&m.ID, &m.Index, &m.Type, &m.Title, &m.IconURL, &m.IsActive); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if menus == nil {
		menus = []Menu{}
	}

	return menus, nil
}
