package common

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

func Terminate() {
	fmt.Printf("Terminate App Version: %s [%s]\n", viper.GetString("app_version"), os.Getenv("GAE_VERSION"))
	zero := 0
	i := 1 / zero
	log.Panic(i)
	os.Exit(0)
}
