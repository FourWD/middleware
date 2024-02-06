package common

import (
	"github.com/jung-kurt/gofpdf"
)

func SavePdf(pdf *gofpdf.Fpdf, fileName string, location string) (string, error) {
	newFilename := DateString() + "_" + fileName
	savePath := location + newFilename
	err := pdf.OutputFileAndClose(savePath)
	if err != nil {
		return "", err
	}

	return savePath, nil
}
