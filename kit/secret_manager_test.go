package kit

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

// stubSecretFetcher makes ResolveSecretEnv resolve against an in-memory map for
// the duration of the test.
func stubSecretFetcher(t *testing.T, payloads map[string]string, fetchErr error) *[]string {
	t.Helper()
	var seen []string
	newSecretFetcher = func(context.Context) (secretFetcher, func(), error) {
		fetch := func(_ context.Context, name string) (string, error) {
			seen = append(seen, name)
			if fetchErr != nil {
				return "", fetchErr
			}
			payload, ok := payloads[name]
			if !ok {
				return "", errors.New("not found")
			}
			return payload, nil
		}
		return fetch, func() {}, nil
	}
	t.Cleanup(func() { newSecretFetcher = newGoogleSecretFetcher })
	return &seen
}

func TestSecretVersionName(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		project string
		want    string
		wantErr bool
	}{
		{"latest by default", "db-password", "proj", "projects/proj/secrets/db-password/versions/latest", false},
		{"pinned version", "db-password#5", "proj", "projects/proj/secrets/db-password/versions/5", false},
		{"empty version falls back to latest", "db-password#", "proj", "projects/proj/secrets/db-password/versions/latest", false},
		{"full resource name passes through", "projects/other/secrets/n/versions/3", "", "projects/other/secrets/n/versions/3", false},
		{"empty name", "", "proj", "", true},
		{"no project", "db-password", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := secretVersionName(tt.ref, tt.project)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveSecretEnvNoRefs(t *testing.T) {
	t.Setenv("PLAIN_VALUE", "literal")
	seen := stubSecretFetcher(t, nil, nil)

	if err := ResolveSecretEnv(context.Background()); err != nil {
		t.Fatalf("ResolveSecretEnv: %v", err)
	}
	if len(*seen) != 0 {
		t.Errorf("fetched %v, want no Secret Manager calls", *seen)
	}
}

func TestResolveSecretEnvReplacesRefs(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "proj")
	t.Setenv("DB_PASSWORD", "sm://db-password")
	t.Setenv("API_KEY", "sm://api-key#2")
	t.Setenv("CROSS", "sm://projects/other/secrets/n/versions/3")
	t.Setenv("PLAIN_VALUE", "literal")

	stubSecretFetcher(t, map[string]string{
		"projects/proj/secrets/db-password/versions/latest": "s3cret\n",
		"projects/proj/secrets/api-key/versions/2":          "key-2",
		"projects/other/secrets/n/versions/3":               "cross-value",
	}, nil)

	if err := ResolveSecretEnv(context.Background()); err != nil {
		t.Fatalf("ResolveSecretEnv: %v", err)
	}

	want := map[string]string{
		"DB_PASSWORD": "s3cret", // trailing newline trimmed
		"API_KEY":     "key-2",
		"CROSS":       "cross-value",
		"PLAIN_VALUE": "literal",
	}
	for key, val := range want {
		if got := os.Getenv(key); got != val {
			t.Errorf("%s = %q, want %q", key, got, val)
		}
	}
}

func TestResolveSecretEnvLeavesEnvUntouchedOnFailure(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "proj")
	t.Setenv("DB_PASSWORD", "sm://db-password")
	stubSecretFetcher(t, nil, errors.New("permission denied"))

	err := ResolveSecretEnv(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Errorf("error %q should name the failing env var", err)
	}
	if got := os.Getenv("DB_PASSWORD"); got != "sm://db-password" {
		t.Errorf("DB_PASSWORD = %q, want the reference left untouched", got)
	}
}
