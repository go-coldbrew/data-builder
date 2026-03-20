package databuilder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSetString(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{"empty set", nil, "[]"},
		{"single item", []string{"a"}, "[a]"},
		{"multiple items sorted", []string{"c", "a", "b"}, "[a b c]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStringSet(tt.items...)
			assert.Equal(t, tt.expected, s.String())
		})
	}
}

func TestStringSetStringInFmt(t *testing.T) {
	s := newStringSet("x", "y")
	result := fmt.Sprintf("missing fields %s", s)
	assert.Equal(t, "missing fields [x y]", result)
}
