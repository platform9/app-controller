package util

import (
	"fmt"
	"path/filepath"
	"regexp"
)

var Kubeconfig string

var (
	// Valid NameSpace.
	ValidNameSpaceRegex = fmt.Sprintf(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	NoSpecialChar       = "[^a-zA-Z0-9]+"

	// Maximum App Deploy Error
	MaxAppDeployError = "Maximum App deploy limit reached!"
	ErrorsToken       = []string{"Token is expired", "Forbidden", "Token Invalid"}
)

const (
	lowerCharSet = "abcdedfghijklmnopqrst"
	numberSet    = "0123456789"
	AllCharSet   = lowerCharSet + numberSet
	// Status Code for maximum app deployed limit.
	MaxAppDeployStatusCode = 429
	MaxAppDeployCount      = 7
	MaxAppScaleCount       = 1
)

// Validate a regex.
func RegexValidate(name string) bool {
	Regex := regexp.MustCompile(ValidNameSpaceRegex)
	return Regex.MatchString(name)
}

// Logger Variables.
var (
	// Logs location: /var/log/pf9/fast-path/fast-path.log
	logDir = "/var/log/"
	//Pf9Dir is the base Pf9Dir to store logs.
	Pf9Dir = filepath.Join(logDir, "pf9")
	//FastPathLogDir - Base Dir to store fast-path logs.
	FastPathLogDir = filepath.Join(Pf9Dir, "fast-path")
	//FastPathLog represents location of the log.
	FastPathLog = filepath.Join(FastPathLogDir, "fast-path.log")
)
