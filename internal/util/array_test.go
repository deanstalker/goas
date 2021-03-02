package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInStringList(t *testing.T) {
	tests := map[string]struct {
		list   []string
		search string
		want   bool
	}{
		"no match": {
			[]string{"ant", "aphid", "grasshopper"},
			"beetle",
			false,
		},
		"match": {
			[]string{"ant", "aphid", "grasshopper"},
			"ant",
			true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsInStringList(tc.list, tc.search))
		})
	}
}
