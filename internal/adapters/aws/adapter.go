package aws

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/signing"
	"github.com/functionfly/functionfly/internal/storage"
)

// AWSAdapter implements both ProviderAdapter and DeploymentAdapter for AWS Lambda.
type AWSAdapter struct {
	signer *signing.RequestSigner
	client *http.Client
}

// NewAWSAdapter creates a new AWS Lambda adapter.
func NewAWSAdapter() *AWSAdapter {
	return &AWSAdapter{
		signer: &signing.RequestSigner{},
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// NewAWSAdapterAlias is an alias constructor for consistency with other adapters.
func NewAWSAdapterAlias() *AWSAdapter {
	return NewAWSAdapter()
}

// GetName returns the provider name.
func (a *AWSAdapter) GetName() string {
	return ProviderName
}

// ValidateConfig validates AWS Lambda–specific configuration.
func (a *AWSAdapter) ValidateConfig(region, urlStr string) error {
	regionValid := false
	for _, r := range AWSSupportedRegions {
		if r == region {
			regionValid = true
			break
		}
	}
	if !regionValid {
		return fmt.Errorf("invalid AWS region '%s'", region)
	}

	if urlStr != "" {
		parsed := strings.TrimPrefix(urlStr, "https://")
		parsed = strings.TrimPrefix(parsed, "http://")
		if !strings.Contains(parsed, ".lambda-url.") && !strings.Contains(parsed, ".amazonaws.com") {
			return fmt.Errorf("URL must be a Lambda Function URL or API Gateway endpoint, got: %s", urlStr)
		}
	}

	return nil
}

// GetRegions returns available AWS Lambda regions.
func (a *AWSAdapter) GetRegions() []string {
	return AWSSupportedRegions
}

// HealthCheck performs a health check by listing functions (1 item) to verify credentials.
func (a *AWSAdapter) HealthCheck(ctx context.Context, backend *storage.Backend) (*common.HealthCheckResult, error) {
	startTime := time.Now()

	creds, err := ParseCredentialsFromBackend(backend)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to parse credentials: %v", err),
		}, nil
	}

	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			ErrorMessage: fmt.Sprintf("failed to create Lambda client: %v", err),
		}, nil
	}

	maxItems := int32(1)
	_, err = client.ListFunctions(ctx, &lambda.ListFunctionsInput{
		MaxItems: &maxItems,
	})
	latencyMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return &common.HealthCheckResult{
			OK:           false,
			LatencyMs:    latencyMs,
			Region:       backend.Region,
			ErrorMessage: fmt.Sprintf("Lambda health check failed: %v", err),
		}, nil
	}

	return &common.HealthCheckResult{
		OK:         true,
		StatusCode: http.StatusOK,
		LatencyMs:  latencyMs,
		Region:     backend.Region,
	}, nil
}

// SignRequest adds FunctionFly request signing. AWS SDK handles SigV4 internally.
func (a *AWSAdapter) SignRequest(req *http.Request, backend *storage.Backend, timestamp time.Time) error {
	return a.signer.SignRequest(req, backend.SharedSecret, timestamp)
}

// GetRequestTimeout returns the recommended timeout for AWS Lambda requests.
func (a *AWSAdapter) GetRequestTimeout() time.Duration {
	return RequestTimeout
}

// newLambdaClient creates an AWS Lambda client from credentials.
func (a *AWSAdapter) newLambdaClient(ctx context.Context, creds *AWSCredentials) (*lambda.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(creds.Region),
		awsconfig.WithCredentialsProvider(
			awscredentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return lambda.NewFromConfig(cfg), nil
}

// ParseCredentialsFromBackend extracts AWS credentials from a Backend's shared secret.
// The shared secret stores "ACCESS_KEY_ID|SECRET_ACCESS_KEY|REGION|ROLE_ARN".
func ParseCredentialsFromBackend(backend *storage.Backend) (*AWSCredentials, error) {
	return ParseCredentials(backend.SharedSecret)
}

// ParseCredentials parses a pipe-delimited credential string.
func ParseCredentials(raw string) (*AWSCredentials, error) {
	parts := strings.SplitN(raw, "|", 4)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid AWS credential format: expected ACCESS_KEY_ID|SECRET_ACCESS_KEY|REGION[|ROLE_ARN]")
	}

	creds := &AWSCredentials{
		AccessKeyID:    strings.TrimSpace(parts[0]),
		SecretAccessKey: strings.TrimSpace(parts[1]),
		Region:         strings.TrimSpace(parts[2]),
	}

	if len(parts) == 4 {
		creds.ExecutionRoleARN = strings.TrimSpace(parts[3])
	}

	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" || creds.Region == "" {
		return nil, fmt.Errorf("AWS credentials must not be empty")
	}

	return creds, nil
}

// FormatCredentials creates a pipe-delimited credential string for storage.
func FormatCredentials(accessKeyID, secretAccessKey, region, roleARN string) string {
	if roleARN != "" {
		return fmt.Sprintf("%s|%s|%s|%s", accessKeyID, secretAccessKey, region, roleARN)
	}
	return fmt.Sprintf("%s|%s|%s", accessKeyID, secretAccessKey, region)
}
