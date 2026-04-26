// kotlin_wasm_compiler.go implements Kotlin → WASM compilation.
// Primary: Kotlin/WASM direct compilation via Gradle with wasmWasi target.
// Fallback: Kotlin → JS → Javy → WASM.
package bundler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/functionfly/functionfly/internal/manifest"
)

const (
	kotlinBuildTimeout = 120 * time.Second
	kotlinMainTemplate = `package functionfly

fun handler(input: String): String {
%s
}
`
)

// bundleKotlinForWasmRuntime compiles Kotlin source to WASM.
func bundleKotlinForWasmRuntime(m *manifest.Manifest) ([]byte, error) {
	_, src, err := ReadEntryFile(m)
	if err != nil {
		return nil, NewBundlerErrorWithCause("kotlin bundle", "failed to read entry file", err)
	}

	// Try Kotlin/WASM direct compilation first
	wasm, err := compileKotlinWasmDirect(src, m)
	if err == nil {
		return wasm, nil
	}

	fmt.Printf("Warning: Kotlin/WASM direct compilation failed (%v), trying Kotlin→JS→Javy fallback\n", err)

	// Fallback: Kotlin → JS → Javy → WASM
	return compileKotlinViaJSFallback(src, m)
}

// compileKotlinWasmDirect uses Kotlin/WASM (wasmWasi target) for direct compilation.
func compileKotlinWasmDirect(src []byte, m *manifest.Manifest) ([]byte, error) {
	if _, err := exec.LookPath("kotlin"); err != nil {
		if _, err2 := exec.LookPath("kotlinc"); err2 != nil {
			return nil, fmt.Errorf("kotlin compiler not found; install Kotlin 1.9+ from https://kotlinlang.org/docs/wasm-getting-started.html")
		}
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-kotlin-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("kotlin bundle", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Gradle project structure for Kotlin/WASM
	if err := createKotlinWasmProject(tmpDir, src, m); err != nil {
		return nil, NewBundlerErrorWithCause("kotlin bundle", "failed to create project", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), kotlinBuildTimeout)
	defer cancel()

	// Build with Gradle
	buildCmd := exec.CommandContext(ctx, "./gradlew", "wasmWasiNodeProductionRun", "--no-daemon")
	buildCmd.Dir = tmpDir
	buildCmd.Env = os.Environ()

	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("gradle", "kotlin wasm build", string(out), err)
	}

	// Find the output .wasm file
	wasmBytes, err := findKotlinWasmOutput(tmpDir)
	if err != nil {
		return nil, NewBundlerErrorWithCause("kotlin bundle", "failed to find output WASM", err)
	}

	if valErr := validateWasmModule(wasmBytes); valErr != nil {
		return nil, NewBundlerErrorWithCause("kotlin bundle", "output WASM validation failed", valErr)
	}

	return wasmBytes, nil
}

// createKotlinWasmProject scaffolds a Kotlin/WASM Gradle project.
func createKotlinWasmProject(dir string, src []byte, m *manifest.Manifest) error {
	// settings.gradle.kts
	settings := `rootProject.name = "functionfly-kotlin"
pluginManagement {
    repositories {
        gradlePluginPortal()
        mavenCentral()
        maven("https://maven.pkg.jetbrains.space/kotlin/p/wasm/experimental")
    }
}
`
	if err := os.WriteFile(filepath.Join(dir, "settings.gradle.kts"), []byte(settings), 0644); err != nil {
		return err
	}

	// build.gradle.kts
	buildGradle := fmt.Sprintf(`plugins {
    kotlin("multiplatform") version "1.9.24"
}

repositories {
    mavenCentral()
    maven("https://maven.pkg.jetbrains.space/kotlin/p/wasm/experimental")
}

kotlin {
    wasmWasi {
        binaries.executable()
        nodejs()
    }
    sourceSets {
        val wasmWasiMain by getting {
            dependencies {
                implementation(kotlin("stdlib"))
            }
        }
    }
}
`)
	if err := os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte(buildGradle), 0644); err != nil {
		return err
	}

	// Gradle wrapper (use gradlew if available, otherwise use system gradle)
	if err := createGradleWrapper(dir); err != nil {
		// Not fatal - caller may have gradle on PATH
		fmt.Printf("Warning: could not create Gradle wrapper: %v\n", err)
	}

	// Source directory
	srcDir := filepath.Join(dir, "src", "wasmWasiMain", "kotlin")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	// Write user source as Main.kt
	mainFile := filepath.Join(srcDir, "Main.kt")
	if err := os.WriteFile(mainFile, src, 0644); err != nil {
		return err
	}

	return nil
}

// createGradleWrapper creates a minimal Gradle wrapper.
func createGradleWrapper(dir string) error {
	wrapperDir := filepath.Join(dir, "gradle", "wrapper")
	if err := os.MkdirAll(wrapperDir, 0755); err != nil {
		return err
	}

	properties := `distributionBase=GRADLE_USER_HOME
distributionPath=wrapper/dists
distributionUrl=https\://services.gradle.org/distributions/gradle-8.5-bin.zip
networkTimeout=10000
zipStoreBase=GRADLE_USER_HOME
zipStorePath=wrapper/dists
`
	if err := os.WriteFile(filepath.Join(wrapperDir, "gradle-wrapper.properties"), []byte(properties), 0644); err != nil {
		return err
	}

	// Try to generate wrapper using system gradle
	if gradlePath, err := exec.LookPath("gradle"); err == nil {
		cmd := exec.Command(gradlePath, "wrapper")
		cmd.Dir = dir
		_ = cmd.Run()
	}

	return nil
}

// findKotlinWasmOutput locates the compiled .wasm file from Gradle build output.
func findKotlinWasmOutput(projectDir string) ([]byte, error) {
	// Walk the build directory tree looking for WASM files
	var found []byte
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".wasm" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if len(data) >= 8 && data[0] == 0x00 && data[1] == 0x61 && data[2] == 0x73 && data[3] == 0x6D {
			found = data
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found != nil {
		return found, nil
	}
	return nil, fmt.Errorf("no .wasm file found in build output")
}

// compileKotlinViaJSFallback compiles Kotlin source to JS, then to WASM via Javy.
func compileKotlinViaJSFallback(src []byte, m *manifest.Manifest) ([]byte, error) {
	kotlinc, err := exec.LookPath("kotlinc-js")
	if err != nil {
		kotlinc, err = exec.LookPath("kotlinc")
		if err != nil {
			return nil, fmt.Errorf("kotlin compiler not found; install Kotlin 1.9+ from https://kotlinlang.org")
		}
	}

	tmpDir, err := os.MkdirTemp("", "functionfly-kotlin-js-*")
	if err != nil {
		return nil, NewBundlerErrorWithCause("kotlin js fallback", "failed to create temp dir", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	srcFile := filepath.Join(tmpDir, "Main.kt")
	if err := os.WriteFile(srcFile, src, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("kotlin js fallback", "failed to write source", err)
	}

	outDir := filepath.Join(tmpDir, "out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, NewBundlerErrorWithCause("kotlin js fallback", "failed to create output dir", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), kotlinBuildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, kotlinc, srcFile, "-output", outDir, "-target", "v5")
	cmd.Dir = tmpDir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, NewCompilationErrorWithOutput("kotlinc-js", srcFile, string(out), err)
	}

	// Find the generated JS file
	jsBytes, err := findKotlinJSOutput(outDir)
	if err != nil {
		return nil, NewBundlerErrorWithCause("kotlin js fallback", "failed to find JS output", err)
	}

	// Compile JS → WASM via Javy (reuse existing JS→WASM pipeline)
	jsManifest := &manifest.Manifest{
		Name:    m.Name,
		Version: m.Version,
		Runtime: "node18",
	}

	// Write JS to a temp file and set it as entry
	jsFile := filepath.Join(tmpDir, "index.js")
	if err := os.WriteFile(jsFile, jsBytes, 0644); err != nil {
		return nil, NewBundlerErrorWithCause("kotlin js fallback", "failed to write JS", err)
	}
	jsManifest.Entry = "index.js"

	return bundleJSForWasmRuntime(jsManifest)
}

// findKotlinJSOutput locates the main .js file from Kotlin/JS compilation.
func findKotlinJSOutput(outDir string) ([]byte, error) {
	matches, err := filepath.Glob(filepath.Join(outDir, "*.js"))
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("no .js files found in %s", outDir)
	}

	// Look for the main JS file (usually the project name)
	for _, match := range matches {
		base := filepath.Base(match)
		if base != "kotlin.js" && base != "kotlin-test.js" {
			return os.ReadFile(match)
		}
	}

	// Fallback: return the first non-stdlib JS file
	return os.ReadFile(matches[0])
}
