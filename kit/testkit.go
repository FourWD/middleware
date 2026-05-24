package kit

import (
	"fmt"
	"os"
)

// SkippableTest is satisfied by *testing.T and *testing.B.
type SkippableTest interface {
	Skip(args ...any)
}

func MarkAsIntegrationTest(t SkippableTest) {
	skipUnlessEnv(t, "TEST_INTEGRATION", "integration")
}

func MarkAsE2ETest(t SkippableTest) {
	skipUnlessEnv(t, "TEST_E2E", "end to end")
}

func MarkAsFunctionalTest(t SkippableTest) {
	skipUnlessEnv(t, "TEST_FUNCTIONAL", "functional")
}

func skipUnlessEnv(t SkippableTest, envVar, label string) {
	if os.Getenv(envVar) != "true" {
		t.Skip(fmt.Sprintf("skipping %s tests: set %s=true environment variable", label, envVar))
	}
}
