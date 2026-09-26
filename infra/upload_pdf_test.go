package infra

import (
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func TestUploadPdfToGoogle_RequiresBucket(t *testing.T) {
	for _, bucket := range []string{"", "   "} {
		if _, err := UploadPdfToGoogle(gofpdf.New("P", "mm", "A4", ""), "x", "app", bucket); err == nil {
			t.Errorf("bucket %q: expected error, got nil", bucket)
		}
	}
}
