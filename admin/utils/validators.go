package utils

import (
	"regexp"
	"unicode"
)

func ContainsCapitalLetter(password string) bool {
	for _, ch := range password {
		if unicode.IsUpper(ch) {
			return true
		}
	}
	return false
}

func ContainsSymbol(regex *regexp.Regexp, password string) bool {
	return regex.MatchString(password)
}
