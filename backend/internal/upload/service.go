package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

type Service struct {
	client     *minio.Client
	bucket     string
	baseURL    string
	maxSizeMB  int64
}

func NewService(cfg *config.Config) (*Service, error) {
	client, err := minio.New(cfg.StorageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.StorageAccessKey, cfg.StorageSecretKey, ""),
		Secure: cfg.StorageUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	svc := &Service{
		client:    client,
		bucket:    cfg.StorageBucket,
		baseURL:   cfg.StorageBaseURL,
		maxSizeMB: cfg.UploadMaxSizeMB,
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.StorageBucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.StorageBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]
		}`, cfg.StorageBucket)
		_ = client.SetBucketPolicy(ctx, cfg.StorageBucket, policy)
	}

	return svc, nil
}

func (s *Service) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > s.maxSizeMB*1024*1024 {
		return "", fmt.Errorf("file too large (max %dMB)", s.maxSizeMB)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !isAllowedExt(ext) {
		return "", fmt.Errorf("file type not allowed")
	}

	objectName := fmt.Sprintf("uploads/%s%s", uuid.New().String(), ext)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, objectName), nil
}

func (s *Service) UploadFromURL(ctx context.Context, imageURL string) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := contentTypeToExt(contentType)
	objectName := fmt.Sprintf("uploads/%s%s", uuid.New().String(), ext)

	_, err = s.client.PutObject(ctx, s.bucket, objectName, resp.Body, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload from url: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, objectName), nil
}

func (s *Service) PresignedURL(ctx context.Context, filename string) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !isAllowedExt(ext) {
		return "", "", fmt.Errorf("file type not allowed")
	}

	objectName := fmt.Sprintf("uploads/%s%s", uuid.New().String(), ext)
	presignedURL, err := s.client.PresignedPutObject(ctx, s.bucket, objectName, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("presign: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s", s.baseURL, objectName)
	return presignedURL.String(), publicURL, nil
}

func isAllowedExt(ext string) bool {
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".svg": true,
	}
	return allowed[ext]
}

func contentTypeToExt(ct string) string {
	m := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	if ext, ok := m[ct]; ok {
		return ext
	}
	return ".jpg"
}

func (s *Service) ReadCloser(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
