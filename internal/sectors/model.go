package sectors

import "github.com/google/uuid"

type Sector struct {
	ID         uuid.UUID `db:"id" json:"id"`
	SectorName string    `db:"sector_name" json:"sector_name"`
}
type SwaggerSectorsResponse struct {
	Success bool     `json:"success" example:"true"`
	Message string   `json:"message" example:"Get data success"`
	Data    []Sector `json:"data"`
}
