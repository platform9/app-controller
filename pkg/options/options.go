package options

import (
	"fmt"
	"strconv"

	"github.com/platform9/fast-path/pkg/util"
	"github.com/spf13/viper"
)

const (
	defaultDBType = "mysql"
	defaultDBSrc  = "file::memory:?cache=shared"
)

func init() {
	viper.SetDefault("db.type", defaultDBType)
	viper.SetDefault("db.src", defaultDBSrc)
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

func GetConstraintMaxScale() (int, error) {
	max_scale_str := viper.GetString("constraints.max-scale")
	max_scale, err := strconv.Atoi(max_scale_str)
	if err != nil {
		return util.MaxAppScaleCount, err
	}
	return max_scale, nil
}

func GetConstraintMaxAppDeploy() (int, error) {
	max_app_str := viper.GetString("constraints.max-app")
	fmt.Printf("Max app str is %v", max_app_str)
	max_app, err := strconv.Atoi(max_app_str)
	if err != nil {
		return util.MaxAppDeployCount, err
	}
	return max_app, nil
}
