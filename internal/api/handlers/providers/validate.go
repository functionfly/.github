package providers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleValidateProvider validates a provider API token
func (h *Handler) HandleValidateProvider(w http.ResponseWriter, r *http.Request) {
	var req ProviderValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	response := h.validateProviderToken(req.Provider, req.Token)

	if response.IsValid {
		idBytes := make([]byte, 16)
		rand.Read(idBytes)
		providerID := hex.EncodeToString(idBytes)

		provider := &storage.Provider{
			ID:       providerID,
			UserID:   claims.UserID,
			Provider: req.Provider,
			Token:    req.Token,
			Status:   "active",
		}

		if err := h.repo.CreateProvider(r.Context(), provider); err != nil {
			logrus.WithError(err).Error("Failed to store provider")
			response.IsValid = false
			response.Message = "Failed to save provider configuration"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) validateProviderToken(provider, token string) ProviderValidationResponse {
	switch provider {
	case "cloudflare":
		return h.validateCloudflareToken(token)
	case "vercel":
		return h.validateVercelToken(token)
	case "fly":
		return h.validateFlyToken(token)
	case "aws-lambda":
		return h.validateAWSToken(token)
	default:
		return ProviderValidationResponse{
			IsValid: false,
			Message: "Unsupported provider",
		}
	}
}

func (h *Handler) validateCloudflareToken(token string) ProviderValidationResponse {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/accounts", nil)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to create validation request: %v", err),
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to validate token: %v", err),
		}
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to parse validation response: %v", err),
		}
	}

	if !result.Success || len(result.Errors) > 0 {
		errorMsg := "token validation failed"
		if len(result.Errors) > 0 {
			errorMsg = result.Errors[0].Message
		}
		return ProviderValidationResponse{
			IsValid: false,
			Message: errorMsg,
		}
	}

	if len(result.Result) == 0 {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "token is valid but has no accessible accounts",
		}
	}

	return ProviderValidationResponse{
		IsValid: true,
		Message: "Cloudflare token validated successfully",
		UserID:  result.Result[0].ID,
		Email:   result.Result[0].Name,
	}
}

func (h *Handler) validateVercelToken(token string) ProviderValidationResponse {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.vercel.com/v2/user", nil)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to create validation request: %v", err),
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to validate token: %v", err),
		}
	}
	defer resp.Body.Close()

	var result struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"user"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to parse validation response: %v", err),
		}
	}

	if result.Error.Code != "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: result.Error.Message,
		}
	}

	if result.User.ID == "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "invalid token or unable to retrieve user information",
		}
	}

	return ProviderValidationResponse{
		IsValid: true,
		Message: "Vercel token validated successfully",
		UserID:  result.User.ID,
		Email:   result.User.Email,
	}
}

func (h *Handler) validateFlyToken(token string) ProviderValidationResponse {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.fly.io/api/v1/user", nil)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to create validation request: %v", err),
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to validate token: %v", err),
		}
	}
	defer resp.Body.Close()

	var result struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to parse validation response: %v", err),
		}
	}

	if result.Error != "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: result.Error,
		}
	}

	if result.ID == "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "invalid token or unable to retrieve user information",
		}
	}

	return ProviderValidationResponse{
		IsValid: true,
		Message: "Fly.io token validated successfully",
		UserID:  result.ID,
		Email:   result.Email,
	}
}

func (h *Handler) validateAWSToken(token string) ProviderValidationResponse {
	parts := strings.SplitN(token, "|", 4)
	if len(parts) < 3 {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "Invalid AWS credential format. Expected: AccessKeyID|SecretAccessKey|Region[|RoleARN]",
		}
	}

	accessKeyID := strings.TrimSpace(parts[0])
	secretAccessKey := strings.TrimSpace(parts[1])
	region := strings.TrimSpace(parts[2])

	if accessKeyID == "" || secretAccessKey == "" || region == "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "AWS Access Key ID, Secret Access Key, and Region are all required",
		}
	}

	if !strings.HasPrefix(accessKeyID, "AKIA") || len(accessKeyID) != 20 {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "Invalid AWS Access Key ID format (must start with AKIA and be 20 characters)",
		}
	}

	if len(secretAccessKey) != 40 {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "Invalid AWS Secret Access Key format (must be 40 characters)",
		}
	}

	validRegion := false
	for _, r := range []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"af-south-1", "ap-east-1", "ap-south-1", "ap-south-2",
		"ap-southeast-1", "ap-southeast-2", "ap-southeast-3",
		"ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
		"ca-central-1",
		"eu-central-1", "eu-central-2", "eu-west-1", "eu-west-2", "eu-west-3",
		"eu-south-1", "eu-south-2", "eu-north-1",
		"me-south-1", "me-central-1",
		"sa-east-1",
	} {
		if region == r {
			validRegion = true
			break
		}
	}
	if !validRegion {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("Invalid AWS region: %s", region),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			awscredentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("Failed to configure AWS client: %v", err),
		}
	}

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("AWS credential validation failed: %v", err),
		}
	}

	accountID := aws.ToString(identity.Account)
	arn := aws.ToString(identity.Arn)

	return ProviderValidationResponse{
		IsValid: true,
		Message: fmt.Sprintf("AWS credentials validated — Account: %s", accountID),
		UserID:  arn,
		Email:   accountID,
	}
}
