package sectors

import "github.com/google/uuid"

type Sector struct {
	ID         uuid.UUID `db:"id" json:"id"`
	SectorName string    `db:"sector_name" json:"sector_name"`
}
