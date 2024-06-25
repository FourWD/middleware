package common

import (
	"fmt"
	"sync"

	"github.com/jung-kurt/gofpdf"
)

// func SavePdf(pdf *gofpdf.Fpdf, fileName string, location string) (string, error) {
// 	newFilename := fileName + "_" + DateString() + "_" + ".pdf"
// 	savePath := location + newFilename
// 	err := pdf.OutputFileAndClose(savePath)
// 	if err != nil {
// 		return "", err
// 	}

//		return savePath, nil
//	}
var (
	counterMutex sync.Mutex
	callCounter  map[string]int
)

func init() {
	callCounter = make(map[string]int)
}
func SavePdf(pdf *gofpdf.Fpdf, fileName string, location string, auctionID string) (string, error) {
	counterMutex.Lock()
	callCounter[auctionID]++
	currentCall := callCounter[auctionID]
	counterMutex.Unlock()

	newFilename := fmt.Sprintf("%s_%s#%d.pdf", fileName, DateString(), currentCall)
	savePath := location + newFilename

	err := pdf.OutputFileAndClose(savePath)
	if err != nil {
		return "", err
	}

	return savePath, nil
}
