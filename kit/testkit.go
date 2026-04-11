package kit

import "os"

// SkippableTest is satisfied by *testing.T and *testing.B.
type SkippableTest interface {
	Skip(args ...any)
}

// MarkAsIntegrationTest skips the test unless TEST_INTEGRATION=true.
func MarkAsIntegrationTest(t SkippableTest) {
	if os.Getenv("TEST_INTEGRATION") != "true" {
		t.Skip("skipping integration tests: set TEST_INTEGRATION=true environment variable")
	}
}

// MarkAsE2ETest skips the test unless TEST_E2E=true.
func MarkAsE2ETest(t SkippableTest) {
	if os.Getenv("TEST_E2E") != "true" {
		t.Skip("skipping end to end tests: set TEST_E2E=true environment variable")
	}
}

// MarkAsFunctionalTest skips the test unless TEST_FUNCTIONAL=true.
func MarkAsFunctionalTest(t SkippableTest) {
	if os.Getenv("TEST_FUNCTIONAL") != "true" {
		t.Skip("skipping functional tests: set TEST_FUNCTIONAL=true environment variable")
	}
}
