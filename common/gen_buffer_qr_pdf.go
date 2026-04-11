package common

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"strings"

	"github.com/skip2/go-qrcode"
)

func GenBufferQrPdf(text string) (bytes.Buffer, error) {
	var buf bytes.Buffer

	qrImg, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		return buf, err
	}

	// Convert QR code image to base64 string
	base64Image := base64.StdEncoding.EncodeToString(qrImg)
	imageData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(base64Image))
	if err != nil {
		return buf, err
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return buf, err
	}

	// Convert the image to JPEG format (you may skip this step if the image is already JPEG)
	err = jpeg.Encode(&buf, img, nil)
	if err != nil {
		return buf, err
	}

	return buf, nil
}
