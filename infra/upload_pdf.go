package infra

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/FourWD/middleware/kit"
	"github.com/jung-kurt/gofpdf"
)

// UploadPdfToGoogle saves a PDF to the OS temp directory, uploads it to the
// given GCS bucket, then removes the local copy.
func UploadPdfToGoogle(pdf *gofpdf.Fpdf, filename string, appID string, bucket string) (string, error) {
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return "", errors.New("upload pdf: bucket is required")
	}

	// Strip directory components so a caller-supplied name cannot escape the temp dir.
	filename = filepath.Base(filename)

	path, err := kit.SavePdf(pdf, filename, os.TempDir()+string(os.PathSeparator))
	if err != nil {
		return "", err
	}
	// GAE /tmp is memory-backed; leftover files count against instance RAM.
	defer os.Remove(path)

	return kit.UploadFileToGoogle(path, appID, bucket)
}
