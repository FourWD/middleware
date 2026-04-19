package common

import (
	"github.com/FourWD/middleware/infra"
	"github.com/FourWD/middleware/kit"
	"github.com/jung-kurt/gofpdf"
)

func UploadPdfToGoogle(pdf *gofpdf.Fpdf, filename string, appID string, bucket string) (string, error) {
	localPath := "tmp/"
	if infra.IsGAE() {
		localPath = "/tmp/"
	}

	path, err := kit.SavePdf(pdf, filename, localPath)
	if err != nil {
		return "", err
	}

	pathUpload, errUpload := UploadFileToGoogle(path, "auction", "fourwd-auction")
	if errUpload != nil {
		return "", err
	}
	return pathUpload, errUpload
}
