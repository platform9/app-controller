package options

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

const (
	defaultDBType     = "mysql"
	defaultDBSrc      = "file::memory:?cache=shared"
	maxAppScaleCount  = 1
	maxAppDeployCount = 7
)

func init() {
	viper.SetDefault("db.type", defaultDBType)
	viper.SetDefault("db.src", defaultDBSrc)
	viper.SetDefault("constraints.max-scale", maxAppScaleCount)
	viper.SetDefault("constraints.max-app", maxAppDeployCount)
}

// GetDBType returns database type
func GetDBType() string {
	return viper.GetString("db.type")
}

// GetDBSrc returns database source string
func GetDBSrc() string {
	return viper.GetString("db.src")
}

// GetDefaultDBSrc returns database source string
func GetDefaultDBSrc() string {
	return defaultDBSrc
}

// GetDBCreds returns MySQL db string
func GetDBCreds() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		viper.GetString("db.user"),
		viper.GetString("db.password"),
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.name"))
}

// GetConstraintMaxScale returns the maximum app scale count.
func GetConstraintMaxScale() int {
	max_scale_str := viper.GetString("constraints.max-scale")
	max_scale, err := strconv.Atoi(max_scale_str)
	if err != nil {
		return maxAppScaleCount
	}
	return max_scale
}

// GetAuth0ClientId returns the auth0 client-id.
func GetAuth0ClientId() string {
	return viper.GetString("auth0.client-id")
}

// GetConstraintMaxAppDeploy returns the maximum apps allowed to deploy.
func GetConstraintMaxAppDeploy() int {
	max_app_str := viper.GetString("constraints.max-app")
	max_app, err := strconv.Atoi(max_app_str)
	if err != nil {
		return maxAppDeployCount
	}
	return max_app
}

// GetJWKSURL returns the JWKS-URL for validation.
func GetJWKSURL() string {
	return viper.GetString("jwks.url")
}
