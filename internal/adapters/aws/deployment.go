package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/functionfly/functionfly/internal/adapters/common"
)

// Deploy creates or updates an AWS Lambda function.
func (a *AWSAdapter) Deploy(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	creds, functionName, cfg := a.extractDeployConfig(spec)
	if creds == nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "missing AWS credentials in provider config",
		}, nil
	}
	if functionName == "" {
		functionName = spec.AppName
	}
	if functionName == "" {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "missing function name (set app_name or function_name in provider config)",
		}, nil
	}

	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to create Lambda client: %v", err),
		}, nil
	}

	runtime := a.resolveRuntime(spec, cfg)
	handler := a.resolveHandler(cfg)
	roleARN := a.resolveRoleARN(creds, cfg)
	memoryMB := a.resolveMemoryMB(cfg)
	timeoutSec := a.resolveTimeoutSec(cfg)

	exists, err := a.functionExists(ctx, client, functionName)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to check function existence: %v", err),
		}, nil
	}

	if exists {
		return a.updateFunction(ctx, client, spec, functionName, runtime, handler, memoryMB, timeoutSec)
	}
	return a.createFunction(ctx, client, spec, functionName, runtime, handler, roleARN, memoryMB, timeoutSec)
}

// SetEnv updates environment variables for an existing Lambda function.
func (a *AWSAdapter) SetEnv(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, envVars, secrets map[string]string) error {
	creds, err := a.parseProviderCreds(providerConfig)
	if err != nil {
		return err
	}

	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create Lambda client: %w", err)
	}

	merged := make(map[string]string)
	for k, v := range envVars {
		merged[k] = v
	}
	for k, v := range secrets {
		merged[k] = v
	}

	_, err = client.UpdateFunctionConfiguration(ctx, &lambdaservice.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(deploymentID),
		Environment:  &lambdatypes.Environment{Variables: merged},
	})
	if err != nil {
		return fmt.Errorf("failed to update function environment: %w", err)
	}

	return nil
}

// BindRoutes creates a Lambda Function URL for the deployment.
func (a *AWSAdapter) BindRoutes(ctx context.Context, deploymentID string, providerConfig map[string]interface{}, routes []common.RouteBinding) error {
	creds, err := a.parseProviderCreds(providerConfig)
	if err != nil {
		return err
	}

	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create Lambda client: %w", err)
	}

	authType := lambdatypes.FunctionUrlAuthTypeNone
	cors := &lambdatypes.Cors{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:       aws.Int32(86400),
	}

	_, err = client.CreateFunctionUrlConfig(ctx, &lambdaservice.CreateFunctionUrlConfigInput{
		FunctionName: aws.String(deploymentID),
		AuthType:     authType,
		Cors:         cors,
	})
	if err != nil {
		return fmt.Errorf("failed to create Function URL: %w", err)
	}

	return nil
}

// GetDeploymentStatus returns the current status of a Lambda function.
func (a *AWSAdapter) GetDeploymentStatus(ctx context.Context, deploymentID string, providerConfig map[string]interface{}) (common.DeploymentStatus, error) {
	creds, err := a.parseProviderCreds(providerConfig)
	if err != nil {
		return common.DeploymentStatusFailed, err
	}

	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to create Lambda client: %w", err)
	}

	output, err := client.GetFunction(ctx, &lambdaservice.GetFunctionInput{
		FunctionName: aws.String(deploymentID),
	})
	if err != nil {
		return common.DeploymentStatusFailed, fmt.Errorf("failed to get function: %w", err)
	}

	switch output.Configuration.State {
	case lambdatypes.StateActive:
		if output.Configuration.LastUpdateStatus == lambdatypes.LastUpdateStatusInProgress {
			return common.DeploymentStatusDeploying, nil
		}
		return common.DeploymentStatusSuccess, nil
	case lambdatypes.StatePending:
		return common.DeploymentStatusDeploying, nil
	case lambdatypes.StateFailed:
		return common.DeploymentStatusFailed, nil
	default:
		return common.DeploymentStatusPending, nil
	}
}

// Rollback updates a Lambda function code to a previous artifact.
func (a *AWSAdapter) Rollback(ctx context.Context, spec *common.DeploymentSpec) (*common.DeploymentResult, error) {
	creds, functionName, _ := a.extractDeployConfig(spec)
	if creds == nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "missing AWS credentials for rollback",
		}, nil
	}
	if functionName == "" {
		functionName = spec.AppName
	}

	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to create Lambda client: %v", err),
		}, nil
	}

	output, err := client.UpdateFunctionCode(ctx, &lambdaservice.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		ZipFile:      spec.Artifact,
	})
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("rollback failed: %v", err),
		}, nil
	}

	return &common.DeploymentResult{
		DeploymentID:  aws.ToString(output.FunctionArn),
		Status:        common.DeploymentStatusSuccess,
		Message:       "Rollback completed successfully",
		DeploymentURL: a.buildFunctionURL(output),
		Metadata: map[string]interface{}{
			"function_name": aws.ToString(output.FunctionName),
			"function_arn":  aws.ToString(output.FunctionArn),
			"last_modified":  aws.ToString(output.LastModified),
			"version":       aws.ToString(output.Version),
		},
	}, nil
}

// --- internal helpers ---

func (a *AWSAdapter) extractDeployConfig(spec *common.DeploymentSpec) (creds *AWSCredentials, functionName string, cfg map[string]interface{}) {
	cfg = spec.ProviderConfig
	if cfg == nil {
		return nil, "", nil
	}

	if ak, ok := cfg["access_key_id"].(string); ok {
		sk, _ := cfg["secret_access_key"].(string)
		region, _ := cfg["region"].(string)
		role, _ := cfg["execution_role_arn"].(string)
		creds = &AWSCredentials{
			AccessKeyID:     ak,
			SecretAccessKey:  sk,
			Region:          region,
			ExecutionRoleARN: role,
		}
	}

	if fn, ok := cfg["function_name"].(string); ok {
		functionName = fn
	}

	return creds, functionName, cfg
}

func (a *AWSAdapter) parseProviderCreds(providerConfig map[string]interface{}) (*AWSCredentials, error) {
	if providerConfig == nil {
		return nil, fmt.Errorf("provider config is nil")
	}

	ak, _ := providerConfig["access_key_id"].(string)
	sk, _ := providerConfig["secret_access_key"].(string)
	region, _ := providerConfig["region"].(string)
	role, _ := providerConfig["execution_role_arn"].(string)

	if ak == "" || sk == "" || region == "" {
		return nil, fmt.Errorf("missing required AWS credentials (access_key_id, secret_access_key, region)")
	}

	return &AWSCredentials{
		AccessKeyID:     ak,
		SecretAccessKey:  sk,
		Region:          region,
		ExecutionRoleARN: role,
	}, nil
}

func (a *AWSAdapter) resolveRuntime(spec *common.DeploymentSpec, cfg map[string]interface{}) string {
	if cfg != nil {
		if rt, ok := cfg["runtime"].(string); ok && rt != "" {
			return rt
		}
	}
	switch spec.Runtime {
	case common.RuntimeJavaScript:
		return "nodejs20.x"
	case common.RuntimeWASM:
		return "provided.al2023"
	default:
		return "nodejs20.x"
	}
}

func (a *AWSAdapter) resolveHandler(cfg map[string]interface{}) string {
	if cfg != nil {
		if h, ok := cfg["handler"].(string); ok && h != "" {
			return h
		}
	}
	return "index.handler"
}

func (a *AWSAdapter) resolveRoleARN(creds *AWSCredentials, cfg map[string]interface{}) string {
	if cfg != nil {
		if role, ok := cfg["execution_role_arn"].(string); ok && role != "" {
			return role
		}
	}
	if creds != nil && creds.ExecutionRoleARN != "" {
		return creds.ExecutionRoleARN
	}
	return ""
}

func (a *AWSAdapter) resolveMemoryMB(cfg map[string]interface{}) int32 {
	if cfg != nil {
		if mem, ok := cfg["memory_mb"].(float64); ok && mem > 0 {
			return int32(mem)
		}
		if mem, ok := cfg["memory_mb"].(int); ok && mem > 0 {
			return int32(mem)
		}
	}
	return defaultMemoryMB
}

func (a *AWSAdapter) resolveTimeoutSec(cfg map[string]interface{}) int32 {
	if cfg != nil {
		if t, ok := cfg["timeout_sec"].(float64); ok && t > 0 {
			return int32(t)
		}
		if t, ok := cfg["timeout_sec"].(int); ok && t > 0 {
			return int32(t)
		}
	}
	return defaultTimeoutSec
}

func (a *AWSAdapter) functionExists(ctx context.Context, client *lambdaservice.Client, name string) (bool, error) {
	_, err := client.GetFunction(ctx, &lambdaservice.GetFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (a *AWSAdapter) createFunction(ctx context.Context, client *lambdaservice.Client, spec *common.DeploymentSpec, functionName, runtime, handler, roleARN string, memoryMB, timeoutSec int32) (*common.DeploymentResult, error) {
	if roleARN == "" {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: "execution_role_arn is required to create a new Lambda function",
		}, nil
	}

	input := &lambdaservice.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Runtime:      lambdatypes.Runtime(runtime),
		Handler:      aws.String(handler),
		Role:         aws.String(roleARN),
		Code:         &lambdatypes.FunctionCode{ZipFile: spec.Artifact},
		MemorySize:   aws.Int32(memoryMB),
		Timeout:      aws.Int32(timeoutSec),
		Publish:      true,
		Environment:  a.buildEnvironment(spec),
		Tags:         a.buildTags(spec),
		Description:  aws.String(fmt.Sprintf("Deployed via FunctionFly — %s", spec.Version)),
	}

	output, err := client.CreateFunction(ctx, input)
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to create function: %v", err),
		}, nil
	}

	return &common.DeploymentResult{
		DeploymentID:  aws.ToString(output.FunctionArn),
		Status:        common.DeploymentStatusSuccess,
		Message:       fmt.Sprintf("Function '%s' created successfully", functionName),
		DeploymentURL: a.buildFunctionURL(output),
		Metadata: map[string]interface{}{
			"function_name": aws.ToString(output.FunctionName),
			"function_arn":  aws.ToString(output.FunctionArn),
			"runtime":       string(output.Runtime),
			"handler":       aws.ToString(output.Handler),
			"memory_size":   aws.ToInt32(output.MemorySize),
			"timeout":       aws.ToInt32(output.Timeout),
			"last_modified": aws.ToString(output.LastModified),
			"version":       aws.ToString(output.Version),
		},
	}, nil
}

func (a *AWSAdapter) updateFunction(ctx context.Context, client *lambdaservice.Client, spec *common.DeploymentSpec, functionName, runtime, handler string, memoryMB, timeoutSec int32) (*common.DeploymentResult, error) {
	_, err := client.UpdateFunctionCode(ctx, &lambdaservice.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		ZipFile:      spec.Artifact,
		Publish:      true,
	})
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to update function code: %v", err),
		}, nil
	}

	_, err = client.UpdateFunctionConfiguration(ctx, &lambdaservice.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		Runtime:      lambdatypes.Runtime(runtime),
		Handler:      aws.String(handler),
		MemorySize:   aws.Int32(memoryMB),
		Timeout:      aws.Int32(timeoutSec),
		Environment:  a.buildEnvironment(spec),
		Description:  aws.String(fmt.Sprintf("Updated via FunctionFly — %s", spec.Version)),
	})
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusFailed,
			Message: fmt.Sprintf("failed to update function config: %v", err),
		}, nil
	}

	output, err := client.GetFunction(ctx, &lambdaservice.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return &common.DeploymentResult{
			Status:  common.DeploymentStatusSuccess,
			Message: fmt.Sprintf("Function '%s' updated successfully", functionName),
			Metadata: map[string]interface{}{
				"function_name": functionName,
			},
		}, nil
	}

	cfg := output.Configuration
	return &common.DeploymentResult{
		DeploymentID: aws.ToString(cfg.FunctionArn),
		Status:       common.DeploymentStatusSuccess,
		Message:      fmt.Sprintf("Function '%s' updated successfully", functionName),
		Metadata: map[string]interface{}{
			"function_name": aws.ToString(cfg.FunctionName),
			"function_arn":  aws.ToString(cfg.FunctionArn),
			"runtime":       string(cfg.Runtime),
			"handler":       aws.ToString(cfg.Handler),
			"memory_size":   aws.ToInt32(cfg.MemorySize),
			"timeout":       aws.ToInt32(cfg.Timeout),
			"last_modified": aws.ToString(cfg.LastModified),
			"version":       aws.ToString(cfg.Version),
			"state":         string(cfg.State),
			"state_reason":  aws.ToString(cfg.StateReason),
		},
	}, nil
}

func (a *AWSAdapter) buildEnvironment(spec *common.DeploymentSpec) *lambdatypes.Environment {
	if len(spec.EnvVars) == 0 && len(spec.Secrets) == 0 {
		return nil
	}
	variables := make(map[string]string)
	for k, v := range spec.EnvVars {
		variables[k] = v
	}
	for k, v := range spec.Secrets {
		variables[k] = v
	}
	return &lambdatypes.Environment{Variables: variables}
}

func (a *AWSAdapter) buildTags(spec *common.DeploymentSpec) map[string]string {
	tags := map[string]string{
		"ManagedBy":  "FunctionFly",
		"DeployedAt": time.Now().UTC().Format(time.RFC3339),
	}
	if spec.Version != "" {
		tags["Version"] = spec.Version
	}
	if spec.Environment != "" {
		tags["Environment"] = spec.Environment
	}
	return tags
}

func (a *AWSAdapter) buildFunctionURL(output interface{}) string {
	// The Function URL is created separately via BindRoutes.
	// This constructs the URL from the function ARN if available.
	switch o := output.(type) {
	case *lambdaservice.CreateFunctionOutput:
		if o.FunctionArn != nil {
			return fmt.Sprintf("https://%s.lambda-url.%s.on.aws", *o.FunctionName, "us-east-1")
		}
	case *lambdaservice.GetFunctionOutput:
		if o.Configuration != nil && o.Configuration.FunctionArn != nil {
			return ""
		}
	}
	return ""
}
