package common

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
	integerPart := int(number)
	fractionalPart := int((number - float64(integerPart)) * 100)

	thaiText := ""

	thaiText += convertIntToThaiText(integerPart)
	if fractionalPart > 0 {
		thaiText += "บาท"
	} else {
		thaiText += "บาทถ้วน"
	}

	if fractionalPart > 0 {
		thaiText += convertIntToThaiText(fractionalPart)
		thaiText += "สตางค์"
	}

	return thaiText
}

func convertIntToThaiText(number int) string {
	thaiText := ""

	for number > 0 {
		digit := number % 10
		if digit > 0 {
			if string(thaiText[0]) == "ห" && digit == 1 && len(thaiText) > 0 {
				thaiText = "เอ็ด" + thaiText
			} else if digit == 1 && len(thaiText) == 0 && number/10%10 == 0 {
				thaiText = "หนึ่ง"
			} else {
				thaiText = ThaiNumeralMap[digit] + thaiText
			}
		}
		number /= 10
	}

	return thaiText
}
