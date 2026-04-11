package kit

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/jung-kurt/gofpdf"
)

type GCSClient struct {
	client *storage.Client
	bucket string
}

func NewGCSClient(ctx context.Context, bucket string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCSClient{
		client: client,
		bucket: bucket,
	}, nil
}

func (g *GCSClient) Close() error {
	return g.client.Close()
}

func (g *GCSClient) UploadFile(ctx context.Context, localPath, remotePath string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	objectPath := filepath.ToSlash(remotePath)
	writer := g.client.Bucket(g.bucket).Object(objectPath).NewWriter(ctx)

	if _, err := io.Copy(writer, file); err != nil {
		_ = writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucket, objectPath), nil
}

func (g *GCSClient) UploadPDF(ctx context.Context, pdf *gofpdf.Fpdf, filename, tempDir, remotePath string) (string, error) {
	localPath := filepath.Join(tempDir, filename)
	if err := pdf.OutputFileAndClose(localPath); err != nil {
		return "", fmt.Errorf("saving PDF: %w", err)
	}
	defer os.Remove(localPath)

	publicURL, err := g.UploadFile(ctx, localPath, remotePath)
	if err != nil {
		return "", fmt.Errorf("uploading PDF: %w", err)
	}

	return publicURL, nil
}
