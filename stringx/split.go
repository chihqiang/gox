package stringx

import (
	"github.com/samber/lo"
	"strings"
)

// Split splits a string by sep and trims each part, ignoring empty strings
func Split(s, sep string) []string {
	var sp []string
	for _, p := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			sp = append(sp, trimmed)
		}
	}
	return sp
}

// SplitUniq splits each string in ss by sep and flattens the result
func SplitUniq(ss []string, sep string) []string {
	var sp []string
	for _, s := range ss {
		sp = append(sp, Split(s, sep)...)
	}
	return lo.Uniq(sp)
}
