package common

import "strings"

var thaiNumeralMap = map[int]string{
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

var thaiPlaces = []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน", "ล้าน"}

func ConvertFloatToThaiText(number float64) string {
	integerPart := int(number)
	fractionalPart := int((number - float64(integerPart)) * 100)

	thaiText := ""

	if integerPart == 0 {
		thaiText = thaiNumeralMap[0]
	} else {
		thaiText += convertIntToThaiText(integerPart)
	}
	thaiText += "บาท"

	if fractionalPart > 0 {
		thaiText += convertIntToThaiText(fractionalPart)
		thaiText += "สตางค์"
	}

	return thaiText
}

func convertIntToThaiText(number int) string {
	if number == 0 {
		return ""
	}

	thaiText := ""

	for i := 0; number > 0; i++ {
		digit := number % 10
		if digit > 0 {
			if digit == 1 && i%6 == 1 && number/10%10 == 0 {
				thaiText = "เอ็ด" + thaiText
			} else {
				thaiText = thaiNumeralMap[digit] + thaiPlaces[i%7] + thaiText
			}
		}
		number /= 10
	}

	thaiText = strings.Replace(thaiText, "หนึ่งสิบหนึ่ง", "สิบเอ็ด", -1)
	thaiText = strings.Replace(thaiText, "หนึ่งสิบ", "สิบ", -1)
	thaiText = strings.Replace(thaiText, "สองสิบ", "ยี่สิบ", -1)
	return thaiText
}
