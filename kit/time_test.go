package kit

import (
	"testing"
	"time"
)

// FreezeAt overrides kit.Now() to return a fixed time for the duration of the test.
func FreezeAt(t *testing.T, at time.Time) {
	nowFunc = func() time.Time {
		return at
	}

	t.Cleanup(func() {
		nowFunc = time.Now
	})
}
