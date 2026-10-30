package sectors

import (
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/gofiber/fiber/v3"
)

type SectorHandler struct {
	sectorsUsecase SectorsUsecase
}

func NewSectorsHandler(sU SectorsUsecase) *SectorHandler {
	return &SectorHandler{
		sectorsUsecase: sU,
	}
}

// @Summary Get All Sectors
// @Description Mengambil seluruh sektor pekerjaan
// @Tags Sectors
// @Produce json
// @Success 200 {object} SwaggerSectorsResponse "Sukses mengambil data"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/sectors [get]
func (h *SectorHandler) GetAllSectors(c fiber.Ctx) error {
	sectors, err := h.sectorsUsecase.GetAllSectors(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerSectorsResponse{Success: true, Message: "Sukses mengambil data", Data: sectors})
}
