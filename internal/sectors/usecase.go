package sectors

import (
	"context"
	"log/slog"
)

type sectorsUsecase struct {
	sectorsRepo SectorRepository
}

type SectorsUsecase interface {
	GetAllSectors(ctx context.Context) ([]Sector, error)
}

func NewSectorsUsecase(sr SectorRepository) SectorsUsecase {
	return &sectorsUsecase{
		sectorsRepo: sr,
	}
}

func (u *sectorsUsecase) GetAllSectors(ctx context.Context) ([]Sector, error) {
	sector, err := u.sectorsRepo.GetAllSectors(ctx)
	if err != nil {
		slog.Error("failed to get all sectors", "error", err)
		return nil, err
	}
	return sector, nil
}
