package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// InvokeRequest contains parameters for a Lambda invocation.
type InvokeRequest struct {
	FunctionName string
	Payload      interface{}
	InvocationType string // "RequestResponse", "Event", "DryRun"
	Qualifier    string // version or alias
}

// InvokeResult contains the result of a Lambda invocation.
type InvokeResult struct {
	StatusCode    int
	ExecutedVersion string
	Payload       []byte
	FunctionError string
	LogResult     string
}

// InvokeFunction invokes an AWS Lambda function synchronously.
func (a *AWSAdapter) InvokeFunction(ctx context.Context, creds *AWSCredentials, req InvokeRequest) (*InvokeResult, error) {
	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda client: %w", err)
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	invType := lambdatypes.InvocationTypeRequestResponse
	switch req.InvocationType {
	case "Event":
		invType = lambdatypes.InvocationTypeEvent
	case "DryRun":
		invType = lambdatypes.InvocationTypeDryRun
	}

	input := &lambdaservice.InvokeInput{
		FunctionName:   aws.String(req.FunctionName),
		Payload:        payload,
		InvocationType: invType,
	}

	if req.Qualifier != "" {
		input.Qualifier = aws.String(req.Qualifier)
	}

	output, err := client.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}

	result := &InvokeResult{
		StatusCode:      int(output.StatusCode),
		ExecutedVersion: aws.ToString(output.ExecutedVersion),
		Payload:         output.Payload,
	}

	if output.FunctionError != nil {
		result.FunctionError = *output.FunctionError
	}
	if output.LogResult != nil {
		result.LogResult = *output.LogResult
	}

	return result, nil
}

// InvokeAsync invokes an AWS Lambda function asynchronously.
func (a *AWSAdapter) InvokeAsync(ctx context.Context, creds *AWSCredentials, functionName string, payload interface{}) (*InvokeResult, error) {
	return a.InvokeFunction(ctx, creds, InvokeRequest{
		FunctionName:   functionName,
		Payload:        payload,
		InvocationType: "Event",
	})
}

// PublishVersion publishes a new version of a Lambda function.
func (a *AWSAdapter) PublishVersion(ctx context.Context, creds *AWSCredentials, functionName, description string) (string, error) {
	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return "", fmt.Errorf("failed to create Lambda client: %w", err)
	}

	input := &lambdaservice.PublishVersionInput{
		FunctionName: aws.String(functionName),
	}
	if description != "" {
		input.Description = aws.String(description)
	}

	output, err := client.PublishVersion(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to publish version: %w", err)
	}

	return aws.ToString(output.Version), nil
}

// CreateAlias creates an alias for a Lambda function version.
func (a *AWSAdapter) CreateAlias(ctx context.Context, creds *AWSCredentials, functionName, aliasName, functionVersion string) error {
	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create Lambda client: %w", err)
	}

	_, err = client.CreateAlias(ctx, &lambdaservice.CreateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(functionVersion),
	})
	if err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	return nil
}

// UpdateAlias updates an alias to point to a new function version.
func (a *AWSAdapter) UpdateAlias(ctx context.Context, creds *AWSCredentials, functionName, aliasName, functionVersion string) error {
	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create Lambda client: %w", err)
	}

	_, err = client.UpdateAlias(ctx, &lambdaservice.UpdateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(functionVersion),
	})
	if err != nil {
		return fmt.Errorf("failed to update alias: %w", err)
	}

	return nil
}

// GetAccountSettings returns Lambda account usage and limits.
func (a *AWSAdapter) GetAccountSettings(ctx context.Context, creds *AWSCredentials) (*lambdaservice.GetAccountSettingsOutput, error) {
	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda client: %w", err)
	}

	return client.GetAccountSettings(ctx, &lambdaservice.GetAccountSettingsInput{})
}

// DeleteFunction deletes an AWS Lambda function.
func (a *AWSAdapter) DeleteFunction(ctx context.Context, creds *AWSCredentials, functionName string) error {
	client, err := a.newLambdaClient(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to create Lambda client: %w", err)
	}

	_, err = client.DeleteFunction(ctx, &lambdaservice.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	return nil
}
