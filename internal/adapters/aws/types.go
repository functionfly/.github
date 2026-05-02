package aws

import "time"

const (
	ProviderName   = "aws-lambda"
	RequestTimeout = 30 * time.Second

	defaultMemoryMB    = 128
	defaultTimeoutSec  = 30
	maxMemoryMB        = 10240
	maxTimeoutSec      = 900
	maxFunctionNameLen = 64
)

// AWSCredentials holds parsed AWS credentials from the provider token.
type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey  string
	Region          string
	ExecutionRoleARN string
}

// LambdaFunctionConfig holds Lambda function configuration.
type LambdaFunctionConfig struct {
	FunctionName string
	Runtime      string // nodejs18.x, python3.11, etc.
	Handler      string
	MemoryMB     int
	TimeoutSec   int
	RoleARN      string
	Description  string
	Tags         map[string]string
}

// DefaultLambdaConfig returns a LambdaFunctionConfig with sensible defaults.
func DefaultLambdaConfig() LambdaFunctionConfig {
	return LambdaFunctionConfig{
		Runtime:    "nodejs18.x",
		Handler:    "index.handler",
		MemoryMB:   defaultMemoryMB,
		TimeoutSec: defaultTimeoutSec,
	}
}

// SupportedRuntimes lists the Lambda runtimes we support.
var SupportedRuntimes = map[string]bool{
	"nodejs18.x":  true,
	"nodejs20.x":  true,
	"python3.10":  true,
	"python3.11":  true,
	"python3.12":  true,
	"java21":      true,
	"dotnet8":     true,
	"provided.al2023": true,
}

// AWSSupportedRegions lists Lambda-supported regions.
var AWSSupportedRegions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	"af-south-1",
	"ap-east-1",
	"ap-south-1",
	"ap-south-2",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",
	"ca-central-1",
	"eu-central-1",
	"eu-central-2",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"eu-south-1",
	"eu-south-2",
	"eu-north-1",
	"me-south-1",
	"me-central-1",
	"sa-east-1",
}
