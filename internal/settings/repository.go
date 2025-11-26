package settings

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Get(ctx context.Context, key string) (*AppSetting, error)
	GetAll(ctx context.Context) ([]AppSetting, error)
	Update(ctx context.Context, key string, value string, updatedBy int64) error
	GetActiveElectionID(ctx context.Context) (int, error)
	GetDefaultElectionID(ctx context.Context) (int, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Get(ctx context.Context, key string) (*AppSetting, error) {
	query := `
		SELECT key, value, description, updated_at, updated_by
		FROM app_settings
		WHERE key = $1
	`
	
	var setting AppSetting
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
		&setting.UpdatedBy,
	)
	
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &setting, nil
}

func (r *repository) GetAll(ctx context.Context) ([]AppSetting, error) {
	query := `
		SELECT key, value, description, updated_at, updated_by
		FROM app_settings
		ORDER BY key
	`
	
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var settings []AppSetting
	for rows.Next() {
		var s AppSetting
		err := rows.Scan(&s.Key, &s.Value, &s.Description, &s.UpdatedAt, &s.UpdatedBy)
		if err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	
	return settings, nil
}

func (r *repository) Update(ctx context.Context, key string, value string, updatedBy int64) error {
	query := `
		UPDATE app_settings
		SET value = $1, updated_at = NOW(), updated_by = $2
		WHERE key = $3
	`
	
	_, err := r.pool.Exec(ctx, query, value, updatedBy, key)
	return err
}

func (r *repository) GetActiveElectionID(ctx context.Context) (int, error) {
	setting, err := r.Get(ctx, "active_election_id")
	if err != nil {
		return 0, err
	}
	
	if setting == nil {
		return 1, nil // Default
	}
	
	id, err := strconv.Atoi(setting.Value)
	if err != nil {
		return 1, nil // Default on error
	}
	
	return id, nil
}

func (r *repository) GetDefaultElectionID(ctx context.Context) (int, error) {
	setting, err := r.Get(ctx, "default_election_id")
	if err != nil {
		return 0, err
	}
	
	if setting == nil {
		return 1, nil // Default
	}
	
	id, err := strconv.Atoi(setting.Value)
	if err != nil {
		return 1, nil // Default on error
	}
	
	return id, nil
}
