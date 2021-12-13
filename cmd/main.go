package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/platform9/fast-path/pkg/api"
	"github.com/platform9/fast-path/pkg/db"
	"github.com/platform9/fast-path/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Stub to read any environment variables
func readEnv() {
}


func run(*cobra.Command, []string) {
	log.Info("Starting fast-path...")
	router := api.New()
	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:6112",
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
		log.Infof("server stopping...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	readEnv()
	cmd := buildCmds()
	cmd.Execute()
}

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
				panic(err)
			}
		},
	}
	rootCmd.AddCommand(migrateCmd)

	return rootCmd
}

func initCfg() {
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	util.Kubeconfig = viper.GetString("kubeconfig.file")
}
