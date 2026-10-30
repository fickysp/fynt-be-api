package storage

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"time"

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
	if url == "" || key == "" || bucket == "" {
		slog.Error("konfigurasi supabase tidak lengkap", "url_empty", url == "", "key_empty", key == "", "bucket_empty", bucket == "")
	}

	storageURL := fmt.Sprintf("%s/storage/v1", url)
	client := storage_go.NewClient(storageURL, key, nil)

	slog.Info("supabase storage diinisialisasi", "storage_url", storageURL, "bucket", bucket)

	return &supabaseStorage{
		client: client,
		bucket: bucket,
		url:    url,
	}
}

func (s *supabaseStorage) UploadFile(file *multipart.FileHeader, folderName string) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("terjadi masalah pada client supabase")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file header: %w", err)
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return "", fmt.Errorf("gagal membaca stream file ke buffer: %w", err)
	}

	ext := filepath.Ext(file.Filename)
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := fmt.Sprintf("%s/%s", folderName, uniqueName)

	contentType := file.Header.Get("Content-Type")

	resp, err := s.client.UploadFile(s.bucket, filePath, buf, storage_go.FileOptions{
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("gagal terhubung ke supabase storage: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("gagal mengunggah ke supabase: %s", resp.Error)
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucket, filePath)
	slog.Info("file berhasil diunggah", "path", filePath, "url", publicURL)

	return publicURL, nil
}
