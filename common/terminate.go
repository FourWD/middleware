package common

import (
	"fmt"
	"log"
	"os"
)

func Terminate() {
	//if App.Env != "local" {
	fmt.Printf("Terminate %s [%s] Version: [%s - %s]\n", App.GaeService, App.Env, App.AppVersion, App.GaeVersion)
	zero := 0
	i := 1 / zero
	log.Panic(i)
	os.Exit(0)
	//}
}
