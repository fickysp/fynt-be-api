package storage

import (
	"bytes"
	"fmt"
	"mime/multipart"

	storage_go "github.com/supabase-community/storage-go"
)

type SupabaseStorage interface {
	UploadFile(file *multipart.FileHeader, folderName string) (string, error)
}

type supabaseStorage struct {
	client *storage_go.Client
	bucket string
	url    string
}

func NewSupabaseStorage(url, key, bucket string) SupabaseStorage {
	client := storage_go.NewClient(url, key, nil)
	return &supabaseStorage{
		client: client,
		bucket: bucket,
		url:    url,
	}
}

func (s *supabaseStorage) UploadFile(file *multipart.FileHeader, folderName string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("%s/%s", folderName, file.Filename)

	resp, err := s.client.UploadFile(s.bucket, filePath, buf)
	if resp.Error != "" {
		return "", fmt.Errorf("gagal mengunggah ke supabase: %s", resp.Error)
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucket, filePath)
	return publicURL, nil
}
