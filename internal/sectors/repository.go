package sectors

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type SectorRepository interface {
	Create(ctx context.Context, secName string) (string, error)
	GetSectorByName(ctx context.Context, name string) (*Sector, error)
	GetAllSectors(ctx context.Context) ([]Sector, error)
}

type sectorRepository struct {
	db *sqlx.DB
}

func NewSectorRepository(db *sqlx.DB) SectorRepository {
	return &sectorRepository{
		db: db,
	}
}

func (r *sectorRepository) Create(ctx context.Context, secName string) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	query := `INSERT INTO sectors (sector_name) 
	VALUES ($1)
	RETURNING id`

	var sectorID string
	err = tx.QueryRowContext(ctx, query, secName).Scan(&sectorID)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}

	return sectorID, nil
}

func (r *sectorRepository) GetSectorByName(ctx context.Context, name string) (*Sector, error) {
	query := `SELECT * FROM sectors WHERE sector_name = $1`
	var sector Sector

	if err := r.db.GetContext(ctx, &sector, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sector, nil

}

func (r *sectorRepository) GetAllSectors(ctx context.Context) ([]Sector, error) {
	query := `SELECT * FROM sectors ORDER BY sector_name ASC`
	var sector []Sector

	if err := r.db.SelectContext(ctx, &sector, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return sector, nil

}
