package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/tools"
	"github.com/functionfly/functionfly/internal/agent/tools/search"
	"github.com/functionfly/functionfly/internal/agent/tools/search/providers"
	"github.com/functionfly/functionfly/internal/api"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func findEnvFile() string {
	candidates := []string{}
	exePath, err := os.Executable()
	if err == nil {
		candidates = append(candidates, filepath.Join(exePath, ".env"))
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), ".env"))
		candidates = append(candidates, filepath.Join(filepath.Dir(filepath.Dir(exePath)), ".env"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
		candidates = append(candidates, filepath.Join(cwd, "..", ".env"))
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func main() {
	// Load .env file if present (for local development)
	envFile := findEnvFile()
	if envFile != "" {
		if err := godotenv.Load(envFile); err == nil {
			logrus.WithField("path", envFile).Info("Loaded .env file")
		} else {
			logrus.WithError(err).WithField("path", envFile).Warn("Failed to load .env file")
		}
	}

	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	skipMigrations := flag.Bool("skip-migrations", false, "Skip database migrations")
	flag.Parse()

	// Prefer PORT from environment (e.g. Fly.io, 12-factor)
	if p := os.Getenv("PORT"); p != "" {
		var portVal int
		if n, err := fmt.Sscanf(p, "%d", &portVal); n == 1 && err == nil && portVal > 0 {
			*port = portVal
		}
	}
	if *port <= 0 {
		*port = 8080
	}

	// Configure logrus for structured JSON logging
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Check if running in development mode
	devMode := os.Getenv("DEVELOPMENT") == "true"
	if devMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Info("Running in development mode")
	}

	logrus.Info("Starting FunctionFly Orchestrator API")

	// Initialize database connection
	db, err := storage.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations if not skipped
	if !*skipMigrations {
		logrus.Info("Running database migrations...")
		if err := storage.RunMigrations(db); err != nil {
			logrus.WithError(err).Fatal("Failed to run migrations")
		}
		logrus.Info("Database migrations completed")
	} else {
		logrus.Info("Skipping database migrations")
	}

	// Initialize search tools for agents
	if err := initializeSearchTools(); err != nil {
		logrus.WithError(err).Warn("Failed to initialize search tools")
	} else {
		logrus.Info("Search tools initialized")
	}

	// Create the API server
	server := api.NewServer(db)

	// Start AI service (FlyMind) if not already running
	aiServicePID := startAIService()

	// ListenAndServe registers SIGINT/SIGTERM and runs graceful shutdown internally.
	// Do not also handle signals in main — duplicate Shutdown caused panic (close of closed channel).
	done := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", *port)
		logrus.WithField("addr", addr).Info("Starting HTTP server")
		done <- server.ListenAndServe(addr)
	}()

	if err := <-done; err != nil {
		log.Fatalf("Server error: %v", err)
	}
	logrus.Info("Server stopped")

	// Cleanup AI service if we started it
	cleanupAIService(aiServicePID)
}

const defaultAIPort = 18081

func startAIService() int {
	aiPort := getAIPort()
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", aiPort)

	// Check if AI service is already running
	if aiIsUp(healthURL) {
		logrus.Info("AI service (FlyMind) already running")
		return 0
	}

	// Check if uv is available
	aiServicePath := filepath.Join(findRoot(), "ai-service", "pyproject.toml")
	if _, err := os.Stat(aiServicePath); os.IsNotExist(err) {
		logrus.Warn("AI service not found, skipping auto-start")
		return 0
	}

	logrus.Info("Starting AI service (FlyMind)...")

	cmd := exec.Command("uv", "run", "uvicorn", "src.main:app",
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(aiPort))
	cmd.Dir = filepath.Join(findRoot(), "ai-service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Inherit environment but set AI_SERVICE_URL
	env := os.Environ()
	env = append(env, fmt.Sprintf("AI_SERVICE_URL=http://127.0.0.1:%d", aiPort))
	env = append(env, "ENABLE_RAG=false")
	env = append(env, "REDIS_URL=redis://localhost:6379")
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		logrus.WithError(err).Warn("Failed to start AI service")
		return 0
	}

	// Wait for AI service to be ready
	for i := 0; i < 80; i++ {
		if aiIsUp(healthURL) {
			logrus.Info("AI service (FlyMind) ready")
			return cmd.Process.Pid
		}
		time.Sleep(250 * time.Millisecond)
	}

	logrus.Warn("AI service did not become healthy in time, continuing anyway")
	return cmd.Process.Pid
}

func getAIPort() int {
	if port := os.Getenv("AI_SERVICE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			return p
		}
	}
	return defaultAIPort
}

func aiIsUp(healthURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "curl", "-sf", healthURL)
	return cmd.Run() == nil
}

func cleanupAIService(pid int) {
	if pid > 0 {
		logrus.Info("Stopping AI service...")
		proc, _ := os.FindProcess(pid)
		if proc != nil {
			proc.Kill()
			proc.Wait()
		}
	}
}

func findRoot() string {
	cwd, _ := os.Getwd()
	return cwd
}

// initializeSearchTools initializes the search tools and registers them with the main tool registry.
func initializeSearchTools() error {
	provider := loadSearchProvider()
	if err := search.Initialize(provider); err != nil {
		return err
	}

	// Register search tools with the main tool registry
	registry := tools.GetRegistry()
	for _, tool := range search.GetRegistry().List() {
		if err := registry.Register(tool); err != nil {
			// Log but don't fail - tool might already be registered
			logrus.Warnf("tool %s already registered: %v", tool.Name(), err)
		}
	}

	return nil
}

// loadSearchProvider loads the search provider based on environment configuration.
func loadSearchProvider() search.SearchProvider {
	providerType := os.Getenv("SEARCH_PROVIDER")
	switch providerType {
	case "serpapi":
		return providers.NewSERPProvider(search.ProviderConfig{
			APIKey:           os.Getenv("SERPAPI_KEY"),
			APIURL:           os.Getenv("SERPAPI_URL"),
			TimeoutMs:        30000,
			RateLimitPerMinute: 60,
		})
	default:
		// Default to mock provider for development
		return providers.NewMockProvider()
	}
}
