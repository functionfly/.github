package testing

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/sirupsen/logrus"
)

// RuntimeResult represents the result of a runtime execution
type RuntimeResult struct {
	Status    int
	Output    string
	Error     string
	LatencyMs int
}

// Runtime defines the interface for function runtimes
type Runtime interface {
	Initialize(ctx context.Context, bundle []byte, manifest *manifest.Manifest) error
	Execute(ctx context.Context, input string) (*RuntimeResult, error)
	Cleanup() error
}

// RuntimeRegistry holds available runtimes
var runtimeRegistry = make(map[string]Runtime)

// RegisterRuntime registers a runtime implementation
func RegisterRuntime(name string, runtime Runtime) {
	runtimeRegistry[name] = runtime
}

// GetRuntime returns the runtime for the given name
func GetRuntime(name string) Runtime {
	runtime, exists := runtimeRegistry[name]
	if !exists {
		// Return a generic runtime as fallback
		return &GenericRuntime{}
	}
	return runtime
}

// GenericRuntime provides a basic runtime implementation for testing
type GenericRuntime struct {
	manifest *manifest.Manifest
	bundle   []byte
}

func (r *GenericRuntime) Initialize(ctx context.Context, bundle []byte, manifest *manifest.Manifest) error {
	r.bundle = bundle
	r.manifest = manifest
	if manifest != nil {
		logrus.WithFields(logrus.Fields{
			"runtime": manifest.Runtime,
			"version": manifest.Version,
		}).Debug("GenericRuntime initialized")
	}
	return nil
}

func (r *GenericRuntime) Execute(ctx context.Context, input string) (*RuntimeResult, error) {
	start := time.Now()

	runtime := "generic"
	if r.manifest != nil {
		runtime = r.manifest.Runtime
	}

	// Simulate function execution
	output := fmt.Sprintf("Processed: %s", input)

	latency := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"runtime":    runtime,
		"latency_ms": latency.Milliseconds(),
	}).Debug("GenericRuntime execution completed")

	return &RuntimeResult{
		Status:    200,
		Output:    output,
		LatencyMs: int(latency.Milliseconds()),
	}, nil
}

func (r *GenericRuntime) Cleanup() error {
	if r.manifest != nil {
		logrus.WithField("runtime", r.manifest.Runtime).Debug("GenericRuntime cleaned up")
	}
	r.bundle = nil
	r.manifest = nil
	return nil
}

// Initialize runtimes
func init() {
	RegisterRuntime("node", &GenericRuntime{})
	RegisterRuntime("python", &GenericRuntime{})
	RegisterRuntime("go", &GenericRuntime{})
	RegisterRuntime("generic", &GenericRuntime{})
}
