package common

import (
	"runtime"
	"time"

	"github.com/patrickmn/go-cache"
)

func ClearCache() {
	var c = cache.New(5*time.Minute, 10*time.Minute)
	c.Flush()
}

func ClearMemory() {
	runtime.GC()
}
