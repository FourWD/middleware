package kit

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/jung-kurt/gofpdf"
)

// --- Modern client-style API (preferred for new code) -----------------------

// GCSClient wraps storage.Client with bucket-scoped operations. Each instance
// is bucket-bound; create one per bucket if you work with multiple.
type GCSClient struct {
	client *storage.Client
	bucket string
}

// NewGCSClient creates a bucket-scoped GCS client. Caller owns the lifecycle
// — defer Close().
func NewGCSClient(ctx context.Context, bucket string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GCSClient{client: client, bucket: bucket}, nil
}

func (g *GCSClient) Close() error { return g.client.Close() }

// UploadFile uploads a local file to the given remote path (caller controls
// the path) and returns the public URL.
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

// UploadPDF saves the PDF to <tempDir>/<filename>, uploads to remotePath,
// then deletes the temp file.
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

// --- Legacy free-function API (signature-stable bridges) --------------------
//
// UploadFileToGoogle and DownloadFile keep the exact signatures used by
// pre-migration callers (common.UploadFileToGoogle, common.DownloadFile).
// They use a process-wide singleton client and apply conventional path
// defaults — convenience that costs flexibility. New code should use the
// GCSClient API above.

const gcsLegacyTimeout = 50 * time.Second

var legacyGCSClient = sync.OnceValues(func() (*storage.Client, error) {
	return storage.NewClient(context.Background())
})

// UploadFileToGoogle uploads filePath to GCS under "uploads/YYYY/MM/<basename>"
// and returns the public URL. The appID parameter is unused (kept for
// signature compatibility). Prefer GCSClient.UploadFile for new code.
func UploadFileToGoogle(filePath string, appID string, bucket string) (string, error) {
	client, err := legacyGCSClient()
	if err != nil {
		return "", fmt.Errorf("gcs client: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	now := time.Now()
	savePath := fmt.Sprintf("uploads/%04d/%02d/%s", now.Year(), int(now.Month()), filepath.Base(filePath))

	ctx, cancel := context.WithTimeout(context.Background(), gcsLegacyTimeout)
	defer cancel()

	wc := client.Bucket(bucket).Object(savePath).NewWriter(ctx)
	wc.ChunkSize = 0 // retries are not supported when ChunkSize is 0.

	if _, err := io.Copy(wc, file); err != nil {
		_ = wc.Close()
		return "", err
	}
	if err := wc.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, savePath), nil
}

// DownloadFile downloads a GCS object to destfilePath. The object name used
// is "vehicle/<basename(filePath)>" — historical naming. The appID parameter
// is unused. Prefer using GCSClient with explicit paths for new code.
func DownloadFile(filePath string, destfilePath string, appID string, bucket string) error {
	client, err := legacyGCSClient()
	if err != nil {
		return fmt.Errorf("gcs client: %w", err)
	}

	parts := strings.Split(filePath, "/")
	newFilename := parts[len(parts)-1]

	ctx, cancel := context.WithTimeout(context.Background(), gcsLegacyTimeout)
	defer cancel()

	f, err := os.Create(destfilePath)
	if err != nil {
		return err
	}

	rc, err := client.Bucket(bucket).Object("vehicle/" + newFilename).NewReader(ctx)
	if err != nil {
		_ = f.Close()
		return err
	}
	defer rc.Close()

	if _, err := io.Copy(f, rc); err != nil {
		_ = f.Close()
		return fmt.Errorf("io.Copy: %w", err)
	}
	return f.Close()
}
