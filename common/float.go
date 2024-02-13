package common

import (
	"fmt"
	"strings"
)

func FloatWithCommas(value float64, digit int) string {
	if digit < 0 {
		digit = 0
	} else if digit > 5 {
		digit = 5
	}

	digitFormat := fmt.Sprintf("%%.%df", digit)
	format := fmt.Sprintf(digitFormat, value)
	parts := strings.Split(format, ".")
	integerPart := addCommas(parts[0])
	if len(parts) > 1 {
		decimalPart := parts[1]
		if len(decimalPart) < digit {
			decimalPart += strings.Repeat("0", digit-len(decimalPart))
		}
		return integerPart + "." + decimalPart
	}
	return integerPart
}

func addCommas(amount string) string {
	var withCommas string
	for i, c := range amount {
		if i > 0 && (len(amount)-i)%3 == 0 {
			withCommas += ","
		}
		withCommas += string(c)
	}
	return withCommas
}
