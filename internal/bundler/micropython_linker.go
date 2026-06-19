package bundler

import (
	"fmt"
	"os"

	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/sirupsen/logrus"
)

// ProductionMicroPythonLinker provides production-ready linking of Python code with MicroPython runtime
// This implementation uses micropython.wasm directly - user code is loaded at runtime via mp_js_do_exec
type ProductionMicroPythonLinker struct {
	runtimePath string
	userCode    string
	manifest    *manifest.Manifest
}

// NewProductionMicroPythonLinker creates a production linker instance
func NewProductionMicroPythonLinker(userCode string, m *manifest.Manifest) *ProductionMicroPythonLinker {
	return &ProductionMicroPythonLinker{
		runtimePath: findMicropythonRuntimePath(),
		userCode:    userCode,
		manifest:    m,
	}
}


// Link returns the micropython.wasm directly - user code is loaded at runtime via mp_js_do_exec
// This is the correct approach: micropython is an interpreter that receives code at runtime
func (l *ProductionMicroPythonLinker) Link() ([]byte, error) {
	if l.runtimePath == "" {
		return nil, fmt.Errorf("micropython runtime not found")
	}

	// Read and return micropython.wasm as-is
	// User Python code will be loaded at runtime via mp_js_do_exec
	runtimeBytes, err := os.ReadFile(l.runtimePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read micropython runtime: %v", err)
	}

	logrus.WithField("runtime_bytes", len(runtimeBytes)).Debug("Using micropython runtime")
	logrus.Debug("User code will be loaded at runtime via mp_js_do_exec")

	return runtimeBytes, nil
}

// GetUserCode returns the user Python code that will be loaded at runtime
func (l *ProductionMicroPythonLinker) GetUserCode() string {
	return l.userCode
}

// CompileWithMicropython returns the micropython.wasm for runtime execution
// The user code is stored separately and loaded at runtime via mp_js_do_exec
func CompileWithMicropython(sourceCode string, m *manifest.Manifest) ([]byte, error) {
	logrus.WithField("name", m.Name).Debug("Preparing MicroPython runtime")

	linker := NewProductionMicroPythonLinker(sourceCode, m)
	wasm, err := linker.Link()
	if err != nil {
		return nil, fmt.Errorf("MicroPython preparation failed: %v", err)
	}

	// Validate WASM
	if len(wasm) < 1000 {
		return nil, fmt.Errorf("WASM too small: %d bytes", len(wasm))
	}

	if wasm[0] != 0x00 || wasm[1] != 0x61 || wasm[2] != 0x73 || wasm[3] != 0x6D {
		return nil, fmt.Errorf("invalid WASM magic")
	}

	logrus.WithField("wasm_bytes", len(wasm)).Debug("MicroPython runtime ready")
	return wasm, nil
}

// GetMicropythonRuntimePath returns the path to the micropython runtime
func GetMicropythonRuntimePath() string {
	return findMicropythonRuntimePath()
}

// IsMicropythonAvailable checks if micropython runtime is available
func IsMicropythonAvailable() bool {
	return findMicropythonRuntimePath() != ""
}
