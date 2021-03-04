package utils

import (
	"strings"
)

var badWords = [...]string {
	
}

func ContainsBadWords(content string) bool {
	for _, word := range badWords {
		if strings.Contains(content, word) {
			return true
		}
	}
	return false
}
