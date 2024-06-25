package common

import "github.com/jung-kurt/gofpdf"

func UploadPdfToGoogle(pdf *gofpdf.Fpdf, filename string, appID string, bucket string, auctionID string) (string, error) {
	localPath := "tmp/"
	if App.GaeService != "" {
		localPath = "/tmp/"
	}

	path, err := SavePdf(pdf, filename, localPath, auctionID)
	if err != nil {
		return "", err
	}

	pathUpload, errUpload := UploadFileToGoogle(path, "auction", "fourwd-auction")
	if errUpload != nil {
		return "", err
	}
	return pathUpload, errUpload
}
