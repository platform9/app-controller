package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/platform9/fast-path/pkg/api"
	"github.com/platform9/fast-path/pkg/db"
	"github.com/platform9/fast-path/pkg/log"
	"github.com/platform9/fast-path/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Stub to read any environment variables
func readEnv() {
}

func run(*cobra.Command, []string) {
	zap.S().Info("Starting fast-path...")
	zap.S().Infof("Version of fast-path being used is: %s", util.Version)
	router := api.New()
	srv := &http.Server{
		Handler: router,
		Addr:    ":6112",
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			zap.S().Fatalf(err.Error())
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
		zap.S().Info("server stopping...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			zap.S().Fatalf(err.Error())
		}
	}
}

func main() {
	readEnv()
	cmd := buildCmds()
	cmd.Execute()
}

// Config file to read secrets like kubeconfig path, Database and auth0 credentials.
const (
	cfgFile = "/etc/pf9/fast-path/config.yaml"
)

func buildCmds() *cobra.Command {
	cobra.OnInitialize(initCfg)
	rootCmd := &cobra.Command{
		Use:   "fast-path",
		Short: "fast-path is a service to interact knative kubernetes clusters",
		Long:  "fast-path is a service to interact knative kubernetes clusters",
		Run:   run,
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate initializes and upgrades database",
		Long:  "migrate initializes and upgrades database",
		Run: func(cmd *cobra.Command, args []string) {
			dbHandle := db.Get()
			if err := dbHandle.Migrate(); err != nil {
				zap.S().Errorf(err.Error())
				panic(err)
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Current version of fast-path being used",
		Long:  "Current version of fast-path being used",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.Version)
		},
	}

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}

func initCfg() {
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		zap.S().Errorf(err.Error())
		panic(err)
	}
	util.Kubeconfig = viper.GetString("kubeconfig.file")
}

func init() {
	err := log.Logger()
	if err != nil {
		fmt.Printf("Failed to initiate logger, Error is: %s", err)
	}
}
