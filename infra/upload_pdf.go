package infra

import (
	"github.com/FourWD/middleware/kit"
	"github.com/jung-kurt/gofpdf"
)

// UploadPdfToGoogle saves a PDF to a local tmp directory then uploads it to
// the "fourwd-auction" GCS bucket. Local path is "tmp/" or "/tmp/" depending
// on whether the process runs on App Engine.
func UploadPdfToGoogle(pdf *gofpdf.Fpdf, filename string, appID string, bucket string) (string, error) {
	localPath := "tmp/"
	if IsGAE() {
		localPath = "/tmp/"
	}

	path, err := kit.SavePdf(pdf, filename, localPath)
	if err != nil {
		return "", err
	}

	uploaded, err := kit.UploadFileToGoogle(path, "auction", "fourwd-auction")
	if err != nil {
		return "", err
	}
	return uploaded, nil
}
