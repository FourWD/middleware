package common

import "github.com/jung-kurt/gofpdf"

func UploadPdfToGoogle(pdf *gofpdf.Fpdf, filename string, appID string, bucket string) (string, error) {
	path, err := SavePdf(pdf, filename, "./tmp/")
	if err != nil {
		return "", err
	}

	pathUpload, errUpload := UploadFileToGoogle(path, "auction", "fourwd-auction")
	if errUpload != nil {
		return "", err
	}
	return pathUpload, errUpload
}
