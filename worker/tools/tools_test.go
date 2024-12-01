package tools

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetDomain(t *testing.T) {
	expected := "http://example.com"

	result := GetDomain("http://example.com")

	if result != expected {
		t.Errorf("http://example.com -> %s, expected %s", result, expected)
	}

	result = GetDomain("http://example.com/some/page")

	if result != expected {
		t.Errorf("http://example.com/some/page -> %s, expected %s", result, expected)
	}
}

func TestSelect(t *testing.T) {
	result := strings.Join(Select([]int{1, 2, 3}, func(i int) string {
		return fmt.Sprintf("num %d", i)
	}), ",")
	expected := "num 1,num 2,num 3"

	if result != expected {
		t.Errorf("1,2,3 -> %s, expected: %s", result, expected)
	}

	result = strings.Join(Select([]int{}, func(i int) string {
		return fmt.Sprintf("num %d", i)
	}), ",")
	expected = ""

	if result != expected {
		t.Errorf("none -> %s, expected: none", result)
	}
}
