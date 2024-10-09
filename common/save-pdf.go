package common

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
	// currentTime := time.Now()
	// dateString := fmt.Sprintf("%d", currentTime.UnixNano())
	randomDigits := generateRandomString(5)
	return randomDigits
}

// var (
// 	counterMutex sync.Mutex
// 	callCounter  map[string]int
// )

// func init() {
// 	callCounter = make(map[string]int)
// }
// func SavePdf(pdf *gofpdf.Fpdf, fileName string, location string, auctionID string) (string, error) {
// 	counterMutex.Lock()
// 	callCounter[auctionID]++
// 	currentCall := callCounter[auctionID]
// 	counterMutex.Unlock()

// 	newFilename := fmt.Sprintf("%s_%s#%d.pdf", fileName, DateString(), currentCall)
// 	savePath := location + newFilename

// 	err := pdf.OutputFileAndClose(savePath)
// 	if err != nil {
// 		return "", err
// 	}

// 	return savePath, nil
// }
