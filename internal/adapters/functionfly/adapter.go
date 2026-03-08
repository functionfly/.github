package functionfly

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/signing"
	"github.com/functionfly/functionfly/internal/storage"
)

const (
	ProviderName   = "functionfly-edge"
	RequestTimeout = 30 * time.Second
)

type FunctionFlyAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

func NewFunctionFlyAdapter() *FunctionFlyAdapter {
	return &FunctionFlyAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{Timeout: RequestTimeout},
	}
}

func (a *FunctionFlyAdapter) GetName() string { return ProviderName }

func (a *FunctionFlyAdapter) ValidateConfig(region, url string) error {
	if url == "" {
		return fmt.Errorf("URL is required for %s provider", ProviderName)
	}
	return nil
}

func (a *FunctionFlyAdapter) GetRegions() []string {
	return []string{"us-east-1", "eu-west-1", "ap-southeast-1"}
}

func (a *FunctionFlyAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
	return &common.HealthCheckResult{OK: true, StatusCode: 200}, nil
}

func (a *FunctionFlyAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	secret := ""
	if backend != nil {
		secret = backend.SharedSecret
	}
	return a.signer.SignRequest(req, secret, timestamp)
}

func (a *FunctionFlyAdapter) GetRequestTimeout() time.Duration { return RequestTimeout }

func (a *FunctionFlyAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	return &common.DeploymentResult{Status: common.DeploymentStatusPending, Message: "not implemented"}, nil
}

func (a *FunctionFlyAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	return nil
}

func (a *FunctionFlyAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	return nil
}

func (a *FunctionFlyAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	return common.DeploymentStatusPending, nil
}

func (a *FunctionFlyAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	return &common.DeploymentResult{Status: common.DeploymentStatusPending, Message: "not implemented"}, nil
}
