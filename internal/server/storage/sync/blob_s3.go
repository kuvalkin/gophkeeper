package sync

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

func NewS3BlobRepository(client *minio.Client, bucket string) *S3BlobRepository {
	return &S3BlobRepository{
		client: client,
		bucket: bucket,
	}
}

type S3BlobRepository struct {
	client *minio.Client
	bucket string
}

func (s *S3BlobRepository) CopyFrom(ctx context.Context, key string, r io.Reader) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, -1, minio.PutObjectOptions{})

	if err != nil {
		return fmt.Errorf("put object to s3: %w", err)
	}

	return nil
}
