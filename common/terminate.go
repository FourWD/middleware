package common

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func Terminate() {
	log.Println("Terminate App Version: ", viper.GetString("app_version"))
	zero := 0
	i := 1 / zero
	log.Panic(i)
	os.Exit(0)
}
