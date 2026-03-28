package runpod

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// InstanceState represents the lifecycle state of a GPU instance
type InstanceState string

const (
	InstanceStatePending    InstanceState = "pending"
	InstanceStateStarting   InstanceState = "starting"
	InstanceStateRunning    InstanceState = "running"
	InstanceStateIdle       InstanceState = "idle"
	InstanceStateStopping   InstanceState = "stopping"
	InstanceStateFailed     InstanceState = "failed"
	InstanceStateTerminated InstanceState = "terminated"
)

// GPUInstance represents a managed GPU instance
type GPUInstance struct {
	ID           string
	Name         string
	PodID        string
	State        InstanceState
	Endpoint     string
	Region       string    // Region identifier (e.g., "us-east-1", "eu-west-1")
	GPUType      string    // GPU type (e.g., "NVIDIA A100", "NVIDIA RTX A5000")
	ClusterID    string    // Cluster identifier this instance belongs to
	LastUsed     time.Time
	CreatedAt    time.Time
	RequestCount int
	mu           sync.RWMutex
}

// InstancePool manages a pool of GPU instances
type InstancePool struct {
	config    *Config
	client    *RunPodClient
	instances map[string]*GPUInstance
	region    string // Region this pool manages (empty means all regions)
	clusterID string // Cluster ID this pool belongs to
	mu        sync.RWMutex

	// Callbacks
	onInstanceReady  func(*GPUInstance)
	onInstanceFailed func(*GPUInstance, error)
}

// NewInstancePool creates a new GPU instance pool
func NewInstancePool(config *Config, client *RunPodClient) *InstancePool {
	return &InstancePool{
		config:    config,
		client:    client,
		instances: make(map[string]*GPUInstance),
	}
}

// NewRegionalPool creates a new GPU instance pool for a specific region
func NewRegionalPool(config *Config, client *RunPodClient, region, clusterID string) *InstancePool {
	return &InstancePool{
		config:    config,
		client:    client,
		instances: make(map[string]*GPUInstance),
		region:    region,
		clusterID: clusterID,
	}
}

// Provision provisions a new GPU instance
func (p *InstancePool) Provision(ctx context.Context) (*GPUInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we have an idle instance
	for _, inst := range p.instances {
		if inst.State == InstanceStateIdle {
			inst.mu.Lock()
			inst.State = InstanceStateRunning
			inst.LastUsed = time.Now()
			inst.RequestCount++
			inst.mu.Unlock()
			return inst, nil
		}
	}

	// Check if we've reached max instances
	if p.config.MaxInstances > 0 && len(p.instances) >= p.config.MaxInstances {
		// Wait for an instance to become available
		return nil, fmt.Errorf("max instances reached, waiting for available instance")
	}

	// Create a new instance
	instance := &GPUInstance{
		ID:        fmt.Sprintf("inst-%d", time.Now().UnixNano()),
		Name:      fmt.Sprintf("fnswarm-gpu-%d", time.Now().UnixNano()),
		State:     InstanceStatePending,
		Region:    p.region,
		ClusterID: p.clusterID,
		GPUType:   p.config.GPUType,
		CreatedAt: time.Now(),
	}

	// Add to pool immediately
	p.instances[instance.ID] = instance

	// Provision in background
	go func() {
		err := p.provisionInstance(ctx, instance)
		if err != nil {
			instance.mu.Lock()
			instance.State = InstanceStateFailed
			instance.mu.Unlock()
			if p.onInstanceFailed != nil {
				p.onInstanceFailed(instance, err)
			}
		}
	}()

	return instance, nil
}

// provisionInstance performs the actual provisioning
func (p *InstancePool) provisionInstance(ctx context.Context, instance *GPUInstance) error {
	spec := &PodSpec{
		Name:            instance.Name,
		ContainerImage:  p.config.ContainerImage,
		GPUType:         p.config.GPUType,
		GPUCount:        p.config.GPUCount,
		HTTPHost:        p.config.HTTPHost,
		HTTPPort:        p.config.HTTPPort,
		HealthCheckPath: p.config.HealthCheckPath,
		ModelName:       p.config.ModelName,
	}

	// Update state
	instance.mu.Lock()
	instance.State = InstanceStateStarting
	instance.mu.Unlock()

	// Create the pod
	pod, err := p.client.CreatePod(ctx, spec)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	instance.mu.Lock()
	instance.PodID = pod.ID
	instance.mu.Unlock()

	// Wait for pod to be ready
	readyPod, err := p.client.WaitForPodReady(ctx, pod.ID, p.config.ProvisioningTimeout)
	if err != nil {
		return fmt.Errorf("failed to wait for pod: %w", err)
	}

	// Update instance with endpoint
	instance.mu.Lock()
	instance.State = InstanceStateRunning
	instance.Endpoint = fmt.Sprintf("http://%s", readyPod.ContainerURL)
	instance.LastUsed = time.Now()
	instance.RequestCount = 1
	instance.mu.Unlock()

	log.Printf("GPU instance %s ready at %s", instance.ID, instance.Endpoint)

	// Trigger callback
	if p.onInstanceReady != nil {
		p.onInstanceReady(instance)
	}

	return nil
}

// Release releases an instance back to the pool (marks as idle)
func (p *InstancePool) Release(instanceID string) {
	p.mu.RLock()
	instance, ok := p.instances[instanceID]
	p.mu.RUnlock()

	if !ok {
		return
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.State == InstanceStateRunning {
		instance.State = InstanceStateIdle
		instance.LastUsed = time.Now()
	}
}

// Terminate terminates an instance and removes it from the pool
func (p *InstancePool) Terminate(ctx context.Context, instanceID string) error {
	p.mu.Lock()
	instance, ok := p.instances[instanceID]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("instance not found")
	}
	delete(p.instances, instanceID)
	p.mu.Unlock()

	instance.mu.Lock()
	instance.State = InstanceStateStopping
	instance.mu.Unlock()

	if instance.PodID != "" {
		if err := p.client.TerminatePod(ctx, instance.PodID); err != nil {
			log.Printf("Warning: failed to terminate pod %s: %v", instance.PodID, err)
		}
	}

	instance.mu.Lock()
	instance.State = InstanceStateTerminated
	instance.mu.Unlock()

	return nil
}

// CleanupIdleInstances terminates idle instances beyond the minimum
func (p *InstancePool) CleanupIdleInstances(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.config.MinInstances > 0 {
		return nil // Always keep minimum instances
	}

	var idleInstances []*GPUInstance
	for _, inst := range p.instances {
		inst.mu.RLock()
		if inst.State == InstanceStateIdle {
			idleInstances = append(idleInstances, inst)
		}
		inst.mu.RUnlock()
	}

	// Terminate idle instances beyond minimum
	toTerminate := len(idleInstances) - p.config.MinInstances
	if toTerminate <= 0 {
		return nil
	}

	for i := 0; i < toTerminate; i++ {
		inst := idleInstances[i]
		delete(p.instances, inst.ID)

		if inst.PodID != "" {
			if err := p.client.TerminatePod(ctx, inst.PodID); err != nil {
				log.Printf("Warning: failed to terminate idle pod %s: %v", inst.PodID, err)
			}
		}
		inst.mu.Lock()
		inst.State = InstanceStateTerminated
		inst.mu.Unlock()
	}

	return nil
}

// StartIdleMonitor starts a background goroutine to monitor idle instances
func (p *InstancePool) StartIdleMonitor(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.cleanupIdleInstances(ctx)
			}
		}
	}()
}

// cleanupIdleInstances cleans up instances that have been idle too long
func (p *InstancePool) cleanupIdleInstances(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for _, inst := range p.instances {
		inst.mu.RLock()
		isIdle := inst.State == InstanceStateIdle
		idleDuration := now.Sub(inst.LastUsed)
		inst.mu.RUnlock()

		if isIdle && idleDuration > p.config.IdleTimeout {
			// Check if we should keep minimum instances
			if p.config.MinInstances > 0 {
				runningCount := 0
				for _, i := range p.instances {
					i.mu.RLock()
					if i.State == InstanceStateRunning || i.State == InstanceStateIdle {
						runningCount++
					}
					i.mu.RUnlock()
				}
				if runningCount <= p.config.MinInstances {
					continue
				}
			}

			delete(p.instances, inst.ID)
			if inst.PodID != "" {
				go func(podID string) {
					if err := p.client.TerminatePod(ctx, podID); err != nil {
						log.Printf("Warning: failed to terminate idle pod %s: %v", podID, err)
					}
				}(inst.PodID)
			}
			inst.mu.Lock()
			inst.State = InstanceStateTerminated
			inst.mu.Unlock()
		}
	}
}

// GetInstance returns an instance by ID
func (p *InstancePool) GetInstance(instanceID string) (*GPUInstance, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	inst, ok := p.instances[instanceID]
	return inst, ok
}

// ListInstances returns all instances
func (p *InstancePool) ListInstances() []*GPUInstance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	instances := make([]*GPUInstance, 0, len(p.instances))
	for _, inst := range p.instances {
		inst.mu.RLock()
		instances = append(instances, &GPUInstance{
			ID:           inst.ID,
			Name:         inst.Name,
			PodID:        inst.PodID,
			State:        inst.State,
			Endpoint:     inst.Endpoint,
			Region:       inst.Region,
			GPUType:      inst.GPUType,
			ClusterID:    inst.ClusterID,
			LastUsed:     inst.LastUsed,
			CreatedAt:    inst.CreatedAt,
			RequestCount: inst.RequestCount,
		})
		inst.mu.RUnlock()
	}
	return instances
}

// ListInstancesByRegion returns all instances in a specific region
func (p *InstancePool) ListInstancesByRegion(region string) []*GPUInstance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	instances := make([]*GPUInstance, 0)
	for _, inst := range p.instances {
		inst.mu.RLock()
		if inst.Region == region {
			instances = append(instances, &GPUInstance{
				ID:           inst.ID,
				Name:         inst.Name,
				PodID:        inst.PodID,
				State:        inst.State,
				Endpoint:     inst.Endpoint,
				Region:       inst.Region,
				GPUType:      inst.GPUType,
				ClusterID:    inst.ClusterID,
				LastUsed:     inst.LastUsed,
				CreatedAt:    inst.CreatedAt,
				RequestCount: inst.RequestCount,
			})
		}
		inst.mu.RUnlock()
	}
	return instances
}

// GetIdleInstance returns an idle instance, optionally filtered by region
func (p *InstancePool) GetIdleInstance(region string) (*GPUInstance, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// First pass: find idle instance in specified region (if any)
	if region != "" {
		for _, inst := range p.instances {
			inst.mu.RLock()
			if inst.State == InstanceStateIdle && inst.Region == region {
				inst.mu.RUnlock()
				return inst, true
			}
			inst.mu.RUnlock()
		}
	}

	// Second pass: find any idle instance
	for _, inst := range p.instances {
		inst.mu.RLock()
		if inst.State == InstanceStateIdle {
			inst.mu.RUnlock()
			return inst, true
		}
		inst.mu.RUnlock()
	}
	return nil, false
}

// Region returns the region this pool manages (empty means all regions)
func (p *InstancePool) Region() string {
	return p.region
}

// ClusterID returns the cluster ID this pool belongs to
func (p *InstancePool) ClusterID() string {
	return p.clusterID
}

// GetRegionStats returns statistics for a specific region
func (p *InstancePool) GetRegionStats(region string) (total, running, idle, failed int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, inst := range p.instances {
		if region != "" && inst.Region != region {
			continue
		}
		inst.mu.RLock()
		total++
		switch inst.State {
		case InstanceStateRunning:
			running++
		case InstanceStateIdle:
			idle++
		case InstanceStateFailed:
			failed++
		}
		inst.mu.RUnlock()
	}
	return
}

// GetStats returns pool statistics
func (p *InstancePool) GetStats() (total, running, idle, failed int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, inst := range p.instances {
		inst.mu.RLock()
		switch inst.State {
		case InstanceStateRunning:
			running++
		case InstanceStateIdle:
			idle++
		case InstanceStateFailed:
			failed++
		}
		total++
	}
	return
}

// SetOnInstanceReady sets the callback for when an instance becomes ready
func (p *InstancePool) SetOnInstanceReady(cb func(*GPUInstance)) {
	p.onInstanceReady = cb
}

// SetOnInstanceFailed sets the callback for when an instance fails
func (p *InstancePool) SetOnInstanceFailed(cb func(*GPUInstance, error)) {
	p.onInstanceFailed = cb
}
