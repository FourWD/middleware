package common

import (
	"github.com/jung-kurt/gofpdf"
)

func SavePdf(pdf *gofpdf.Fpdf, fileName string, location string) (string, error) {
	newFilename := fileName + "_" + RandomString(5) + "_" + ".pdf"
	savePath := location + newFilename
	err := pdf.OutputFileAndClose(savePath)
	if err != nil {
		return "", err
	}

	return savePath, nil
}
