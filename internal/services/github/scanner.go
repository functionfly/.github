package github

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type Detector interface {
	Name() string
	Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error)
	Priority() int
}

type Scanner struct {
	client    *Client
	detectors []Detector
	logger    *logrus.Logger
}

func NewScanner(client *Client, logger *logrus.Logger) *Scanner {
	s := &Scanner{
		client: client,
		logger: logger,
	}
	s.detectors = []Detector{
		&ExplicitConfigDetector{client: client, logger: logger},
		&ServerlessFrameworkDetector{client: client, logger: logger},
		&NodeDetector{logger: logger},
		&PythonDetector{logger: logger},
		&GoDetector{logger: logger},
		&RustDetector{logger: logger},
	}
	return s
}

func (s *Scanner) ScanRepo(ctx context.Context, owner, repo, branch string) (*ScanResult, error) {
	repoInfo, err := s.client.GetRepo(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("fetch repo info: %w", err)
	}

	if branch == "" {
		branch = repoInfo.DefaultBranch
	}

	branches, err := s.client.ListBranches(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}

	var branchSHA string
	for _, b := range branches {
		if b.Name == branch {
			branchSHA = b.Commit.SHA
			break
		}
	}
	if branchSHA == "" {
		return nil, fmt.Errorf("branch %q not found", branch)
	}

	tree, err := s.client.GetTree(ctx, owner, repo, branchSHA, true)
	if err != nil {
		return nil, fmt.Errorf("fetch tree: %w", err)
	}

	languages, err := s.client.GetLanguages(ctx, owner, repo)
	if err != nil {
		s.logger.WithError(err).Warn("failed to fetch languages, continuing without")
		languages = make(map[string]float64)
	}

	entries := tree.Tree

	var highConfidenceResult *ScanResult
	var allResults []*ScanResult

	for _, detector := range s.detectors {
		result, err := detector.Detect(ctx, repoInfo, entries)
		if err != nil {
			s.logger.WithError(err).WithField("detector", detector.Name()).Warn("detector failed")
			continue
		}
		if result == nil {
			continue
		}
		allResults = append(allResults, result)

		if result.OverallConfidence > 0.7 && highConfidenceResult == nil {
			highConfidenceResult = result
		}
	}

	var finalResult *ScanResult
	if highConfidenceResult != nil {
		finalResult = highConfidenceResult
	} else {
		finalResult = s.mergeResults(allResults)
	}

	if finalResult == nil {
		finalResult = &ScanResult{
			Functions:         []DetectedFunction{},
			OverallConfidence: 0,
			Warnings:          []string{"no functions detected"},
		}
	}

	if finalResult.PrimaryRuntime == "" {
		finalResult.PrimaryRuntime = s.detectPrimaryRuntime(languages)
	}

	finalResult.EstimatedImportTimeS = s.estimateImportTime(finalResult)
	finalResult.EstimatedCostUSD = s.estimateCost(finalResult)

	return finalResult, nil
}

func (s *Scanner) mergeResults(results []*ScanResult) *ScanResult {
	if len(results) == 0 {
		return nil
	}

	merged := &ScanResult{}
	bestConfidence := 0.0
	for _, r := range results {
		merged.Functions = append(merged.Functions, r.Functions...)
		merged.Warnings = append(merged.Warnings, r.Warnings...)
		if r.OverallConfidence > bestConfidence {
			bestConfidence = r.OverallConfidence
			merged.StrategyUsed = r.StrategyUsed
		}
	}
	merged.OverallConfidence = bestConfidence

	seen := make(map[string]bool)
	var deduped []DetectedFunction
	for _, fn := range merged.Functions {
		key := fn.EntryPoint + ":" + fn.Runtime
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, fn)
		}
	}
	merged.Functions = deduped

	return merged
}

func (s *Scanner) detectPrimaryRuntime(languages map[string]float64) string {
	type langPct struct {
		name    string
		percent float64
	}
	var sorted []langPct
	for lang, pct := range languages {
		sorted = append(sorted, langPct{lang, pct})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].percent > sorted[i].percent {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if len(sorted) == 0 {
		return "unknown"
	}

	lang := sorted[0].name
	switch strings.ToLower(lang) {
	case "typescript", "javascript":
		return "node"
	case "python":
		return "python"
	case "go":
		return "go"
	case "rust":
		return "rust"
	default:
		return strings.ToLower(lang)
	}
}

func (s *Scanner) estimateImportTime(result *ScanResult) int {
	base := 10
	perFunction := 5
	return base + len(result.Functions)*perFunction
}

func (s *Scanner) estimateCost(result *ScanResult) float64 {
	base := 0.01
	perFunction := 0.005
	return base + float64(len(result.Functions))*perFunction
}

// ──────────────────────────────────────────────
// ExplicitConfigDetector
// ──────────────────────────────────────────────

type ExplicitConfigDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *ExplicitConfigDetector) Name() string { return "explicit-config" }
func (d *ExplicitConfigDetector) Priority() int { return 100 }

func (d *ExplicitConfigDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var configPath string
	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		if base == "functionfly.jsonc" || base == "functionfly.json" {
			configPath = e.Path
			break
		}
	}
	if configPath == "" {
		return nil, nil
	}

	raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, configPath, "")
	if err != nil {
		return nil, fmt.Errorf("fetch functionfly config: %w", err)
	}

	var config struct {
		Functions []struct {
			Name       string `json:"name"`
			EntryPoint string `json:"entry_point"`
			Runtime    string `json:"runtime"`
			Directory  string `json:"directory"`
		} `json:"functions"`
		Runtime string `json:"runtime"`
	}

	cleaned := stripJSONComments(string(raw))
	if err := json.Unmarshal([]byte(cleaned), &config); err != nil {
		return nil, fmt.Errorf("parse functionfly config: %w", err)
	}

	if len(config.Functions) == 0 {
		return &ScanResult{
			Warnings:          []string{"functionfly config found but contains no functions"},
			OverallConfidence: 0.3,
			StrategyUsed:      d.Name(),
		}, nil
	}

	var fns []DetectedFunction
	for _, f := range config.Functions {
		fn := DetectedFunction{
			Name:         f.Name,
			EntryPoint:   f.EntryPoint,
			Runtime:      f.Runtime,
			SubDirectory: f.Directory,
			Confidence:   1.0,
			Strategy:     d.Name(),
		}
		if fn.Runtime == "" {
			fn.Runtime = config.Runtime
		}
		fns = append(fns, fn)
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    config.Runtime,
		OverallConfidence: 1.0,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// ServerlessFrameworkDetector
// ──────────────────────────────────────────────

type ServerlessFrameworkDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *ServerlessFrameworkDetector) Name() string { return "serverless-framework" }
func (d *ServerlessFrameworkDetector) Priority() int { return 90 }

func (d *ServerlessFrameworkDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var configPath string
	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		if base == "serverless.yml" || base == "serverless.yaml" {
			configPath = e.Path
			break
		}
	}
	if configPath == "" {
		return nil, nil
	}

	raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, configPath, "")
	if err != nil {
		return nil, fmt.Errorf("fetch serverless config: %w", err)
	}

	content := string(raw)
	var fns []DetectedFunction

	lines := strings.Split(content, "\n")
	inFunctions := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "functions:" {
			inFunctions = true
			continue
		}
		if inFunctions {
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" {
				inFunctions = false
				continue
			}
			if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "handler:") && !strings.Contains(trimmed, "runtime:") && !strings.Contains(trimmed, "events:") {
				fnName := strings.TrimSuffix(trimmed, ":")
				fnName = strings.TrimSpace(fnName)
				if fnName != "" && fnName != "functions" {
					fns = append(fns, DetectedFunction{
						Name:       fnName,
						EntryPoint: fnName,
						Confidence: 0.85,
						Strategy:   d.Name(),
					})
				}
			}
		}
	}

	if len(fns) == 0 {
		return &ScanResult{
			Warnings:          []string{"serverless.yml found but no functions parsed"},
			OverallConfidence: 0.3,
			StrategyUsed:      d.Name(),
		}, nil
	}

	return &ScanResult{
		Functions:         fns,
		OverallConfidence: 0.85,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// NodeDetector
// ──────────────────────────────────────────────

type NodeDetector struct {
	logger *logrus.Logger
}

func (d *NodeDetector) Name() string { return "node-detector" }
func (d *NodeDetector) Priority() int { return 50 }

func (d *NodeDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	hasPackageJSON := false
	var entryFiles []string
	var lockfile string
	var packages []string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		switch base {
		case "package.json":
			if !hasPackageJSON {
				hasPackageJSON = true
				dir := filepath.Dir(e.Path)
				if dir == "." {
					dir = ""
				}
				_ = dir
			}
		case "package-lock.json":
			lockfile = e.Path
		case "yarn.lock":
			lockfile = e.Path
		case "pnpm-lock.yaml":
			lockfile = e.Path
		case "bun.lockb":
			lockfile = e.Path
		}

		if base == "index.ts" || base == "index.js" || base == "index.mjs" ||
			base == "handler.ts" || base == "handler.js" || base == "handler.mjs" {
			entryFiles = append(entryFiles, e.Path)
		}
	}

	if !hasPackageJSON || len(entryFiles) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	for _, ef := range entryFiles {
		dir := filepath.Dir(ef)
		name := "default"
		if dir != "." {
			name = filepath.Base(dir)
		}

		runtime := "node18"
		if strings.HasSuffix(ef, ".ts") {
			runtime = "node18-typescript"
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   ef,
			Runtime:      runtime,
			Confidence:   0.6,
			Strategy:     d.Name(),
			SubDirectory: dir,
		}
		if lockfile != "" {
			pkgDir := filepath.Dir(lockfile)
			fn.Dependencies = &DependencyInfo{
				Manager:  detectNodePkgManager(lockfile),
				Lockfile: lockfile,
				Packages: packages,
			}
			_ = pkgDir
		}
		fns = append(fns, fn)
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "node",
		OverallConfidence: 0.6,
		StrategyUsed:      d.Name(),
	}, nil
}

func detectNodePkgManager(lockfile string) string {
	base := filepath.Base(lockfile)
	switch base {
	case "yarn.lock":
		return "yarn"
	case "pnpm-lock.yaml":
		return "pnpm"
	case "bun.lockb":
		return "bun"
	default:
		return "npm"
	}
}

// ──────────────────────────────────────────────
// PythonDetector
// ──────────────────────────────────────────────

type PythonDetector struct {
	logger *logrus.Logger
}

func (d *PythonDetector) Name() string { return "python-detector" }
func (d *PythonDetector) Priority() int { return 40 }

func (d *PythonDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var entryFiles []string
	var lockfile string
	var hasRequirements bool

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		switch base {
		case "requirements.txt":
			hasRequirements = true
		case "Pipfile.lock":
			lockfile = e.Path
		case "poetry.lock":
			lockfile = e.Path
		}
		if base == "main.py" || base == "handler.py" || base == "app.py" {
			entryFiles = append(entryFiles, e.Path)
		}
	}

	if len(entryFiles) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	for _, ef := range entryFiles {
		dir := filepath.Dir(ef)
		name := "default"
		if dir != "." {
			name = filepath.Base(dir)
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   ef,
			Runtime:      "python3.11",
			Confidence:   0.55,
			Strategy:     d.Name(),
			SubDirectory: dir,
		}
		if lockfile != "" {
			fn.Dependencies = &DependencyInfo{
				Manager:  detectPythonPkgManager(lockfile),
				Lockfile: lockfile,
			}
		} else if hasRequirements {
			fn.Dependencies = &DependencyInfo{
				Manager: "pip",
			}
		}
		fns = append(fns, fn)
	}

	confidence := 0.55
	if hasRequirements {
		confidence = 0.65
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "python",
		OverallConfidence: float64(confidence),
		StrategyUsed:      d.Name(),
	}, nil
}

func detectPythonPkgManager(lockfile string) string {
	base := filepath.Base(lockfile)
	switch base {
	case "Pipfile.lock":
		return "pipenv"
	case "poetry.lock":
		return "poetry"
	default:
		return "pip"
	}
}

// ──────────────────────────────────────────────
// GoDetector
// ──────────────────────────────────────────────

type GoDetector struct {
	logger *logrus.Logger
}

func (d *GoDetector) Name() string { return "go-detector" }
func (d *GoDetector) Priority() int { return 30 }

func (d *GoDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	hasGoMod := false
	var mainGoFiles []string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		if base == "go.mod" {
			hasGoMod = true
		}
		if base == "main.go" {
			mainGoFiles = append(mainGoFiles, e.Path)
		}
	}

	if !hasGoMod || len(mainGoFiles) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	for _, ef := range mainGoFiles {
		dir := filepath.Dir(ef)
		name := "default"
		if dir != "." {
			name = filepath.Base(dir)
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   ef,
			Runtime:      "go1.22",
			Confidence:   0.6,
			Strategy:     d.Name(),
			SubDirectory: dir,
			Dependencies: &DependencyInfo{
				Manager: "gomod",
			},
		}
		fns = append(fns, fn)
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "go",
		OverallConfidence: 0.6,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// RustDetector
// ──────────────────────────────────────────────

type RustDetector struct {
	logger *logrus.Logger
}

func (d *RustDetector) Name() string { return "rust-detector" }
func (d *RustDetector) Priority() int { return 20 }

func (d *RustDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	hasCargoToml := false
	var mainRsFiles []string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		if base == "Cargo.toml" {
			hasCargoToml = true
		}
		if base == "main.rs" && strings.Contains(e.Path, "src/") {
			mainRsFiles = append(mainRsFiles, e.Path)
		}
	}

	if !hasCargoToml || len(mainRsFiles) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	for _, ef := range mainRsFiles {
		dir := filepath.Dir(filepath.Dir(ef))
		name := "default"
		if dir != "." {
			name = filepath.Base(dir)
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   ef,
			Runtime:      "rust1.75",
			Confidence:   0.55,
			Strategy:     d.Name(),
			SubDirectory: dir,
			Dependencies: &DependencyInfo{
				Manager: "cargo",
			},
		}
		fns = append(fns, fn)
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "rust",
		OverallConfidence: 0.55,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func stripJSONComments(s string) string {
	var result strings.Builder
	inString := false
	inLineComment := false
	inBlockComment := false
	escaped := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
				result.WriteRune(ch)
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if escaped {
			escaped = false
			result.WriteRune(ch)
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			result.WriteRune(ch)
			continue
		}
		if ch == '"' {
			inString = !inString
			result.WriteRune(ch)
			continue
		}
		if inString {
			result.WriteRune(ch)
			continue
		}
		if ch == '/' && i+1 < len(runes) {
			next := runes[i+1]
			if next == '/' {
				inLineComment = true
				i++
				continue
			}
			if next == '*' {
				inBlockComment = true
				i++
				continue
			}
		}
		result.WriteRune(ch)
	}
	return result.String()
}
