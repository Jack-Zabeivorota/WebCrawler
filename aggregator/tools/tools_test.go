package tools

import (
	"fmt"
	"strings"
	"testing"
)

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
