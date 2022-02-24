package util

import (
	"fmt"
	"path/filepath"
	"regexp"
)

var (
	//Kubeconfig path.
	Kubeconfig string

	//Valid NameSpace.
	ValidNameSpaceRegex = fmt.Sprintf(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	NoSpecialChar       = "[^a-zA-Z0-9]+"

	//Maximum App Deploy Error
	MaxAppDeployError = "Maximum App deploy limit reached!"
	ErrorsToken       = []string{"Token is expired", "Forbidden", "Token Invalid"}
	Errors            = []string{"Failed to parse image"}
)

//Logger Variables.
var (
	//Logs location: /var/log/pf9/fast-path/fast-path.log
	logDir = "/var/log/"
	//Pf9Dir is the base Pf9Dir to store logs.
	Pf9Dir = filepath.Join(logDir, "pf9")
	//FastPathLogDir - Base Dir to store fast-path logs.
	FastPathLogDir = filepath.Join(Pf9Dir, "fast-path")
	//FastPathLog represents location of the log.
	FastPathLog = filepath.Join(FastPathLogDir, "fast-path.log")
)

const (
	//To create random code.
	lowerCharSet = "abcdedfghijklmnopqrst"
	numberSet    = "0123456789"
	AllCharSet   = lowerCharSet + numberSet

	//Status Code for maximum app deployed limit.
	MaxAppDeployStatusCode = 429

	// Secret URL constants
	HTTPURL         = "http://"
	HTTPSURL        = "https://"
	DockerURL       = "docker.io"
	DockerServerURL = "https://index.docker.io/v1/"
	AWSURL          = "amazonaws"
	GCRURL          = "gcr.io"

	// fast-path version
	Version = "fast-path version: v1.0"
)

//Validate a regex.
func RegexValidate(name string) bool {
	Regex := regexp.MustCompile(ValidNameSpaceRegex)
	return Regex.MatchString(name)
}
