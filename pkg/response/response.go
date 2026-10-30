package response

type BaseResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Berhasil"`
	Data    any    `json:"data"`
}

type SwaggerErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Terjadi kesalahan / Data tidak valid"`
	Data    any    `json:"data"`
}
