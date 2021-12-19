package log

import (
	"fmt"
	"os"
	"time"

	"github.com/platform9/fast-path/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func createDirectoryIfNotExists() error {
	var err error
	// Create Pf9Dir
	if _, err = os.Stat(util.Pf9Dir); os.IsNotExist(err) {
		errdir := os.Mkdir(util.Pf9Dir, os.ModePerm)
		if errdir != nil {
			return errdir
		}
	}
	// Create FastPathLogDir.
	if _, err = os.Stat(util.FastPathLogDir); os.IsNotExist(err) {
		errlogdir := os.Mkdir(util.FastPathLogDir, os.ModePerm)
		if errlogdir != nil {
			return errlogdir
		}
		return nil
	}
	return err
}

func fileConfig() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.9999Z"))
	})
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

func Logger() error {
	//Create the Pf9Dir, FastPathLogDir directory to store logs.
	err := createDirectoryIfNotExists()
	if err != nil {
		return fmt.Errorf("Failed to create Director. \nError is: %s", err)
	}

	// Open/Create the fast-path.log file.
	file, err := os.OpenFile(util.FastPathLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Couldn't open the log file: %s. \nError is: %s", util.FastPathLog, err)
	}

	core := zapcore.NewCore(fileConfig(), zapcore.AddSync(file), zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller())
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	return nil
}
