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
	publicURL string // CDN URL prefix for serving files
}

func NewService(s3Client *s3.Client, bucket, publicURL string) *Service {
	return &Service{
		s3Client:  s3Client,
		bucket:    bucket,
		publicURL: publicURL,
	}
}

// UploadResult contains URLs for all generated variants of an uploaded photo.
type UploadResult struct {
	ID           string `json:"id"`
	URLThumbnail string `json:"url_thumbnail"`
	URLMedium    string `json:"url_medium"`
	URLLarge     string `json:"url_large"`
}

// GeneratePresignedUploadURL creates a presigned URL for direct client-to-R2 upload.
// This avoids routing large files through our server.
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

// UploadPhoto uploads a photo directly (for server-side processing).
func (s *Service) UploadPhoto(ctx context.Context, userID string, reader io.Reader, filename string) (*UploadResult, error) {
	fileID := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}

	// Upload original
	originalKey := fmt.Sprintf("photos/%s/%s-original%s", userID, fileID, ext)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(originalKey),
		Body:        reader,
		ContentType: aws.String(mimeType(ext)),
	})
	if err != nil {
		return nil, fmt.Errorf("uploading original: %w", err)
	}

	// TODO: Trigger async image processing pipeline:
	// 1. Generate thumbnail (150x150)
	// 2. Generate medium (400x600)
	// 3. Generate large (800x1200)
	// 4. Convert to WebP for 30% size reduction
	// 5. Run content moderation (Rekognition)
	// 6. Strip EXIF data (remove GPS, camera info)
	//
	// For MVP, we serve the original at all sizes.
	// Image processing can be added as a background job.

	result := &UploadResult{
		ID:           fileID,
		URLThumbnail: s.publicURL + "/" + originalKey,
		URLMedium:    s.publicURL + "/" + originalKey,
		URLLarge:     s.publicURL + "/" + originalKey,
	}

	return result, nil
}

// UploadVoiceNote uploads a voice note for audio bio or chat.
func (s *Service) UploadVoiceNote(ctx context.Context, userID string, reader io.Reader, durationSec int) (string, error) {
	fileID := uuid.New().String()
	key := fmt.Sprintf("audio/%s/%s.ogg", userID, fileID)

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("audio/ogg"),
		Metadata: map[string]string{
			"duration": fmt.Sprintf("%d", durationSec),
		},
	})
	if err != nil {
		return "", fmt.Errorf("uploading voice note: %w", err)
	}

	return s.publicURL + "/" + key, nil
}

// DeleteFile removes a file from R2.
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
	case ".gif":
		return "image/gif"
	case ".ogg":
		return "audio/ogg"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}
