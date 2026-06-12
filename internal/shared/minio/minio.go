package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/alberdjuniawan/votesystem/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

func NewClient(cfg config.MinIOConfig) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[{
				"Effect":"Allow",
				"Principal":{"AWS":["*"]},
				"Action":["s3:GetObject"],
				"Resource":["arn:aws:s3:::%s/*"]
			}]
		}`, cfg.Bucket)

		if err := mc.SetBucketPolicy(ctx, cfg.Bucket, policy); err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return &Client{
		client:   mc,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
		useSSL:   cfg.UseSSL,
	}, nil
}

type UploadInput struct {
	ObjectName  string
	Reader      io.Reader
	Size        int64
	ContentType string
}

func (c *Client) Upload(ctx context.Context, input UploadInput) error {
	_, err := c.client.PutObject(ctx, c.bucket, input.ObjectName, input.Reader, input.Size, minio.PutObjectOptions{
		ContentType: input.ContentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}

func (c *Client) GetPublicURL(objectName string) string {
	scheme := "http"
	if c.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, c.endpoint, c.bucket, objectName)
}
