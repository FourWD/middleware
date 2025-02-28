package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var App model.AppInfo
var AppLog *zap.Logger

func InitEnv() {
	if App.GaeService != "" {
		log.SetOutput(io.Discard)
	}

	initLog()
	// os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "google.json")
	App.Env = os.Getenv("ENV")
	if App.Env == "" {
		App.Env = "local"
	}

	App.GaeProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	App.GaeService = os.Getenv("GAE_SERVICE")
	App.GaeVersion = os.Getenv("GAE_VERSION")
	App.BucketName = os.Getenv("BUCKET")

	viper.SetConfigName(fmt.Sprintf("config.%s", App.Env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error loading config")
		panic(err)
	}

	log.Println("ENV = ", App.Env)

	// err := godotenv.Load(".env")
	// if err != nil {
	// 	fmt.Println("Error loading .env file")
	// 	panic(err)
	// }
}

func initLog() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	config.DisableStacktrace = true
	config.DisableCaller = true
	AppLog, _ = config.Build()
	defer AppLog.Sync()
}
