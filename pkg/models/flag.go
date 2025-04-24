package models

import (
	"strings"
)

// StringFlags allows the application to take multiple arguments with the
// same name and combine them into a slice.
type StringFlags []string

func (i *StringFlags) String() string {
	s := []string(*i)
	return strings.Join(s, ", ")
}

func (i *StringFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
