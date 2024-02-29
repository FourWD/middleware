package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func InitEnv(name string) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "google.json")
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error loading %s.yaml", name)
		panic(err)
	}

	// err := godotenv.Load(".env")
	// if err != nil {
	// 	fmt.Println("Error loading .env file")
	// 	panic(err)
	// }
}
