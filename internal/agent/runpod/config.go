package runpod

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// InferenceMode defines the inference mode to use
type InferenceMode string

const (
	// InferenceModeAPI uses OpenRouter API-based inference (default)
	InferenceModeAPI InferenceMode = "api"
	// InferenceModeSelfHosted uses self-hosted RunPod GPU inference
	InferenceModeSelfHosted InferenceMode = "self_hosted"
	// InferenceModeCluster uses multi-region cluster mode with load balancing
	InferenceModeCluster InferenceMode = "cluster"
)

// RegionConfig holds configuration for a specific region
type RegionConfig struct {
	Region         string            // Region identifier (e.g., "us-east-1", "eu-west-1")
	GPUType        string            // GPU type for this region (e.g., "NVIDIA A100")
	GPUCount       int               // Number of GPUs per instance in this region
	MaxInstances   int               // Maximum instances in this region
	MinInstances   int               // Minimum instances to maintain in this region
	ContainerImage string            // Docker image for this region
	BaseURL        string            // RunPod API base URL for this region (defaults to global)
	Weight         int               // Weight for load balancing (higher = more traffic)
	Enabled        bool              // Whether this region is enabled
}

// Config holds all RunPod configuration options
type Config struct {
	// Mode determines whether to use API-based or self-hosted inference
	Mode InferenceMode

	// RunPod API Configuration
	RunPodAPIKey     string
	RunPodAPIBaseURL string

	// GPU Instance Configuration (used in single-region mode)
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

	// Multi-Region Configuration (used in cluster mode)
	Region           string                   // Primary region (default: "us-east-1")
	Regions          []RegionConfig           // Slice of regional configurations
	PreferredRegion  string                   // Preferred region for geo-aware routing
	RegionWeights    map[string]int          // Traffic distribution weights by region
	ClusterMode      bool                    // Whether to use cluster mode with multiple regions
	MinInstancesPerRegion int                 // Minimum instances per region in cluster mode
	MaxInstancesPerRegion int                 // Maximum instances per region in cluster mode
	ScaleUpThresholdPerRegion int              // Scale up threshold per region
	ScaleDownThresholdPerRegion int            // Scale down threshold per region
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
		// Multi-region defaults
		Region:                     "us-east-1",
		PreferredRegion:            "us-east-1",
		ClusterMode:                false,
		MinInstancesPerRegion:      0,
		MaxInstancesPerRegion:      10,
		ScaleUpThresholdPerRegion:  50,
		ScaleDownThresholdPerRegion: 5,
		RegionWeights:              make(map[string]int),
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

	// Load multi-region configuration
	if region := os.Getenv("RUNPOD_REGION"); region != "" {
		cfg.Region = region
	}
	if preferredRegion := os.Getenv("RUNPOD_PREFERRED_REGION"); preferredRegion != "" {
		cfg.PreferredRegion = preferredRegion
	}
	if clusterMode := os.Getenv("RUNPOD_CLUSTER_MODE"); clusterMode == "true" || clusterMode == "1" {
		cfg.ClusterMode = true
		cfg.Mode = InferenceModeCluster
	}

	// Load per-region scaling configuration
	if minPerRegion := os.Getenv("RUNPOD_MIN_INSTANCES_PER_REGION"); minPerRegion != "" {
		if count, err := strconv.Atoi(minPerRegion); err == nil && count >= 0 {
			cfg.MinInstancesPerRegion = count
		}
	}
	if maxPerRegion := os.Getenv("RUNPOD_MAX_INSTANCES_PER_REGION"); maxPerRegion != "" {
		if count, err := strconv.Atoi(maxPerRegion); err == nil && count > 0 {
			cfg.MaxInstancesPerRegion = count
		}
	}
	if scaleUpRegion := os.Getenv("RUNPOD_SCALE_UP_THRESHOLD_PER_REGION"); scaleUpRegion != "" {
		if count, err := strconv.Atoi(scaleUpRegion); err == nil && count > 0 {
			cfg.ScaleUpThresholdPerRegion = count
		}
	}
	if scaleDownRegion := os.Getenv("RUNPOD_SCALE_DOWN_THRESHOLD_PER_REGION"); scaleDownRegion != "" {
		if count, err := strconv.Atoi(scaleDownRegion); err == nil && count > 0 {
			cfg.ScaleDownThresholdPerRegion = count
		}
	}

	// Load region weights (comma-separated: us-east-1:3,eu-west-1:2)
	if weights := os.Getenv("RUNPOD_REGION_WEIGHTS"); weights != "" {
		cfg.RegionWeights = parseRegionWeights(weights)
	}

	// Load multi-region configurations (RUNPOD_REGIONS as JSON array or semicolon-separated)
	if regions := os.Getenv("RUNPOD_REGIONS"); regions != "" {
		cfg.Regions = parseRegions(regions, cfg)
	}

	return cfg
}

// parseRegionWeights parses region weights from string format "region1:weight1,region2:weight2"
func parseRegionWeights(weights string) map[string]int {
	result := make(map[string]int)
	pairs := strings.Split(weights, ",")
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), ":")
		if len(kv) == 2 {
			if weight, err := strconv.Atoi(strings.TrimSpace(kv[1])); err == nil {
				result[strings.TrimSpace(kv[0])] = weight
			}
		}
	}
	return result
}

// parseRegions parses region configurations from environment string
func parseRegions(regions string, cfg *Config) []RegionConfig {
	var result []RegionConfig
	// Format: "region1:GPUType1:GPUCount1:Max1:Min1,region2:GPUType2:GPUCount2:Max2:Min2"
	regionEntries := strings.Split(regions, ",")
	for _, entry := range regionEntries {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) >= 1 && parts[0] != "" {
			rc := RegionConfig{
				Region:  strings.TrimSpace(parts[0]),
				Enabled: true,
			}
			// Set defaults then override with provided values
			rc.GPUType = cfg.GPUType
			rc.GPUCount = cfg.GPUCount
			rc.MaxInstances = cfg.MaxInstancesPerRegion
			rc.MinInstances = cfg.MinInstancesPerRegion
			rc.ContainerImage = cfg.ContainerImage
			rc.Weight = 1

			if len(parts) >= 2 && parts[1] != "" {
				rc.GPUType = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 {
				if count, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
					rc.GPUCount = count
				}
			}
			if len(parts) >= 4 {
				if max, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil {
					rc.MaxInstances = max
				}
			}
			if len(parts) >= 5 {
				if min, err := strconv.Atoi(strings.TrimSpace(parts[4])); err == nil {
					rc.MinInstances = min
				}
			}
			result = append(result, rc)
		}
	}
	return result
}

// IsAPIOnly returns true if using API-based inference only
func (c *Config) IsAPIOnly() bool {
	return c.Mode == InferenceModeAPI
}

// IsSelfHosted returns true if using self-hosted inference
func (c *Config) IsSelfHosted() bool {
	return c.Mode == InferenceModeSelfHosted
}

// IsClusterMode returns true if using cluster mode with multi-region support
func (c *Config) IsClusterMode() bool {
	return c.Mode == InferenceModeCluster || c.ClusterMode
}

// GetRegionConfig returns the RegionConfig for a specific region
func (c *Config) GetRegionConfig(region string) *RegionConfig {
	for i := range c.Regions {
		if c.Regions[i].Region == region {
			return &c.Regions[i]
		}
	}
	return nil
}

// GetEnabledRegions returns all enabled regions
func (c *Config) GetEnabledRegions() []RegionConfig {
	var enabled []RegionConfig
	for _, r := range c.Regions {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled
}

// GetRegionWeight returns the weight for a specific region (default: 1)
func (c *Config) GetRegionWeight(region string) int {
	if w, ok := c.RegionWeights[region]; ok {
		return w
	}
	return 1
}

// AddRegion adds or updates a region configuration
func (c *Config) AddRegion(region RegionConfig) {
	for i, r := range c.Regions {
		if r.Region == region.Region {
			c.Regions[i] = region
			return
		}
	}
	c.Regions = append(c.Regions, region)
}

// RemoveRegion removes a region configuration
func (c *Config) RemoveRegion(region string) {
	for i, r := range c.Regions {
		if r.Region == region {
			c.Regions = append(c.Regions[:i], c.Regions[i+1:]...)
			return
		}
	}
}

// DefaultRegions returns the default regional configuration
func DefaultRegions() []RegionConfig {
	return []RegionConfig{
		{
			Region:         "us-east-1",
			GPUType:        "NVIDIA A100",
			GPUCount:       1,
			MaxInstances:   16,
			MinInstances:   0,
			ContainerImage: "ghcr.io/huggingface/text-generation-inference:2.2.0",
			Weight:         3,
			Enabled:        true,
		},
		{
			Region:         "us-east-1",
			GPUType:        "NVIDIA RTX A5000",
			GPUCount:       1,
			MaxInstances:   32,
			MinInstances:   0,
			ContainerImage: "ghcr.io/huggingface/text-generation-inference:2.2.0",
			Weight:         2,
			Enabled:        true,
		},
		{
			Region:         "eu-west-1",
			GPUType:        "NVIDIA A100",
			GPUCount:       1,
			MaxInstances:   8,
			MinInstances:   0,
			ContainerImage: "ghcr.io/huggingface/text-generation-inference:2.2.0",
			Weight:         2,
			Enabled:        true,
		},
		{
			Region:         "ap-southeast-1",
			GPUType:        "NVIDIA A100",
			GPUCount:       1,
			MaxInstances:   8,
			MinInstances:   0,
			ContainerImage: "ghcr.io/huggingface/text-generation-inference:2.2.0",
			Weight:         1,
			Enabled:        true,
		},
	}
}
