package options

import (
	"fmt"

	"github.com/spf13/viper"
)


const (
	defaultDBType                = "mysql"
	defaultDBSrc                 = "file::memory:?cache=shared"
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
