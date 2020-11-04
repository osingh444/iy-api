package utils

import (
	"testing"
)

func TestBasicFilter(t *testing.T) {
	contains := "fuck bitches"
	doesntContain := "ayy lmao"

	if ContainsBadWords(doesntContain) {
		t.Error("filter error")
	}

	if !ContainsBadWords(contains) {
		t.Error("filter error")
	}
}
