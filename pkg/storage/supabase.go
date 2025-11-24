package storage

import (
"bytes"
"context"
"fmt"
"io"
"os"
"time"

storage_go "github.com/supabase-community/storage-go"
)

type SupabaseStorage struct {
client *storage_go.Client
url    string
}

func NewSupabaseStorage() (*SupabaseStorage, error) {
url := os.Getenv("SUPABASE_URL")
key := os.Getenv("SUPABASE_SECRET_KEY")

if url == "" || key == "" {
return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SECRET_KEY must be set")
}

client := storage_go.NewClient(url+"/storage/v1", key, nil)

return &SupabaseStorage{
client: client,
url:    url,
}, nil
}

// Upload uploads a file to Supabase Storage and returns public URL
func (s *SupabaseStorage) Upload(ctx context.Context, bucket, path string, data []byte, contentType string) (string, error) {
reader := bytes.NewReader(data)

_, err := s.client.UploadFile(bucket, path, reader)
if err != nil {
return "", fmt.Errorf("failed to upload to supabase: %w", err)
}

publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, bucket, path)
return publicURL, nil
}

// Delete deletes a file from Supabase Storage
func (s *SupabaseStorage) Delete(ctx context.Context, bucket, path string) error {
_, err := s.client.RemoveFile(bucket, []string{path})
if err != nil {
return fmt.Errorf("failed to delete from supabase: %w", err)
}
return nil
}

// GenerateMediaPath generates a unique path for media files
func GenerateMediaPath(candidateID int64, slot, extension string) string {
timestamp := time.Now().Unix()
return fmt.Sprintf("candidates/%d/%s_%d%s", candidateID, slot, timestamp, extension)
}

// GetExtension gets file extension from content type
func GetExtension(contentType string) string {
switch contentType {
case "image/jpeg", "image/jpg":
return ".jpg"
case "image/png":
return ".png"
case "image/webp":
return ".webp"
case "application/pdf":
return ".pdf"
default:
return ".bin"
}
}

// GetMediaBucket returns the bucket name for media
func GetMediaBucket() string {
bucket := os.Getenv("SUPABASE_MEDIA_BUCKET")
if bucket == "" {
return "pemira"
}
return bucket
}
