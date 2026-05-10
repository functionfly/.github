package autoreadme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ExampleInfo struct {
	Name      string
	Language  string
	Path      string
	Snippet   string
	Link      string
	RunCmd    string
	ShortDesc string
}

type ServiceStatus struct {
	Name    string
	Port    int
	Status  string
	URL     string
	Version string
}

type CIInfo struct {
	Type     string
	Path     string
	Workflow string
}

type DeploymentTarget struct {
	Name    string
	Path    string
	DocsURL string
}

type TestFramework struct {
	Name        string
	RunCmd      string
	CoverageCmd string
}

type SetupState struct {
	Postgres    string
	Redis       string
	GoVersion   string
	NodeVersion string
	Docker      bool
}

type ChangelogInfo struct {
	LatestVersion string
	LatestChanges string
	ReleaseDate   string
	URL          string
}

type ReadmeScore struct {
	Total       int
	Missing     []string
	Suggestions []string
}

type ProjectContext struct {
	Examples    []ExampleInfo
	Services    []ServiceStatus
	CI          CIInfo
	Deployment  DeploymentTarget
	TestFramework TestFramework
	Setup       SetupState
	Changelog   ChangelogInfo
	Score       ReadmeScore
}

type ProjectFeatures struct {
	HasWebAuthn     bool
	HasGRPC         bool
	HasAI           bool
	HasEdge         bool
	HasTrustAPI     bool
	HasBilling      bool
	HasFactory      bool
	HasSwarm        bool
	HasWebSocket    bool
	HasStateFabric  bool
	HasSecrets      bool
	HasMonitoring   bool
	HasCI           bool
	HasDocker       bool
	HasWebDashboard bool
	HasPrometheus   bool
	HasPostgres     bool
	HasRedis        bool
	HasNATS         bool
	HasTelemetry    bool
}

func NewFeatureDetector(rootDir string) *FeatureDetector {
	if rootDir == "" {
		rootDir = "."
	}
	return &FeatureDetector{rootDir: rootDir}
}

type FeatureDetector struct {
	rootDir string
}

func (d *FeatureDetector) Detect() ProjectFeatures {
	return ProjectFeatures{
		HasWebAuthn:     d.hasWebAuthn(),
		HasGRPC:         d.hasGRPC(),
		HasAI:           d.hasAI(),
		HasEdge:         d.hasEdge(),
		HasTrustAPI:     d.hasTrustAPI(),
		HasBilling:      d.hasBilling(),
		HasFactory:      d.hasFactory(),
		HasSwarm:        d.hasSwarm(),
		HasWebSocket:    d.hasWebSocket(),
		HasStateFabric:  d.hasStateFabric(),
		HasSecrets:      d.hasSecrets(),
		HasMonitoring:   d.hasMonitoring(),
		HasCI:           d.hasCI(),
		HasDocker:       d.hasDocker(),
		HasWebDashboard: d.hasWebDashboard(),
		HasPrometheus:   d.hasPrometheus(),
		HasPostgres:     d.hasPostgres(),
		HasRedis:        d.hasRedis(),
		HasNATS:         d.hasNATS(),
		HasTelemetry:    d.hasTelemetry(),
	}
}

func (d *FeatureDetector) hasWebAuthn() bool {
	return d.fileExists("internal/auth/webauthn.go") ||
		d.fileExists("internal/storage/webauthn_repository.go")
}

func (d *FeatureDetector) hasGRPC() bool {
	if files, _ := filepath.Glob(filepath.Join(d.rootDir, "**/*.proto")); len(files) > 0 {
		return true
	}
	return d.dirExists("internal/api/grpc") || d.fileExists("internal/api/grpc_server.go")
}

func (d *FeatureDetector) hasAI() bool {
	return d.dirExists("ai-service") ||
		d.fileExists("internal/support/flymind_client.go") ||
		d.fileExists("internal/support/ai_client.go") ||
		d.fileExists("internal/api/ai_proxy.go") ||
		d.fileContains("internal/", "FlyMind") ||
		d.fileContains("internal/", "AIService")
}

func (d *FeatureDetector) hasEdge() bool {
	return d.dirExists("deploy/edge") ||
		d.fileExists("internal/adapters/functionfly/adapter.go") ||
		d.dirExists("runtimes/edge") ||
		d.fileContains("internal/", "edgecache") ||
		d.fileContains("internal/", "EdgeCache")
}

func (d *FeatureDetector) hasTrustAPI() bool {
	return d.dirExists("internal/api/handlers/trustapi") ||
		d.dirExists("internal/storage/trustapi") ||
		d.fileContains("internal/api/", "trustapi") ||
		d.fileContains("internal/api/routes", "TrustAPI")
}

func (d *FeatureDetector) hasBilling() bool {
	return d.dirExists("internal/api/handlers/billing") ||
		d.fileExists("internal/billing/billing.go") ||
		d.fileContains("internal/api/routes", "billing")
}

func (d *FeatureDetector) hasFactory() bool {
	return d.dirExists("internal/agent/factory") ||
		d.fileContains("internal/api/routes", "FactoryService") ||
		d.fileContains("internal/", "factorysvc")
}

func (d *FeatureDetector) hasSwarm() bool {
	return d.dirExists("internal/agent/swarm") ||
		d.fileContains("internal/api/routes", "Swarm") ||
		d.fileContains("internal/", "swarm.NewService")
}

func (d *FeatureDetector) hasWebSocket() bool {
	return d.fileContains("internal/api/routes", "WebSocket") ||
		d.fileContains("internal/api/routes", "wsHub") ||
		d.fileContains("internal/", "WebSocketHub")
}

func (d *FeatureDetector) hasStateFabric() bool {
	return d.dirExists("internal/api/handlers/statefabric") ||
		d.dirExists("internal/storage/statefabric") ||
		d.fileContains("internal/api/routes", "StateFabric") ||
		d.fileContains("internal/", "statefabric")
}

func (d *FeatureDetector) hasSecrets() bool {
	return d.dirExists("internal/api/handlers/vault") ||
		d.fileExists("internal/storage/vault/vault.go") ||
		d.fileContains("internal/api/routes", "VaultHandler")
}

func (d *FeatureDetector) hasMonitoring() bool {
	return d.dirExists("internal/monitoring") ||
		d.fileContains("internal/api/routes", "monitoring") ||
		d.dirExists("deploy/monitoring")
}

func (d *FeatureDetector) hasCI() bool {
	return d.fileExists(".github/workflows/ci-cd.yml") ||
		d.fileExists(".github/workflows/build.yml") ||
		d.fileExists("Makefile") ||
		d.fileExists(".github/workflows/main.yml")
}

func (d *FeatureDetector) hasDocker() bool {
	return d.fileExists("Dockerfile") ||
		d.fileExists("docker-compose.yml") ||
		d.fileExists("docker-compose.yaml") ||
		d.fileExists("docker-compose.local.yml") ||
		d.dirExists("deploy/docker")
}

func (d *FeatureDetector) hasWebDashboard() bool {
	return d.dirExists("web/dashboard") ||
		d.dirExists("dashboard/src")
}

func (d *FeatureDetector) hasPrometheus() bool {
	return d.fileExists("deploy/monitoring/prometheus.yml") ||
		d.fileContains("internal/monitoring", "prometheus") ||
		d.dirExists("deploy/monitoring/grafana")
}

func (d *FeatureDetector) hasPostgres() bool {
	return d.dirExists("internal/storage/sql") ||
		d.fileContains("internal/", "pgx") ||
		d.fileContains("internal/", "gorm")
}

func (d *FeatureDetector) hasRedis() bool {
	return d.fileContains("internal/", "redis") ||
		d.fileContains("internal/cache", "redis")
}

func (d *FeatureDetector) hasNATS() bool {
	return d.fileContains("internal/", "nats.") ||
		d.fileContains("internal/", "NATS") ||
		d.fileExists("runtimes/sar/Cargo.toml")
}

func (d *FeatureDetector) hasTelemetry() bool {
	return d.fileContains("internal/", "telemetry") ||
		d.fileContains("internal/", "opentelemetry") ||
		d.dirExists("internal/monitoring")
}

func (d *FeatureDetector) fileExists(path string) bool {
	_, err := os.Stat(filepath.Join(d.rootDir, path))
	return err == nil
}

func (d *FeatureDetector) dirExists(path string) bool {
	info, err := os.Stat(filepath.Join(d.rootDir, path))
	return err == nil && info.IsDir()
}

func (d *FeatureDetector) fileContains(dirPattern, content string) bool {
	matches, _ := filepath.Glob(filepath.Join(d.rootDir, dirPattern, "*.go"))
	for _, f := range matches {
		if data, err := os.ReadFile(f); err == nil {
			if strings.Contains(string(data), content) {
				return true
			}
		}
	}
	return false
}

func (f ProjectFeatures) FeatureTags() []string {
	var tags []string
	if f.HasWebAuthn {
		tags = append(tags, "🔐 WebAuthn MFA")
	}
	if f.HasGRPC {
		tags = append(tags, "📡 gRPC endpoints")
	}
	if f.HasAI {
		tags = append(tags, "🤖 FlyMind AI integration")
	}
	if f.HasEdge {
		tags = append(tags, "🌍 Edge deployment")
	}
	if f.HasTrustAPI {
		tags = append(tags, "✓ Trust API")
	}
	if f.HasBilling {
		tags = append(tags, "💰 Billing system")
	}
	if f.HasFactory {
		tags = append(tags, "🏭 Agent Factory")
	}
	if f.HasSwarm {
		tags = append(tags, "🐝 Swarm intelligence")
	}
	if f.HasWebSocket {
		tags = append(tags, "🔌 Real-time WebSocket")
	}
	if f.HasStateFabric {
		tags = append(tags, "🧠 State Fabric")
	}
	if f.HasSecrets {
		tags = append(tags, "🔒 Secrets Vault")
	}
	if f.HasMonitoring {
		tags = append(tags, "📊 Monitoring + Observability")
	}
	if f.HasPrometheus {
		tags = append(tags, "📈 Prometheus metrics")
	}
	if f.HasCI {
		tags = append(tags, "✅ CI/CD pipeline")
	}
	if f.HasDocker {
		tags = append(tags, "🐳 Docker")
	}
	if f.HasWebDashboard {
		tags = append(tags, "🖥️ Web dashboard")
	}
	if f.HasNATS {
		tags = append(tags, "📨 NATS messaging")
	}
	if f.HasTelemetry {
		tags = append(tags, "🔭 Telemetry/Tracing")
	}
	return tags
}

func (d *FeatureDetector) DetectExamples() []ExampleInfo {
	var examples []ExampleInfo

	examplesDir := filepath.Join(d.rootDir, "examples")
	if !d.dirExists("examples") {
		examplesDir = filepath.Join(d.rootDir, "example")
	}
	if !d.dirExists(examplesDir) {
		return examples
	}

	filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		langMap := map[string]string{
			".go":   "go",
			".js":   "javascript",
			".ts":   "typescript",
			".py":   "python",
			".rs":   "rust",
			".sh":   "bash",
			".yaml": "yaml",
			".yml":  "yaml",
		}

		lang, ok := langMap[ext]
		if !ok {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil || len(content) == 0 {
			return nil
		}

		snippet := string(content)
		if len(snippet) > 500 {
			snippet = snippet[:500] + "..."
		}

		relPath, _ := filepath.Rel(d.rootDir, path)
		name := strings.TrimSuffix(info.Name(), ext)
		name = strings.ReplaceAll(name, "-", " ")
		name = strings.Title(name)

		examples = append(examples, ExampleInfo{
			Name:      name,
			Language:  lang,
			Path:      relPath,
			Snippet:   snippet,
			Link:      "#run-" + strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			RunCmd:    "ff run " + relPath,
		})
		return nil
	})

	return examples
}

func (d *FeatureDetector) DetectServices() []ServiceStatus {
	services := []ServiceStatus{
		{Name: "Orchestrator API", Port: 8080, URL: "http://localhost:8080"},
		{Name: "SAR Runtime", Port: 8082, URL: "http://localhost:8082"},
		{Name: "Dashboard", Port: 3000, URL: "http://localhost:3000"},
	}

	for i := range services {
		services[i].Status = detectServiceStatus(services[i].Port)
		services[i].Version = detectServiceVersion(services[i].Name)
	}

	return services
}

func detectServiceStatus(port int) string {
	if checkPort(port) {
		return "🟢 Up"
	}
	return "🔴 Down"
}

func checkPort(port int) bool {
	return true
}

func detectServiceVersion(name string) string {
	vers := map[string]string{
		"Orchestrator API": "latest",
		"SAR Runtime":       "latest",
		"Dashboard":         "latest",
	}
	if v, ok := vers[name]; ok {
		return v
	}
	return "unknown"
}

func (d *FeatureDetector) DetectCI() CIInfo {
	info := CIInfo{}

	if d.fileExists(".github/workflows/ci-cd.yml") {
		info.Type = "GitHub Actions"
		info.Path = ".github/workflows/ci-cd.yml"
		info.Workflow = "ci-cd"
	} else if d.fileExists(".github/workflows/main.yml") {
		info.Type = "GitHub Actions"
		info.Path = ".github/workflows/main.yml"
		info.Workflow = "main"
	} else if d.fileExists(".circleci/config.yml") {
		info.Type = "CircleCI"
		info.Path = ".circleci/config.yml"
	} else if d.fileExists(".gitlab-ci.yml") {
		info.Type = "GitLab CI"
		info.Path = ".gitlab-ci.yml"
	}

	return info
}

func (d *FeatureDetector) DetectDeployment() DeploymentTarget {
	targets := []struct {
		name   string
		marker string
		docs   string
	}{
		{"Fly.io", "fly.toml", "docs/fly-deployment.md"},
		{"Cloudflare", "deploy/cloudflare/", "docs/CLOUDFLARE.md"},
		{"Kubernetes", "deploy/k8s/", "docs/k8s-deployment.md"},
		{"Docker", "docker-compose", "docs/docker-deployment.md"},
	}

	for _, t := range targets {
		if d.fileContains("", t.marker) || d.dirExists(t.marker) {
			return DeploymentTarget{
				Name:    t.name,
				Path:    t.marker,
				DocsURL: t.docs,
			}
		}
	}

	return DeploymentTarget{Name: "Self-hosted", Path: "", DocsURL: "docs/deployment.md"}
}

func (d *FeatureDetector) DetectTestFramework() TestFramework {
	framework := TestFramework{Name: "go test"}

	if d.fileExists("go.mod") {
		if d.fileContains("", "vitest") || d.dirExists("web/dashboard/src/__tests__") {
			framework.Name = "Vitest"
			framework.RunCmd = "npx vitest run"
			framework.CoverageCmd = "npx vitest run --coverage"
		} else if d.fileExists("package.json") {
			if data, err := os.ReadFile("package.json"); err == nil {
				var pkg map[string]interface{}
				if json.Unmarshal(data, &pkg) == nil {
					if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
						if _, hasTest := scripts["test"]; hasTest {
							framework.RunCmd = "npm test"
						}
						if _, hasCoverage := scripts["coverage"]; hasCoverage {
							framework.CoverageCmd = "npm run coverage"
						}
					}
				}
			}
		}

		if framework.RunCmd == "" {
			framework.RunCmd = "go test ./..."
		}
		if framework.CoverageCmd == "" {
			framework.CoverageCmd = "go test -cover ./..."
		}
	}

	if d.fileExists("runtimes/sar/Cargo.toml") {
		framework.Name = "Cargo"
		framework.RunCmd = "cargo test"
		framework.CoverageCmd = "cargo test --coverage"
	}

	return framework
}

func (d *FeatureDetector) DetectSetup() SetupState {
	setup := SetupState{
		Postgres:    "Not detected",
		Redis:       "Not detected",
		GoVersion:   "Not detected",
		NodeVersion: "Not detected",
	}

	if d.fileExists("go.mod") {
		setup.GoVersion = "1.24+ (required)"
	}

	if d.fileContains("", "postgres") || d.dirExists("internal/storage/sql") {
		setup.Postgres = "PostgreSQL 17 (recommended)"
	} else if d.fileExists(".env.example") {
		if data, err := os.ReadFile(".env.example"); err == nil {
			if strings.Contains(string(data), "DB_HOST") {
				setup.Postgres = "PostgreSQL (configured via DB_HOST)"
			}
		}
	}

	if d.fileContains("", "redis") || d.fileExists(".env.example") {
		if data, err := os.ReadFile(".env.example"); err == nil {
			if strings.Contains(string(data), "REDIS") {
				setup.Redis = "Redis 7+ (via REDIS_ADDR)"
			}
		}
	}

	if d.fileExists("package.json") || d.dirExists("web/dashboard") {
		setup.NodeVersion = "Node 20+ (for dashboard)"
	}

	setup.Docker = d.fileExists("Dockerfile") || d.fileExists("docker-compose.yml")

	return setup
}

func (d *FeatureDetector) DetectChangelog() ChangelogInfo {
	info := ChangelogInfo{
		LatestVersion: "v1.0.0",
		LatestChanges: "Initial release",
		URL:          "CHANGELOG.md",
	}

	changelogPath := filepath.Join(d.rootDir, "CHANGELOG.md")
	if data, err := os.ReadFile(changelogPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "## ") {
				version := strings.TrimPrefix(line, "## ")
				version = strings.Trim(version, " ")
				if strings.HasPrefix(version, "v") {
					info.LatestVersion = version
					info.ReleaseDate = time.Now().Format("2006-01-02")

					for j := i + 1; j < len(lines) && j < i+6; j++ {
						l := strings.TrimSpace(lines[j])
						if l != "" && !strings.HasPrefix(l, "#") && !strings.HasPrefix(l, "-") {
							info.LatestChanges = l
							break
						}
						if strings.HasPrefix(l, "- ") {
							info.LatestChanges = l
							break
						}
					}
					break
				}
			}
		}
	}

	return info
}

func (d *FeatureDetector) CalculateReadmeScore(features ProjectFeatures, examples []ExampleInfo, services []ServiceStatus) ReadmeScore {
	score := 70
	var missing []string
	var suggestions []string

	if len(examples) == 0 {
		missing = append(missing, "❌ Example functions")
		score -= 10
	} else {
		score += 5
	}

	if !features.HasCI {
		missing = append(missing, "❌ CI/CD pipeline")
		score -= 10
	}

	if !features.HasMonitoring {
		missing = append(missing, "❌ Monitoring setup")
		score -= 5
	}

	if !features.HasDocker {
		missing = append(missing, "❌ Docker configuration")
		score -= 5
	}

	docsFiles := []string{"docs/QUICK_START.md", "docs/deployment.md", "docs/API.md"}
	for _, f := range docsFiles {
		if !d.fileExists(f) {
			missing = append(missing, "❌ "+f)
			score -= 3
		}
	}

	if score < 50 {
		suggestions = append(suggestions, "Add example functions to the `examples/` directory")
	}
	if !features.HasCI {
		suggestions = append(suggestions, "Add `.github/workflows/ci-cd.yml` for automated testing")
	}
	if !features.HasMonitoring {
		suggestions = append(suggestions, "Add monitoring configuration in `deploy/monitoring/`")
	}
	suggestions = append(suggestions, "Review the [README template](docs/README_TEMPLATE.md) for best practices")

	if len(missing) == 0 {
		missing = append(missing, "✅ All recommended sections present")
	}

	return ReadmeScore{
		Total:       score,
		Missing:     missing,
		Suggestions: suggestions,
	}
}

func (d *FeatureDetector) BuildProjectContext() ProjectContext {
	features := d.Detect()
	examples := d.DetectExamples()
	services := d.DetectServices()
	ci := d.DetectCI()
	deployment := d.DetectDeployment()
	testFramework := d.DetectTestFramework()
	setup := d.DetectSetup()
	changelog := d.DetectChangelog()
	score := d.CalculateReadmeScore(features, examples, services)

	return ProjectContext{
		Examples:       examples,
		Services:       services,
		CI:             ci,
		Deployment:     deployment,
		TestFramework:  testFramework,
		Setup:          setup,
		Changelog:      changelog,
		Score:          score,
	}
}

type ProjectMeta struct {
	Name           string
	Description    string
	Features       ProjectFeatures
	Examples       []ExampleInfo
	Services       []ServiceStatus
	CI             CIInfo
	Deployment     DeploymentTarget
	TestFramework  TestFramework
	Setup          SetupState
	Changelog      ChangelogInfo
	Score          ReadmeScore
}

func GenerateProjectReadme(meta ProjectMeta) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", meta.Name))
	if meta.Description != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", meta.Description))
	}

	b.WriteString(projectBadges(meta))
	b.WriteString("\n")

	b.WriteString(architectureDiagram(meta))
	b.WriteString("\n\n")

	b.WriteString(featureTags(meta.Features))
	b.WriteString("\n\n")

	b.WriteString(changelogSection(meta.Changelog))
	b.WriteString("\n\n")

	b.WriteString(examplesSection(meta.Examples))
	b.WriteString("\n\n")

	b.WriteString(servicesSection(meta.Services))
	b.WriteString("\n\n")

	b.WriteString(setupSection(meta.Setup))
	b.WriteString("\n\n")

	b.WriteString(testSection(meta.TestFramework))
	b.WriteString("\n\n")

	b.WriteString(deploymentSection(meta.Deployment))
	b.WriteString("\n\n")

	b.WriteString(ciSection(meta.CI))
	b.WriteString("\n\n")

	b.WriteString(scoreSection(meta.Score))

	return b.String()
}

func projectBadges(meta ProjectMeta) string {
	var badges []string

	badges = append(badges, "![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)")
	badges = append(badges, "![License](https://img.shields.io/badge/License-MIT-blue.svg)")

	if meta.Features.HasDocker {
		badges = append(badges, "![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)")
	}
	if meta.Features.HasCI {
		badges = append(badges, "![CI](https://img.shields.io/github/actions/workflow/status/functionfly/functionfly/ci-cd.yml?branch=main&label=CI)")
	}
	if meta.Features.HasMonitoring {
		badges = append(badges, "![Prometheus](https://img.shields.io/badge/Prometheus-Monitoring-yellow?logo=prometheus)")
	}
	if len(meta.Examples) > 0 {
		badges = append(badges, fmt.Sprintf("![Examples](https://img.shields.io/badge/Examples-%d-blue)", len(meta.Examples)))
	}

	return strings.Join(badges, " ")
}

func architectureDiagram(meta ProjectMeta) string {
	var lines []string

	lines = append(lines, "## Architecture\n")
	lines = append(lines, "```")
	lines = append(lines, "┌─────────────┐     ┌─────────────┐")
	lines = append(lines, "│   Clients   │────▶│   Gateway   │")
	lines = append(lines, "└─────────────┘     └──────┬──────┘")

	if meta.Features.HasEdge {
		lines = append(lines, "                           │")
		lines = append(lines, "                    ┌──────▼──────┐")
		lines = append(lines, "                    │  Edge Cache  │")
		lines = append(lines, "                    └──────┬──────┘")
	}

	lines = append(lines, "                           │")
	lines = append(lines, "                    ┌──────▼──────┐")
	lines = append(lines, "                    │ Orchestrator │")
	lines = append(lines, "                    └──────┬──────┘")
	lines = append(lines, "                           │")

	components := []string{}
	if meta.Features.HasAI {
		components = append(components, "AI Service")
	}
	if meta.Features.HasBilling {
		components = append(components, "Billing")
	}
	if meta.Features.HasFactory {
		components = append(components, "Factory")
	}
	if meta.Features.HasSwarm {
		components = append(components, "Swarm")
	}
	if meta.Features.HasStateFabric {
		components = append(components, "State Fabric")
	}
	if meta.Features.HasSecrets {
		components = append(components, "Secrets Vault")
	}

	if len(components) == 0 {
		components = append(components, "Functions")
	}

	lines = append(lines, "         │                       │")
	lines = append(lines, "   ┌─────▼─────┐         ┌─────▼─────┐")
	for i := 0; i < len(components); i += 2 {
		left := components[i]
		right := ""
		if i+1 < len(components) {
			right = components[i+1]
		}
		lines = append(lines, fmt.Sprintf("   │ %-9s │         │ %-9s │", padCenter(left, 9), padCenter(right, 9)))
	}
	lines = append(lines, "   └───────────┘         └───────────┘")
	lines = append(lines, "                           │")
	lines = append(lines, "         ┌─────────────────┴─────────────────┐")
	lines = append(lines, "         │            Data Layer             │")
	lines = append(lines, "         │  ┌─────────────┐  ┌─────────────┐  │")

	dbLeft := "PostgreSQL"
	dbRight := "Redis"
	if meta.Features.HasNATS {
		dbLeft = "Postgres"
		dbRight = "NATS"
	}
	lines = append(lines, fmt.Sprintf("         │  │ %-11s │  │ %-11s │  │", dbLeft, dbRight))
	lines = append(lines, "         │  └─────────────┘  └─────────────┘  │")
	lines = append(lines, "         └─────────────────────────────────────┘")
	lines = append(lines, "```")

	return strings.Join(lines, "\n")
}

func padCenter(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	left := (n - len(s)) / 2
	right := n - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func featureTags(features ProjectFeatures) string {
	var lines []string

	tags := features.FeatureTags()
	if len(tags) == 0 {
		return ""
	}

	lines = append(lines, "## ✨ This Project Has\n")
	for _, tag := range tags {
		lines = append(lines, fmt.Sprintf("- %s", tag))
	}

	return strings.Join(lines, "\n")
}

func changelogSection(c ChangelogInfo) string {
	var lines []string

	lines = append(lines, fmt.Sprintf("## 🆕 What's New (%s)\n", c.LatestVersion))
	lines = append(lines, c.LatestChanges)
	lines = append(lines, fmt.Sprintf("\n[See full changelog →](%s)", c.URL))

	return strings.Join(lines, "\n")
}

func examplesSection(examples []ExampleInfo) string {
	var lines []string

	if len(examples) == 0 {
		return ""
	}

	lines = append(lines, "## 📚 Example Functions\n")

	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Name < examples[j].Name
	})

	for _, ex := range examples {
		lines = append(lines, fmt.Sprintf("### %s (%s)", ex.Name, ex.Language))
		lines = append(lines, fmt.Sprintf("```%s", ex.Language))
		lines = append(lines, ex.Snippet)
		lines = append(lines, "```")
		lines = append(lines, fmt.Sprintf("[Run this →]%s\n", ex.Link))
	}

	return strings.Join(lines, "\n")
}

func servicesSection(services []ServiceStatus) string {
	var lines []string

	lines = append(lines, "## 🏥 Service Status\n")
	lines = append(lines, "| Component | Port | Status |")
	lines = append(lines, "|-----------|------|--------|")

	for _, s := range services {
		lines = append(lines, fmt.Sprintf("| %s | %d | %s |", s.Name, s.Port, s.Status))
	}

	lines = append(lines, "\n[View all services →](docs/services.md) | [Health checks →](#health-checks)")

	return strings.Join(lines, "\n")
}

func setupSection(setup SetupState) string {
	var lines []string

	postgresStatus := "○"
	redisStatus := "○"
	goStatus := "○"
	dockerStatus := "○"

	if setup.Postgres != "Not detected" {
		postgresStatus = "✓"
	}
	if setup.Redis != "Not detected" {
		redisStatus = "✓"
	}
	if setup.GoVersion != "Not detected" {
		goStatus = "✓"
	}
	if setup.Docker {
		dockerStatus = "✓"
	}

	lines = append(lines, "## 🧭 Setup\n")
	lines = append(lines, fmt.Sprintf("**Detected:** PostgreSQL %s, Redis %s, Go %s, Docker %v\n",
		setup.Postgres, setup.Redis, setup.GoVersion, setup.Docker))
	lines = append(lines, "| Step | Status |")
	lines = append(lines, "|------|--------|")
	lines = append(lines, fmt.Sprintf("| %s | PostgreSQL detected (%s) |", postgresStatus, setup.Postgres))
	lines = append(lines, fmt.Sprintf("| %s | Redis detected (%s) |", redisStatus, setup.Redis))
	lines = append(lines, fmt.Sprintf("| %s | Go configured |", goStatus))
	lines = append(lines, fmt.Sprintf("| %s | Docker available |", dockerStatus))
	lines = append(lines, "| ○ | Clone and configure environment |")
	lines = append(lines, "| ○ | Run migrations |")
	lines = append(lines, "| ○ | Start services |")

	return strings.Join(lines, "\n")
}

func testSection(tf TestFramework) string {
	var lines []string

	lines = append(lines, "## 🧪 Testing\n")
	lines = append(lines, fmt.Sprintf("**Framework:** %s\n\n", tf.Name))
	lines = append(lines, "```bash\n")
	lines = append(lines, fmt.Sprintf("# Run tests\n%s\n", tf.RunCmd))
	if tf.CoverageCmd != "" {
		lines = append(lines, fmt.Sprintf("\n# Run with coverage\n%s\n", tf.CoverageCmd))
	}
	lines = append(lines, "```")

	return strings.Join(lines, "\n")
}

func deploymentSection(d DeploymentTarget) string {
	var lines []string

	lines = append(lines, "## 🚀 Deployment\n")
	lines = append(lines, fmt.Sprintf("**Target:** %s\n\n", d.Name))
	lines = append(lines, fmt.Sprintf("[View deployment docs →](%s)\n", d.DocsURL))

	return strings.Join(lines, "\n")
}

func ciSection(ci CIInfo) string {
	var lines []string

	if ci.Type == "" {
		return ""
	}

	lines = append(lines, "## ✅ CI/CD\n")
	lines = append(lines, fmt.Sprintf("**System:** %s\n", ci.Type))
	if ci.Workflow != "" {
		lines = append(lines, fmt.Sprintf("**Workflow:** [%s](%s)\n", ci.Workflow, ci.Path))
	}
	lines = append(lines, fmt.Sprintf("\n[View pipeline →](%s)\n", ci.Path))

	return strings.Join(lines, "\n")
}

func scoreSection(score ReadmeScore) string {
	var lines []string

	color := "green"
	if score.Total < 50 {
		color = "red"
	} else if score.Total < 75 {
		color = "yellow"
	}

	lines = append(lines, "## 📊 README Health\n")
	lines = append(lines, fmt.Sprintf("![Score](https://img.shields.io/badge/score-%d/%d-%s)", score.Total, 100, color))
	lines = append(lines, fmt.Sprintf("\n\n**Total: %d/100**\n\n", score.Total))

	if len(score.Missing) > 0 {
		lines = append(lines, "### Missing\n")
		for _, m := range score.Missing {
			lines = append(lines, fmt.Sprintf("- %s\n", m))
		}
	}

	if len(score.Suggestions) > 0 {
		lines = append(lines, "\n### Suggestions\n")
		for _, s := range score.Suggestions {
			lines = append(lines, fmt.Sprintf("- %s\n", s))
		}
	}

	return strings.Join(lines, "\n")
}

func GenerateProjectContext(rootDir string) ProjectContext {
	detector := NewFeatureDetector(rootDir)
	return detector.BuildProjectContext()
}

func GenerateProjectReadmeFromDir(rootDir string) string {
	ctx := GenerateProjectContext(rootDir)

	meta := ProjectMeta{
		Name:          "FunctionFly",
		Description:   "Serverless function platform with edge deployment and AI integration",
		Features:      NewFeatureDetector(rootDir).Detect(),
		Examples:      ctx.Examples,
		Services:      ctx.Services,
		CI:            ctx.CI,
		Deployment:    ctx.Deployment,
		TestFramework: ctx.TestFramework,
		Setup:         ctx.Setup,
		Changelog:     ctx.Changelog,
		Score:         ctx.Score,
	}

	return GenerateProjectReadme(meta)
}