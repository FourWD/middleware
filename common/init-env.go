package common

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/FourWD/middleware/model"
	"github.com/spf13/viper"
)

var App model.AppInfo

func InitEnv() {
	// os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "google.json")

	App.Env = os.Getenv("ENV")
	if App.Env == "" {
		App.Env = "local"
	}
	App.GaeService = os.Getenv("GAE_SERVICE")
	App.GaeVersion = os.Getenv("GAE_VERSION")

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
