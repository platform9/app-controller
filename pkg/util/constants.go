package util

import (
	"fmt"
	"regexp"
)

var Kubeconfig string

var ClientId string = "HEVMcEBvvQ1wnRmzOxlShZXvjp07bnMz"

var (
	// Valid NameSpace.
	ValidNameSpaceRegex = fmt.Sprintf(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	NoSpecialChar       = "[^a-zA-Z0-9]+"
)

const (
	lowerCharSet = "abcdedfghijklmnopqrst"
	numberSet    = "0123456789"
	AllCharSet   = lowerCharSet + numberSet
)

// Validate a regex.
func RegexValidate(name string) bool {
	Regex := regexp.MustCompile(ValidNameSpaceRegex)
	return Regex.MatchString(name)
}
