package runpod

import (
	"os"
	"strconv"
	"time"
)

// InferenceMode defines the inference mode to use
type InferenceMode string

const (
	// InferenceModeAPI uses OpenRouter API-based inference (default)
	InferenceModeAPI InferenceMode = "api"
	// InferenceModeSelfHosted uses self-hosted RunPod GPU inference
	InferenceModeSelfHosted InferenceMode = "self_hosted"
)

// Config holds all RunPod configuration options
type Config struct {
	// Mode determines whether to use API-based or self-hosted inference
	Mode InferenceMode

	// RunPod API Configuration
	RunPodAPIKey     string
	RunPodAPIBaseURL string

	// GPU Instance Configuration
	GPUType        string // e.g., "NVIDIA RTX A5000", "NVIDIA A100"
	GPUCount       int    // Number of GPUs to provision
	InstanceCount  int    // Number of concurrent instances
	ContainerImage string // Docker image with LLM server

	// Lifecycle Configuration
	InstanceTimeout     time.Duration // Max time an instance runs before auto-cleanup
	IdleTimeout         time.Duration // Time to wait for requests before tearing down
	ProvisioningTimeout time.Duration // Max time to wait for instance to be ready
	MaxRetries          int           // Number of retries for provisioning

	// Cost Control
	EnableAutoScaling  bool
	MinInstances       int
	MaxInstances       int
	ScaleUpThreshold   int // Requests per minute threshold to scale up
	ScaleDownThreshold int // Requests per minute threshold to scale down

	// Network Configuration
	HTTPHost        string
	HTTPPort        int
	HealthCheckPath string

	// Model Configuration
	ModelName string // Model to use for self-hosted inference
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Mode:                InferenceModeAPI,
		RunPodAPIBaseURL:    "https://api.runpod.io/graphql",
		GPUType:             "NVIDIA RTX A5000",
		GPUCount:            1,
		InstanceCount:       1,
		ContainerImage:      "ghcr.io/huggingface/text-generation-inference:2.2.0",
		InstanceTimeout:     60 * time.Minute,
		IdleTimeout:         5 * time.Minute,
		ProvisioningTimeout: 10 * time.Minute,
		MaxRetries:          3,
		EnableAutoScaling:   false,
		MinInstances:        0,
		MaxInstances:        10,
		ScaleUpThreshold:    100,
		ScaleDownThreshold:  10,
		HTTPHost:            "0.0.0.0",
		HTTPPort:            8080,
		HealthCheckPath:     "/health",
		ModelName:           "microsoft/Phi-3-mini-4k-instruct",
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := DefaultConfig()

	// Load inference mode
	if mode := os.Getenv("INFERENCE_MODE"); mode != "" {
		cfg.Mode = InferenceMode(mode)
	}

	// Load RunPod API configuration
	if apiKey := os.Getenv("RUNPOD_API_KEY"); apiKey != "" {
		cfg.RunPodAPIKey = apiKey
	}
	if baseURL := os.Getenv("RUNPOD_API_BASE_URL"); baseURL != "" {
		cfg.RunPodAPIBaseURL = baseURL
	}

	// Load GPU configuration
	if gpuType := os.Getenv("RUNPOD_GPU_TYPE"); gpuType != "" {
		cfg.GPUType = gpuType
	}
	if gpuCount := os.Getenv("RUNPOD_GPU_COUNT"); gpuCount != "" {
		if count, err := strconv.Atoi(gpuCount); err == nil && count > 0 {
			cfg.GPUCount = count
		}
	}
	if instanceCount := os.Getenv("RUNPOD_INSTANCE_COUNT"); instanceCount != "" {
		if count, err := strconv.Atoi(instanceCount); err == nil && count > 0 {
			cfg.InstanceCount = count
		}
	}
	if containerImage := os.Getenv("RUNPOD_CONTAINER_IMAGE"); containerImage != "" {
		cfg.ContainerImage = containerImage
	}

	// Load timeouts
	if timeout := os.Getenv("RUNPOD_INSTANCE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil && d > 0 {
			cfg.InstanceTimeout = d
		}
	}
	if timeout := os.Getenv("RUNPOD_IDLE_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil && d > 0 {
			cfg.IdleTimeout = d
		}
	}
	if timeout := os.Getenv("RUNPOD_PROVISIONING_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil && d > 0 {
			cfg.ProvisioningTimeout = d
		}
	}

	// Load scaling configuration
	if maxInst := os.Getenv("RUNPOD_MAX_INSTANCES"); maxInst != "" {
		if count, err := strconv.Atoi(maxInst); err == nil && count > 0 {
			cfg.MaxInstances = count
		}
	}
	if minInst := os.Getenv("RUNPOD_MIN_INSTANCES"); minInst != "" {
		if count, err := strconv.Atoi(minInst); err == nil && count > 0 {
			cfg.MinInstances = count
		}
	}
	if scaleUp := os.Getenv("RUNPOD_SCALE_UP_THRESHOLD"); scaleUp != "" {
		if count, err := strconv.Atoi(scaleUp); err == nil && count > 0 {
			cfg.ScaleUpThreshold = count
		}
	}
	if scaleDown := os.Getenv("RUNPOD_SCALE_DOWN_THRESHOLD"); scaleDown != "" {
		if count, err := strconv.Atoi(scaleDown); err == nil && count > 0 {
			cfg.ScaleDownThreshold = count
		}
	}

	// Load network configuration
	if host := os.Getenv("RUNPOD_HTTP_HOST"); host != "" {
		cfg.HTTPHost = host
	}
	if port := os.Getenv("RUNPOD_HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 && p < 65536 {
			cfg.HTTPPort = p
		}
	}
	if path := os.Getenv("RUNPOD_HEALTH_CHECK_PATH"); path != "" {
		cfg.HealthCheckPath = path
	}

	// Load model configuration
	if model := os.Getenv("RUNPOD_MODEL_NAME"); model != "" {
		cfg.ModelName = model
	}

	return cfg
}

// IsAPIOnly returns true if using API-based inference only
func (c *Config) IsAPIOnly() bool {
	return c.Mode == InferenceModeAPI
}

// IsSelfHosted returns true if using self-hosted inference
func (c *Config) IsSelfHosted() bool {
	return c.Mode == InferenceModeSelfHosted
}
