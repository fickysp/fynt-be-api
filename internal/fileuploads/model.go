package fileuploads

type UploadResponseData struct {
	FileURL  string `json:"file_url"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

type SwaggerUploadResponse struct {
	Success bool               `json:"success" example:"true"`
	Message string             `json:"message" example:"File uploaded successfully"`
	Data    UploadResponseData `json:"data"`
}
