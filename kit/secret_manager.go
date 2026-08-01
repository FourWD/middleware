package kit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"golang.org/x/sync/errgroup"
)

// secretRefScheme marks an env value as a Secret Manager reference, not a literal.
const secretRefScheme = "sm://"

// secretFetchLimit caps concurrent AccessSecretVersion calls during boot.
const secretFetchLimit = 8

type secretFetcher func(ctx context.Context, name string) (string, error)

// newSecretFetcher is swapped in tests so they need no GCP credentials.
var newSecretFetcher = newGoogleSecretFetcher

func newGoogleSecretFetcher(ctx context.Context) (secretFetcher, func(), error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("secretmanager client: %w", err)
	}
	fetch := func(ctx context.Context, name string) (string, error) {
		res, err := client.AccessSecretVersion(ctx,
			&secretmanagerpb.AccessSecretVersionRequest{Name: name})
		if err != nil {
			return "", err
		}
		return string(res.Payload.Data), nil
	}
	return fetch, func() { _ = client.Close() }, nil
}

// ResolveSecretEnv rewrites every "sm://" env var in place with its Secret
// Manager payload, so .env keeps only a reference while the credential lives in
// Secret Manager. Plain values pass through untouched — local dev keeps
// plaintext in .env and needs no GCP credentials at all.
//
// Reference forms accepted after the scheme:
//
//	name                            → latest version in the current project
//	name#5                          → version 5
//	projects/p/secrets/n/versions/v → used as-is (cross-project)
//
// Failure is fatal by design: a service booting with an unresolved "sm://"
// string would dial MySQL with a literal reference as the password and fail
// deeper, with a worse error.
//
// infra.NewApp calls this right after loading the .env files. Call it directly
// only from entrypoints that do not go through NewApp.
func ResolveSecretEnv(ctx context.Context) error {
	refs := map[string]string{}
	needProject := false
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(v, secretRefScheme) {
			continue
		}
		ref := strings.TrimPrefix(v, secretRefScheme)
		refs[k] = ref
		if !strings.HasPrefix(ref, "projects/") {
			needProject = true
		}
	}
	if len(refs) == 0 {
		return nil // nothing migrated, or local dev running on plaintext .env
	}

	project := ""
	if needProject {
		project = secretProjectID(ctx)
	}

	fetch, closeFetcher, err := newSecretFetcher(ctx)
	if err != nil {
		return err
	}
	defer closeFetcher()

	var mu sync.Mutex
	out := make(map[string]string, len(refs))

	// Fetched concurrently so boot pays one round trip, not one per secret.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(secretFetchLimit)
	for key, ref := range refs {
		g.Go(func() error {
			name, err := secretVersionName(ref, project)
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			payload, err := fetch(gctx, name)
			if err != nil {
				return fmt.Errorf("%s (%s): %w", key, name, err)
			}
			mu.Lock()
			// A trailing newline is a `gcloud secrets create` footgun (echo pipes
			// one in), never part of a token or password.
			out[key] = strings.TrimRight(payload, "\r\n")
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	// Applied only after every fetch succeeded, so a partial failure leaves the
	// environment untouched rather than half-resolved.
	for key, val := range out {
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}
	return nil
}

// secretProjectID resolves the project that a short "sm://name" reference
// belongs to. The App Engine runtime exports GOOGLE_CLOUD_PROJECT; Cloud Run,
// GKE and Compute only expose the project through the metadata server.
func secretProjectID(ctx context.Context) string {
	for _, key := range []string{"GOOGLE_CLOUD_PROJECT", "GCP_PROJECT", "GCLOUD_PROJECT"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	if metadata.OnGCE() {
		if id, err := metadata.ProjectIDWithContext(ctx); err == nil {
			return id
		}
	}
	return ""
}

func secretVersionName(ref, project string) (string, error) {
	if strings.HasPrefix(ref, "projects/") {
		return ref, nil
	}
	name, version, ok := strings.Cut(ref, "#")
	if !ok || version == "" {
		version = "latest"
	}
	if name == "" {
		return "", fmt.Errorf("empty secret name")
	}
	if project == "" {
		return "", fmt.Errorf("no GCP project detected; set GOOGLE_CLOUD_PROJECT or use a full projects/... reference")
	}
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, name, version), nil
}
