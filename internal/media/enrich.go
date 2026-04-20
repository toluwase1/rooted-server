package media

import (
	"context"
	"strings"
)

// EnrichPhotoURL converts an R2 key to a presigned read URL.
// Returns the original string if it's already a URL or if the service is nil.
func (s *Service) EnrichPhotoURL(ctx context.Context, key string) string {
	if s == nil || key == "" || strings.HasPrefix(key, "http") {
		return key
	}
	url, err := s.GetPresignedReadURL(ctx, key)
	if err != nil {
		return key
	}
	return url
}
