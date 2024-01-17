package common

import (
	"log"
	"os"
)

func Terminate() {
	zero := 0
	i := 1 / zero
	log.Panic(i)
	os.Exit(0)
}
