package providers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/functionfly/functionfly/internal/apierror"
)

type discoveredResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// HandleDiscoverResources proxies provider API calls to list deployable
// resources (Cloudflare Workers, Vercel projects, etc.) so the frontend
// can populate a dropdown without hitting CORS-restricted provider APIs.
func (h *Handler) HandleDiscoverResources(w http.ResponseWriter, r *http.Request) {
	providerID := r.URL.Query().Get("provider")
	apiKey := r.Header.Get("X-Provider-Key")

	if apiKey == "" {
		apierror.WriteError(w, apierror.NewBadRequest("X-Provider-Key header is required"))
		return
	}

	var resources []discoveredResource
	var err error

	switch providerID {
	case "workers":
		resources, err = discoverCloudflareWorkers(apiKey)
	case "vercel":
		resources, err = discoverVercelProjects(apiKey)
	default:
		apierror.WriteError(w, apierror.NewBadRequest(fmt.Sprintf("Discovery not supported for provider: %s", providerID)))
		return
	}

	if err != nil {
		apierror.WriteError(w, apierror.NewInternal(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"resources": resources})
}

func discoverCloudflareWorkers(apiKey string) ([]discoveredResource, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Cloudflare API: %w", err)
	}
	defer resp.Body.Close()

	var accountsResp struct {
		Success bool `json:"success"`
		Result  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accountsResp); err != nil {
		return nil, fmt.Errorf("failed to decode accounts response: %w", err)
	}
	if !accountsResp.Success || len(accountsResp.Result) == 0 {
		if len(accountsResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare: %s", accountsResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("no Cloudflare accounts found for this API key")
	}

	accountID := accountsResp.Result[0].ID

	// Get the workers.dev subdomain (e.g. "microog" from microog.workers.dev)
	reqSub, _ := http.NewRequest("GET", fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/subdomain", accountID), nil)
	reqSub.Header.Set("Authorization", "Bearer "+apiKey)
	respSub, err := client.Do(reqSub)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers.dev subdomain: %w", err)
	}
	defer respSub.Body.Close()

	var subdomainResp struct {
		Success bool `json:"success"`
		Result  struct {
			Subdomain string `json:"subdomain"`
		} `json:"result"`
	}
	if err := json.NewDecoder(respSub.Body).Decode(&subdomainResp); err != nil || !subdomainResp.Success || subdomainResp.Result.Subdomain == "" {
		return nil, fmt.Errorf("workers.dev subdomain not configured — enable it at dash.cloudflare.com → Workers & Pages → Settings")
	}
	subdomain := subdomainResp.Result.Subdomain

	req2, _ := http.NewRequest("GET", fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts", accountID), nil)
	req2.Header.Set("Authorization", "Bearer "+apiKey)
	resp2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("failed to list workers: %w", err)
	}
	defer resp2.Body.Close()

	var workersResp struct {
		Success bool `json:"success"`
		Result  []struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&workersResp); err != nil {
		return nil, fmt.Errorf("failed to decode workers response: %w", err)
	}
	if !workersResp.Success {
		if len(workersResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare: %s", workersResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("failed to list Cloudflare Workers")
	}

	var resources []discoveredResource
	for _, w := range workersResp.Result {
		resources = append(resources, discoveredResource{
			ID:   w.ID,
			Name: w.ID,
			URL:  fmt.Sprintf("https://%s.%s.workers.dev", w.ID, subdomain),
		})
	}

	return resources, nil
}

func discoverVercelProjects(apiKey string) ([]discoveredResource, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.vercel.com/v9/projects", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Vercel API: %w", err)
	}
	defer resp.Body.Close()

	var projectsResp struct {
		Projects []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&projectsResp); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}

	var resources []discoveredResource
	for _, p := range projectsResp.Projects {
		resources = append(resources, discoveredResource{
			ID:   p.ID,
			Name: p.Name,
			URL:  fmt.Sprintf("https://%s.vercel.app", p.Name),
		})
	}
	return resources, nil
}
