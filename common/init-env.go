package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func InitEnv() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "google.json")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error loading config.yaml")
		panic(err)
	}

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		panic(err)
	}
}
