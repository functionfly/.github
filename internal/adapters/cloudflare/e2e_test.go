package cloudflare

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

// E2E tests run against the real Cloudflare API when env vars are set.
// Set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN to run.
// Optionally set CLOUDFLARE_WORKERS_SUBDOMAIN to test HTTP invocation (e.g. "mycompany" for mycompany.workers.dev).
// Run: go test ./internal/adapters/cloudflare -run E2E -v -timeout 120s

const e2eScriptPrefix = "functionfly-e2e-"

func e2eSkipOrConfig(t *testing.T) (accountID, apiToken string) {
	t.Helper()
	accountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	if accountID == "" || apiToken == "" {
		t.Skip("E2E skipped: set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN to run")
	}
	return accountID, apiToken
}

var e2eMinimalScript = []byte(`addEventListener('fetch', e => {
  const u = new URL(e.request.url);
  if (u.pathname === '/healthz') return e.respondWith(new Response('ok', { status: 200 }));
  e.respondWith(new Response('OK', { status: 200 }));
});`)

// e2eGetWithRetry GETs url with retries and backoff (Worker can take a few seconds to propagate).
func e2eGetWithRetry(t *testing.T, ctx context.Context, client *CloudflareDeploymentClient, url, label string) {
	t.Helper()
	const maxAttempts = 4
	const delayBetween = 5 * time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			select {
			case <-ctx.Done():
				t.Errorf("%s: context done before retry", label)
				return
			case <-time.After(delayBetween):
			}
		}
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err := client.httpClient.Do(req)
		if err != nil {
			t.Logf("%s attempt %d: %v", label, attempt, err)
			continue
		}
		_ = resp.Body.Close()
		t.Logf("%s %s -> %d (attempt %d)", label, url, resp.StatusCode, attempt)
		if resp.StatusCode == http.StatusOK {
			return
		}
	}
	t.Errorf("%s: got non-200 after %d attempts", label, maxAttempts)
}

func TestE2EDeployStatusAndCleanup(t *testing.T) {
	accountID, apiToken := e2eSkipOrConfig(t)
	scriptName := e2eScriptPrefix + time.Now().Format("20060102150405")

	client := NewCloudflareDeploymentClient(apiToken, accountID)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Cleanup: delete script at end (best effort)
	t.Cleanup(func() {
		_ = client.DeleteDeployment(context.Background(), scriptName)
	})

	// 1. Deploy
	result, err := client.Deploy(ctx, e2eMinimalScript, scriptName)
	if err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if result.Status != common.DeploymentStatusSuccess {
		t.Fatalf("Deploy status: %s", result.Status)
	}
	t.Logf("Deployed script %s", scriptName)

	// 2. Status
	status, err := client.GetDeploymentStatus(ctx, scriptName)
	if err != nil {
		t.Fatalf("GetDeploymentStatus: %v", err)
	}
	if status != common.DeploymentStatusSuccess {
		t.Fatalf("GetDeploymentStatus: %s", status)
	}

	// 3. Optional: enable workers.dev and HTTP GET if subdomain set (workers.dev is disabled by default when deploying via API)
	subdomain := os.Getenv("CLOUDFLARE_WORKERS_SUBDOMAIN")
	if subdomain != "" {
		if err := client.EnableWorkersDev(ctx, scriptName); err != nil {
			t.Fatalf("EnableWorkersDev: %v", err)
		}
		t.Log("EnableWorkersDev OK")
		baseURL := "https://" + scriptName + "." + subdomain + ".workers.dev"
		e2eGetWithRetry(t, ctx, client, baseURL+"/", "GET /")
		e2eGetWithRetry(t, ctx, client, baseURL+"/healthz", "GET /healthz")
	}
}

func TestE2ESetEnvAndRollback(t *testing.T) {
	accountID, apiToken := e2eSkipOrConfig(t)
	scriptName := e2eScriptPrefix + "env-" + time.Now().Format("20060102150405")

	client := NewCloudflareDeploymentClient(apiToken, accountID)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	t.Cleanup(func() {
		_ = client.DeleteDeployment(context.Background(), scriptName)
	})

	// Deploy
	_, err := client.Deploy(ctx, e2eMinimalScript, scriptName)
	if err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	// Set env vars (plain + secret)
	err = client.SetEnvironmentVariables(ctx, scriptName, map[string]string{"E2E_VAR": "hello"}, map[string]string{"E2E_SECRET": "secret"})
	if err != nil {
		t.Fatalf("SetEnvironmentVariables: %v", err)
	}
	t.Log("SetEnvironmentVariables OK")

	// Rollback (redeploy same script)
	rollbackResult, err := client.Rollback(ctx, e2eMinimalScript, scriptName)
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if rollbackResult.Status != common.DeploymentStatusSuccess {
		t.Fatalf("Rollback status: %s", rollbackResult.Status)
	}
	t.Log("Rollback OK")
}
