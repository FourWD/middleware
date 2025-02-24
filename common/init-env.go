package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/FourWD/middleware/model"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var App model.AppInfo
var AppLog = logrus.New()

func InitEnv() {
	if App.GaeService != "" {
		log.SetOutput(io.Discard)
	}

	AppLog.SetOutput(os.Stdout)
	AppLog.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
	})
	AppLog.SetLevel(logrus.InfoLevel)
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
