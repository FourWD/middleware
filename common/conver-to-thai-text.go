package common

import (
	"fmt"
)

var ThaiNumeralMap = map[int]string{
	0: "ศูนย์",
	1: "หนึ่ง",
	2: "สอง",
	3: "สาม",
	4: "สี่",
	5: "ห้า",
	6: "หก",
	7: "เจ็ด",
	8: "แปด",
	9: "เก้า",
}

var ThaiPlaces = []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน", "ล้าน"}

func ConvertFloatToThaiText(number float64) string {
	// Split the number into integer and fractional parts
	integerPart := int(number)
	fractionalPart := int((number - float64(integerPart)) * 100)

	// Convert integer part to Thai text
	thaiText := ""
	if integerPart == 0 {
		thaiText = ThaiNumeralMap[0] // Special case for zero
	} else {
		for i := 0; integerPart > 0; i++ {
			digit := integerPart % 10
			place := i % 6 // Limit to "ล้าน" (up to one million)
			if digit > 0 {
				if place == 1 && digit == 1 {
					thaiText = "เอ็ด" + thaiText // Special case for one in "สิบ"
				} else if place == 0 && digit == 1 {
					thaiText = ThaiPlaces[place] + thaiText // Skip "หนึ่ง" for units
				} else {
					thaiText = ThaiNumeralMap[digit] + ThaiPlaces[place] + thaiText
				}
			}
			integerPart /= 10
		}
	}

	// Convert fractional part to Thai text
	if fractionalPart > 0 {
		thaiText += fmt.Sprintf("สตางค์%s", ConvertIntToThaiText(fractionalPart))
	}

	return thaiText
}

func ConvertIntToThaiText(number int) string {
	thaiText := ""
	for number > 0 {
		digit := number % 10
		thaiText = ThaiNumeralMap[digit] + thaiText
		number /= 10
	}
	return thaiText
}
