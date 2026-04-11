package kit

import (
	"github.com/jung-kurt/gofpdf"
)

func SavePdf(pdf *gofpdf.Fpdf, fileName string, location string) (string, error) {
	newFilename := fileName + "_" + DateStringPDF() + ".pdf"
	savePath := location + newFilename
	err := pdf.OutputFileAndClose(savePath)
	if err != nil {
		return "", err
	}

	return savePath, nil
}

func DateStringPDF() string {
	randomDigits := RandomString(5)
	return randomDigits
}
