package common

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var ENV string

func InitEnv() {
	// os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "google.json")
	ENV = os.Getenv("ENV")
	if ENV == "" {
		ENV = "local"
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", ENV))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error loading config")
		panic(err)
	}

	log.Println("ENV = ", ENV)

	// err := godotenv.Load(".env")
	// if err != nil {
	// 	fmt.Println("Error loading .env file")
	// 	panic(err)
	// }
}
