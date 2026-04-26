// toolchain_detector.go detects available WASM compilation toolchains for each supported language.
package bundler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ToolchainInfo describes an available compilation toolchain.
type ToolchainInfo struct {
	Name      string `json:"name"`
	Language  string `json:"language"`
	Available bool   `json:"available"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
	Toolchain string `json:"toolchain,omitempty"` // e.g. "emscripten", "wasi-sdk", "cargo", "go"
}

// DetectToolchains checks which WASM compilation toolchains are available on the system.
func DetectToolchains() map[string]ToolchainInfo {
	result := make(map[string]ToolchainInfo)

	result["rust"] = detectRustToolchain()
	result["go"] = detectGoToolchain()
	result["c"] = detectCToolchain()
	result["cpp"] = detectCppToolchain()
	result["ruby"] = detectRubyToolchain()
	result["kotlin"] = detectKotlinToolchain()
	result["swift"] = detectSwiftToolchain()
	result["javascript"] = detectJavaScriptToolchain()
	result["python"] = detectPythonToolchain()

	return result
}

// DetectAvailableRuntimes returns the list of runtime IDs that have working toolchains.
func DetectAvailableRuntimes() []string {
	toolchains := DetectToolchains()
	var available []string

	for lang, info := range toolchains {
		if info.Available {
			available = append(available, lang)
		}
	}

	return available
}

func detectRustToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "Rust (cargo)", Language: "rust"}

	cargo, err := exec.LookPath("cargo")
	if err != nil {
		return info
	}
	info.Available = true
	info.Path = cargo
	info.Toolchain = "cargo"

	cmd := exec.Command("cargo", "version")
	if out, err := cmd.Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}

	return info
}

func detectGoToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "Go", Language: "go"}

	gobin, err := exec.LookPath("go")
	if err != nil {
		return info
	}
	info.Available = true
	info.Path = gobin
	info.Toolchain = "go"

	cmd := exec.Command("go", "version")
	if out, err := cmd.Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}

	return info
}

func detectCToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "C (WASM)", Language: "c"}

	// Check WASI-SDK
	if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
		clang := filepath.Join(sdkPath, "bin", "clang")
		if _, err := os.Stat(clang); err == nil {
			info.Available = true
			info.Path = clang
			info.Toolchain = "wasi-sdk"
			info.Version = "WASI-SDK at " + sdkPath
			return info
		}
	}

	for _, p := range []string{"/opt/wasi-sdk/bin/clang", "/usr/local/opt/wasi-sdk/bin/clang"} {
		if _, err := os.Stat(p); err == nil {
			info.Available = true
			info.Path = p
			info.Toolchain = "wasi-sdk"
			info.Version = "WASI-SDK"
			return info
		}
	}

	// Check Emscripten
	if emcc := findEmscriptenCompiler(); emcc != "" {
		info.Available = true
		info.Path = emcc
		info.Toolchain = "emscripten"

		cmd := exec.Command(emcc, "--version")
		if out, err := cmd.Output(); err == nil {
			info.Version = strings.SplitN(string(out), "\n", 2)[0]
		}
	}

	return info
}

func detectCppToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "C++ (WASM)", Language: "cpp"}

	if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
		clangxx := filepath.Join(sdkPath, "bin", "clang++")
		if _, err := os.Stat(clangxx); err == nil {
			info.Available = true
			info.Path = clangxx
			info.Toolchain = "wasi-sdk"
			info.Version = "WASI-SDK at " + sdkPath
			return info
		}
	}

	for _, p := range []string{"/opt/wasi-sdk/bin/clang++", "/usr/local/opt/wasi-sdk/bin/clang++"} {
		if _, err := os.Stat(p); err == nil {
			info.Available = true
			info.Path = p
			info.Toolchain = "wasi-sdk"
			info.Version = "WASI-SDK"
			return info
		}
	}

	if emcc := findEmscriptenCompiler(); emcc != "" {
		info.Available = true
		info.Path = emcc
		info.Toolchain = "emscripten"

		cmd := exec.Command(emcc, "--version")
		if out, err := cmd.Output(); err == nil {
			info.Version = strings.SplitN(string(out), "\n", 2)[0]
		}
	}

	return info
}

func detectRubyToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "Ruby (mruby WASM)", Language: "ruby"}

	// Check for pre-compiled mruby.wasm
	candidates := []string{
		mrubyWASMPath,
		filepath.Join("..", mrubyWASMPath),
		filepath.Join("..", "..", mrubyWASMPath),
	}
	if envPath := os.Getenv("MRUBY_WASM_PATH"); envPath != "" {
		candidates = append([]string{envPath}, candidates...)
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			info.Available = true
			info.Path = path
			info.Toolchain = "mruby-wasm"
			info.Version = "mruby (pre-compiled WASM)"
			return info
		}
	}

	return info
}

func detectKotlinToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "Kotlin (WASM)", Language: "kotlin"}

	if kotlin, err := exec.LookPath("kotlin"); err == nil {
		info.Available = true
		info.Path = kotlin
		info.Toolchain = "kotlin-wasm"

		cmd := exec.Command("kotlin", "-version")
		if out, err := cmd.Output(); err == nil {
			info.Version = strings.TrimSpace(string(out))
		}
	} else if kotlinc, err := exec.LookPath("kotlinc"); err == nil {
		info.Available = true
		info.Path = kotlinc
		info.Toolchain = "kotlin-js"

		cmd := exec.Command("kotlinc", "-version")
		if out, err := cmd.Output(); err == nil {
			info.Version = strings.TrimSpace(string(out))
		}
	}

	return info
}

func detectSwiftToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "Swift (WASM)", Language: "swift"}

	if carton, err := exec.LookPath("carton"); err == nil {
		info.Available = true
		info.Path = carton
		info.Toolchain = "carton"
		info.Version = "carton (SwiftWasm)"
		return info
	}

	if swiftc, err := exec.LookPath("swiftc"); err == nil {
		info.Available = true
		info.Path = swiftc
		info.Toolchain = "swiftwasm"

		cmd := exec.Command("swiftc", "-version")
		if out, err := cmd.Output(); err == nil {
			info.Version = strings.TrimSpace(string(out))
		}
	}

	return info
}

func detectJavaScriptToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "JavaScript (Javy/QuickJS)", Language: "javascript"}

	if javy, err := exec.LookPath("javy"); err == nil {
		info.Available = true
		info.Path = javy
		info.Toolchain = "javy"

		cmd := exec.Command("javy", "--version")
		if out, err := cmd.Output(); err == nil {
			info.Version = strings.TrimSpace(string(out))
		}
		return info
	}

	// Javy is bundled / may be found at a known path
	info.Available = true
	info.Toolchain = "javy-bundled"
	info.Version = "bundled (may need install)"
	return info
}

func detectPythonToolchain() ToolchainInfo {
	info := ToolchainInfo{Name: "Python (MicroPython WASM)", Language: "python"}

	// Check for micropython.wasm
	candidates := []string{
		"bundler/python/micropython.wasm",
		filepath.Join("..", "bundler/python/micropython.wasm"),
		filepath.Join("..", "..", "bundler/python/micropython.wasm"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			info.Available = true
			info.Path = path
			info.Toolchain = "micropython"
			info.Version = "MicroPython (pre-compiled WASM)"
			return info
		}
	}

	return info
}
