package categories

import (
	"context"
	"database/sql"

	"moneyapp/backend/internal/platform/db"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) base(exec ...db.DBTX) db.DBTX {
	if len(exec) > 0 && exec[0] != nil {
		return exec[0]
	}

	return r.db
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, exec ...db.DBTX) ([]Category, error) {
	query := `
		select id, user_id, kind, name, color, icon, parent_id, is_system, is_archived, created_at, updated_at
		from finance_categories
		where user_id is null or user_id = $1
		order by is_system desc, name asc
	`
	rows, err := r.base(exec...).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Category
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, userID, categoryID uuid.UUID, exec ...db.DBTX) (Category, error) {
	query := `
		select id, user_id, kind, name, color, icon, parent_id, is_system, is_archived, created_at, updated_at
		from finance_categories
		where id = $1 and (user_id is null or user_id = $2)
	`
	var category Category
	err := r.base(exec...).QueryRowContext(ctx, query, categoryID, userID).Scan(
		&category.ID,
		&category.UserID,
		&category.Kind,
		&category.Name,
		&category.Color,
		&category.Icon,
		&category.ParentID,
		&category.IsSystem,
		&category.IsArchived,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	return category, err
}

func (r *Repository) Create(ctx context.Context, category Category, exec ...db.DBTX) error {
	query := `
		insert into finance_categories (
			id, user_id, kind, name, color, icon, parent_id, is_system, is_archived, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		category.ID,
		category.UserID,
		category.Kind,
		category.Name,
		category.Color,
		category.Icon,
		category.ParentID,
		category.IsSystem,
		category.IsArchived,
		category.CreatedAt,
		category.UpdatedAt,
	)
	return err
}

func (r *Repository) Update(ctx context.Context, category Category, exec ...db.DBTX) error {
	query := `
		update finance_categories
		set name = $3,
		    color = $4,
		    icon = $5,
		    parent_id = $6,
		    is_archived = $7,
		    updated_at = $8
		where id = $1 and user_id = $2
	`
	_, err := r.base(exec...).ExecContext(ctx, query,
		category.ID,
		category.UserID,
		category.Name,
		category.Color,
		category.Icon,
		category.ParentID,
		category.IsArchived,
		category.UpdatedAt,
	)
	return err
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(scanner categoryScanner) (Category, error) {
	var category Category
	err := scanner.Scan(
		&category.ID,
		&category.UserID,
		&category.Kind,
		&category.Name,
		&category.Color,
		&category.Icon,
		&category.ParentID,
		&category.IsSystem,
		&category.IsArchived,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	return category, err
}
