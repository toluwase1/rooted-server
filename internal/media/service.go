package media

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Service struct {
	s3Client  *s3.Client
	bucket    string
	publicURL string
}

func NewService(s3Client *s3.Client, bucket, publicURL string) *Service {
	return &Service{s3Client: s3Client, bucket: bucket, publicURL: publicURL}
}

type UploadResult struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	URLThumbnail string `json:"url_thumbnail"`
	URLMedium    string `json:"url_medium"`
	URLLarge     string `json:"url_large"`
}

func (s *Service) GeneratePresignedUploadURL(ctx context.Context, userID, fileExt string) (string, string, error) {
	fileID := uuid.New().String()
	key := fmt.Sprintf("uploads/%s/%s%s", userID, fileID, fileExt)

	presignClient := s3.NewPresignClient(s.s3Client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", "", fmt.Errorf("generating presigned URL: %w", err)
	}

	return req.URL, fileID, nil
}

func (s *Service) UploadPhoto(ctx context.Context, userID string, reader io.Reader, filename string) (*UploadResult, error) {
	fileID := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}

	key := fmt.Sprintf("photos/%s/%s%s", userID, fileID, ext)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(mimeType(ext)),
	})
	if err != nil {
		return nil, fmt.Errorf("uploading photo: %w", err)
	}

	// Generate presigned read URL (valid 24 hours)
	readURL, err := s.GetPresignedReadURL(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("generating read URL: %w", err)
	}

	return &UploadResult{
		ID:           fileID,
		Key:          key,
		URLThumbnail: readURL,
		URLMedium:    readURL,
		URLLarge:     readURL,
	}, nil
}

func (s *Service) UploadVoiceNote(ctx context.Context, userID string, reader io.Reader, durationSec int) (string, error) {
	fileID := uuid.New().String()
	key := fmt.Sprintf("audio/%s/%s.ogg", userID, fileID)

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("audio/ogg"),
	})
	if err != nil {
		return "", fmt.Errorf("uploading voice note: %w", err)
	}

	url, err := s.GetPresignedReadURL(ctx, key)
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetPresignedReadURL generates a temporary read URL for a stored file (24h expiry).
func (s *Service) GetPresignedReadURL(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.s3Client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("generating presigned read URL: %w", err)
	}
	return req.URL, nil
}

func (s *Service) DeleteFile(ctx context.Context, key string) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func mimeType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
