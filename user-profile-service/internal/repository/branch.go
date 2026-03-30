package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/bankease/user-profile-service/internal/models"
)

// BranchRepository handles database operations for branches.
type BranchRepository struct {
	DB *sql.DB
}

// GetAll retrieves all branches.
func (r *BranchRepository) GetAll(ctx context.Context) ([]models.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, distance, latitude, longitude FROM branch`
	return r.scanBranches(ctx, query)
}

// SearchByName retrieves branches whose name matches the query (case-insensitive partial match).
func (r *BranchRepository) SearchByName(ctx context.Context, q string) ([]models.Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, distance, latitude, longitude
		FROM branch WHERE name ILIKE '%' || $1 || '%'`
	return r.scanBranchesWithParam(ctx, query, q)
}

func (r *BranchRepository) scanBranches(ctx context.Context, query string) ([]models.Branch, error) {
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBranchRows(rows)
}

func (r *BranchRepository) scanBranchesWithParam(ctx context.Context, query string, param string) ([]models.Branch, error) {
	rows, err := r.DB.QueryContext(ctx, query, param)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBranchRows(rows)
}

func scanBranchRows(rows *sql.Rows) ([]models.Branch, error) {
	var items []models.Branch
	for rows.Next() {
		var b models.Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Distance, &b.Latitude, &b.Longitude); err != nil {
			return nil, err
		}
		items = append(items, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []models.Branch{}
	}

	return items, nil
}
