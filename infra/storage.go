package infra

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/jung-kurt/gofpdf"
	"google.golang.org/api/option"
)

type StorageClient struct {
	client *storage.Client
	bucket string
}

func NewStorageClient(ctx context.Context, cfg StorageConfig) (*StorageClient, error) {
	opts := []option.ClientOption{}
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	}
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}
	return &StorageClient{client: client, bucket: cfg.Bucket}, nil
}

func (c *StorageClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *StorageClient) UploadFile(ctx context.Context, localPath, remotePath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	objectPath := filepath.ToSlash(remotePath)
	writer := c.client.Bucket(c.bucket).Object(objectPath).NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		_ = writer.Close()
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", c.bucket, objectPath), nil
}

func (c *StorageClient) UploadPDF(ctx context.Context, pdf *gofpdf.Fpdf, filename, tempDir, remotePath string) (string, error) {
	localPath := filepath.Join(tempDir, filename)
	if err := pdf.OutputFileAndClose(localPath); err != nil {
		return "", fmt.Errorf("saving PDF: %w", err)
	}
	defer os.Remove(localPath)

	publicURL, err := c.UploadFile(ctx, localPath, remotePath)
	if err != nil {
		return "", fmt.Errorf("uploading PDF: %w", err)
	}
	return publicURL, nil
}
