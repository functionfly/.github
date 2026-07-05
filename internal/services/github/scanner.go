package github

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

type ScannerConfig struct {
	AIServiceURL string
}

func NewScanner(client *Client, logger *logrus.Logger) *Scanner {
	return NewScannerWithConfig(client, logger, ScannerConfig{
		AIServiceURL: os.Getenv("AISERVICE_URL"),
	})
}

func NewScannerWithConfig(client *Client, logger *logrus.Logger, config ScannerConfig) *Scanner {
	s := &Scanner{
		client: client,
		logger: logger,
	}
	s.detectors = []Detector{
		&ExplicitConfigDetector{client: client, logger: logger},
		&MonorepoDetector{client: client, logger: logger},
		&ServerlessFrameworkDetector{client: client, logger: logger},
		&AWSLambdaDetector{client: client, logger: logger},
		&AzureFunctionsDetector{client: client, logger: logger},
		&GCFFunctionsDetector{client: client, logger: logger},
		&NextJSDetector{client: client, logger: logger},
		&NestJSDetector{client: client, logger: logger},
		&AstroDetector{logger: logger},
		&DenoDetector{client: client, logger: logger},
		&BunDetector{client: client, logger: logger},
		&ContainerDetector{logger: logger},
		&TerraformDetector{client: client, logger: logger},
		&SignatureDetector{client: client, logger: logger},
		&GitHubActionsDetector{client: client, logger: logger},
		&AIDetector{client: client, logger: logger, aiServiceURL: config.AIServiceURL},
		&NodeDetector{logger: logger},
		&PythonDetector{logger: logger},
		&GoDetector{logger: logger},
		&RustDetector{logger: logger},
	}
	return s
}

func (s *Scanner) ScanRepo(ctx context.Context, owner, repo, branch string) (*ScanResult, error) {
	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	repoInfo, err := s.client.GetRepo(scanCtx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("fetch repo info: %w", err)
	}

	if branch == "" {
		branch = repoInfo.DefaultBranch
	}

	branches, err := s.client.ListBranches(scanCtx, owner, repo)
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

	tree, err := s.client.GetTree(scanCtx, owner, repo, branchSHA, true)
	if err != nil {
		return nil, fmt.Errorf("fetch tree: %w", err)
	}

	languages, err := s.client.GetLanguages(scanCtx, owner, repo)
	if err != nil {
		s.logger.WithError(err).Warn("failed to fetch languages, continuing without")
		languages = make(map[string]float64)
	}

	entries := tree.Tree

	var highConfidenceResult *ScanResult
	var allResults []*ScanResult

	for _, detector := range s.detectors {
		result, err := detector.Detect(scanCtx, repoInfo, entries)
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

	s.resolveRuntimeFromStats(repoInfo, finalResult.Functions, languages)

	return finalResult, nil
}

type langStatsMap map[string]float64

func (l langStatsMap) Dominant() string {
	if len(l) == 0 {
		return ""
	}
	var dominant string
	var maxPct float64
	for lang, pct := range l {
		if pct > maxPct {
			maxPct = pct
			dominant = lang
		}
	}
	return dominant
}

func (l langStatsMap) DefaultRuntime() string {
	dominant := l.Dominant()
	switch strings.ToLower(dominant) {
	case "typescript", "javascript":
		return "node18"
	case "python":
		return "python3.11"
	case "go":
		return "go1.22"
	case "rust":
		return "rust1.75"
	case "java":
		return "java17"
	case "ruby":
		return "ruby3.2"
	case "php":
		return "php8.2"
	case "csharp", "c#":
		return "dotnet6"
	default:
		return dominant
	}
}

func (s *Scanner) resolveRuntimeFromStats(repo *GitHubRepo, detected []DetectedFunction, langStats map[string]float64) {
	if len(detected) == 0 {
		return
	}
	stats := langStatsMap(langStats)
	dominant := stats.Dominant()
	if dominant == "" {
		return
	}
	defaultRuntime := stats.DefaultRuntime()

	for i := range detected {
		if detected[i].Runtime == "unknown" || detected[i].Runtime == "" {
			detected[i].Runtime = defaultRuntime
			s.logger.WithFields(logrus.Fields{
				"function":   detected[i].Name,
				"resolved":   defaultRuntime,
				"from_stats": dominant,
			}).Debug("runtime resolved from language stats")
		}
	}
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

// entryPointPrefixes matches common JavaScript/TypeScript entry point naming patterns.
// Order matters: more specific patterns first to avoid false positives.
var entryPointPrefixes = []string{
	// Standard serverless/function handlers
	"handler", "index", "server", "app", "main", "start", "bootstrap",
	// Framework-specific entry points
	"serve", "api", "worker", "lambda", "entry", "http",
	// CLI tools
	"cli", "bin", "command",
}

var commonSourceDirs = []string{
	"src", "lib", "functions", "api", "handlers", "server", "services",
	"src/handlers", "src/api", "src/functions", "src/server", "src/services",
	"src/lib", "functions/src", "api/src", "pages/api",
}

type NodeDetector struct {
	logger *logrus.Logger
}

func (d *NodeDetector) Name() string { return "node-detector" }
func (d *NodeDetector) Priority() int { return 50 }

func (d *NodeDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	// Track all package.json locations (supports monorepos)
	var packageJSONs []string
	lockfile := ""

	// Build a map of directory -> has package.json for subtree detection
	dirHasPackageJSON := make(map[string]bool)

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)

		switch base {
		case "package.json":
			dir := filepath.Dir(e.Path)
			packageJSONs = append(packageJSONs, e.Path)
			dirHasPackageJSON[dir] = true
		case "package-lock.json":
			if lockfile == "" {
				lockfile = e.Path
			}
		case "yarn.lock":
			if lockfile == "" {
				lockfile = e.Path
			}
		case "pnpm-lock.yaml":
			if lockfile == "" {
				lockfile = e.Path
			}
		case "bun.lockb":
			if lockfile == "" {
				lockfile = e.Path
			}
		}
	}

	if len(packageJSONs) == 0 || lockfile == "" {
		return nil, nil
	}

	// Find entry files by scanning for known patterns
	entryFiles := d.findEntryFiles(entries)

	var fns []DetectedFunction
	for _, ef := range entryFiles {
		fn := d.createDetectedFunction(ef, lockfile)
		fns = append(fns, fn)
	}

	// For monorepos: detect functions from subdirectories with own package.json
	for _, pkgJSON := range packageJSONs {
		pkgDir := filepath.Dir(pkgJSON)
		if pkgDir == "." {
			continue // skip root package.json
		}
		// Check for entry files within this package's subtree
		subEntries := d.filterEntriesForDir(entries, pkgDir)
		subEntryFiles := d.findEntryFilesFromDir(pkgDir, subEntries)
		for _, ef := range subEntryFiles {
			// Avoid duplicates by tracking paths
			alreadyAdded := false
			for _, existing := range fns {
				if existing.EntryPoint == ef {
					alreadyAdded = true
					break
				}
			}
			if alreadyAdded {
				continue
			}
			fn := d.createDetectedFunction(ef, lockfile)
			fn.SubDirectory = pkgDir
			fns = append(fns, fn)
		}
	}

	if len(fns) == 0 {
		return nil, nil
	}

	// Calculate overall confidence based on function count and naming clarity
	avgConfidence := d.calculateAvgConfidence(fns)

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "node",
		OverallConfidence: avgConfidence,
		StrategyUsed:      d.Name(),
	}, nil
}

// findEntryFiles scans all entries for known entry point patterns.
func (d *NodeDetector) findEntryFiles(entries []GitHubTreeEntry) []string {
	var entryFiles []string
	seen := make(map[string]bool)

	// First pass: look for top-level and common directory entry points
	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		path := e.Path

		// Skip non-JS/TS files
		if !isJavaScriptFile(base) {
			continue
		}

		if d.isEntryPoint(base) {
			if !seen[path] {
				entryFiles = append(entryFiles, path)
				seen[path] = true
			}
		}
	}

	// Second pass: look for entry files in common source directories
	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		path := e.Path

		// Skip already found
		if seen[path] {
			continue
		}

		// Check if file is in a common source directory
		if d.isInCommonSourceDir(path) {
			base := filepath.Base(path)
			if isJavaScriptFile(base) && d.isLikelyEntryFilename(base) {
				entryFiles = append(entryFiles, path)
				seen[path] = true
			}
		}
	}

	return entryFiles
}

// filterEntriesForDir returns entries within a specific subdirectory.
func (d *NodeDetector) filterEntriesForDir(entries []GitHubTreeEntry, dir string) []GitHubTreeEntry {
	var filtered []GitHubTreeEntry
	prefix := dir + "/"
	for _, e := range entries {
		if strings.HasPrefix(e.Path, prefix) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// findEntryFilesFromDir searches for entry points in entries already filtered to a directory.
func (d *NodeDetector) findEntryFilesFromDir(dir string, entries []GitHubTreeEntry) []string {
	var entryFiles []string
	seen := make(map[string]bool)

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		if !isJavaScriptFile(base) {
			continue
		}

		// Calculate relative path from dir
		relPath := strings.TrimPrefix(e.Path, dir+"/")

		// Only consider files directly in this directory (not nested)
		if strings.Contains(relPath, "/") {
			continue
		}

		if d.isEntryPoint(base) || d.isLikelyEntryFilename(base) {
			if !seen[e.Path] {
				entryFiles = append(entryFiles, e.Path)
				seen[e.Path] = true
			}
		}
	}

	return entryFiles
}

// isEntryPoint checks if a filename matches known entry point patterns.
func (d *NodeDetector) isEntryPoint(base string) bool {
	// Exact matches for index/handler
	switch base {
	case "index.ts", "index.js", "index.mjs", "index.cjs",
		"handler.ts", "handler.js", "handler.mjs", "handler.cjs":
		return true
	}

	// Prefix matches for other entry patterns
	for _, prefix := range entryPointPrefixes {
		if strings.HasPrefix(base, prefix) {
			ext := filepath.Ext(base)
			if ext == ".ts" || ext == ".js" || ext == ".mjs" || ext == ".cjs" {
				return true
			}
		}
	}

	return false
}

// isLikelyEntryFilename checks if a filename in a subdirectory looks like an entry point.
// This is more permissive than isEntryPoint since we're scanning subdirectories.
func (d *NodeDetector) isLikelyEntryFilename(base string) bool {
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Exact match to entry point names
	for _, prefix := range entryPointPrefixes {
		if name == prefix {
			return true
		}
	}

	return false
}

// isInCommonSourceDir checks if the file path is within a common source directory.
func (d *NodeDetector) isInCommonSourceDir(path string) bool {
	dir := filepath.Dir(path)
	for _, commonDir := range commonSourceDirs {
		if dir == commonDir || strings.HasPrefix(dir, commonDir+"/") {
			return true
		}
	}
	return false
}

// createDetectedFunction creates a DetectedFunction from an entry file path.
func (d *NodeDetector) createDetectedFunction(entryFile, lockfile string) DetectedFunction {
	dir := filepath.Dir(entryFile)
	base := filepath.Base(entryFile)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Determine runtime based on file extension
	runtime := "node18"
	if ext == ".ts" {
		runtime = "node18-typescript"
	} else if ext == ".mjs" || ext == ".cjs" {
		runtime = "node18"
	}

	// Build function name from directory structure + filename
	funcName := d.buildFunctionName(dir, nameWithoutExt)

	// Calculate confidence based on naming clarity
	confidence := d.calculateFunctionConfidence(nameWithoutExt, dir)

	fn := DetectedFunction{
		Name:         funcName,
		EntryPoint:   entryFile,
		Runtime:      runtime,
		Confidence:   confidence,
		Strategy:     d.Name(),
		SubDirectory: dir,
	}

	if lockfile != "" {
		fn.Dependencies = &DependencyInfo{
			Manager:  detectNodePkgManager(lockfile),
			Lockfile: lockfile,
		}
	}

	return fn
}

// buildFunctionName creates a descriptive function name from path components.
func (d *NodeDetector) buildFunctionName(dir, filename string) string {
	// If filename is generic "index", use directory name
	if filename == "index" || filename == "handler" || filename == "server" || filename == "app" {
		if dir != "." && dir != "" {
			return sanitizeFunctionName(filepath.Base(dir))
		}
		return "default"
	}

	// If file is in a common subdirectory, prepend the next directory
	parts := strings.Split(dir, "/")
	for i, part := range parts {
		for _, commonDir := range commonSourceDirs {
			if part == commonDir && i+1 < len(parts) {
				return sanitizeFunctionName(parts[i+1] + "-" + filename)
			}
		}
	}

	// Otherwise use filename as-is
	if dir == "." || dir == "" {
		return sanitizeFunctionName(filename)
	}

	return sanitizeFunctionName(filepath.Base(dir) + "-" + filename)
}

// calculateFunctionConfidence returns a confidence score based on naming clarity.
func (d *NodeDetector) calculateFunctionConfidence(filename, dir string) float64 {
	base := 0.6

	// Strong entry point naming gets higher confidence
	strongNames := map[string]bool{
		"handler": true, "server": true, "app": true, "api": true,
		"worker": true, "lambda": true, "entry": true, "http": true,
	}

	if strongNames[filename] {
		base = 0.75
	}

	// Files in clear function directories get a boost
	if dir != "." && dir != "" {
		funcDirNames := map[string]bool{
			"functions": true, "handlers": true, "api": true,
			"workers": true, "lambdas": true, "services": true,
		}
		if funcDirNames[filepath.Base(dir)] {
			base = 0.8
		}
	}

	// Generic "index" gets slight penalty
	if filename == "index" {
		base -= 0.05
	}

	return base
}

// calculateAvgConfidence computes the average confidence across all detected functions.
func (d *NodeDetector) calculateAvgConfidence(fns []DetectedFunction) float64 {
	if len(fns) == 0 {
		return 0.0
	}
	var sum float64
	for _, fn := range fns {
		sum += fn.Confidence
	}
	avg := sum / float64(len(fns))

	// Boost confidence if we found multiple well-named functions
	if len(fns) >= 2 {
		avg = math.Min(0.9, avg+0.05)
	}

	return avg
}

// sanitizeFunctionName converts a path component to a valid function name.
func sanitizeFunctionName(name string) string {
	// Remove path separators and special chars
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove any non-alphanumeric characters except hyphen
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Collapse multiple hyphens
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens from ends
	name = strings.Trim(name, "-")

	if name == "" {
		return "function"
	}

	return name
}

// isJavaScriptFile checks if the file extension indicates a JavaScript/TypeScript file.
func isJavaScriptFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".js" || ext == ".ts" || ext == ".mjs" || ext == ".cjs" || ext == ".jsx" || ext == ".tsx"
}

// detectNodePkgManager identifies the package manager based on lockfile name.
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
// MonorepoDetector
// ──────────────────────────────────────────────

type MonorepoDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *MonorepoDetector) Name() string { return "monorepo-detector" }
func (d *MonorepoDetector) Priority() int { return 95 }

func (d *MonorepoDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var monorepoConfig string
	var monorepoType string
	var fns []DetectedFunction
	packageDirs := make(map[string]bool)

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		dir := filepath.Dir(e.Path)

		switch base {
		case "pnpm-workspace.yaml":
			monorepoConfig = e.Path
			monorepoType = "pnpm"
			fns = append(fns, d.makeMonorepoFn("pnpm-workspace", e.Path, dir, "pnpm", 0.95))
		case "lerna.json":
			monorepoConfig = e.Path
			monorepoType = "lerna"
			fns = append(fns, d.makeMonorepoFn("lerna-monorepo", e.Path, dir, "npm", 0.9))
		case "turbo.json":
			monorepoConfig = e.Path
			monorepoType = "turbo"
			fns = append(fns, d.makeMonorepoFn("turborepo", e.Path, dir, "npm", 0.95))
		case "nx.json":
			monorepoConfig = e.Path
			monorepoType = "nx"
			fns = append(fns, d.makeMonorepoFn("nx-monorepo", e.Path, dir, "nx", 0.9))
		case "package.json":
			if strings.Contains(dir, "packages/") || dir == "packages" {
				packageDirs[dir] = true
			}
		}
	}

	if monorepoConfig == "" {
		hasRootPackage := hasFile(entries, "package.json")
		hasPackagesDir := false
		for _, e := range entries {
			if e.Type == "blob" && strings.HasPrefix(e.Path, "packages/") {
				hasPackagesDir = true
				break
			}
		}
		if hasRootPackage && hasPackagesDir {
			monorepoConfig = "packages/"
			monorepoType = "manual"
			fns = append(fns, d.makeMonorepoFn("manual-monorepo", "packages/", "packages", "npm", 0.8))
			packageDirs["packages"] = true
		}
	}

	if monorepoConfig == "" {
		return nil, nil
	}

	subPkgs := d.detectSubPackages(ctx, repo, entries, monorepoType)
	fns = append(fns, subPkgs...)

	warnings := []string{"monorepo detected - individual functions will be scanned separately"}
	if len(subPkgs) > 0 {
		warnings = append(warnings, fmt.Sprintf("found %d sub-packages", len(subPkgs)))
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "monorepo",
		OverallConfidence: 0.95,
		StrategyUsed:      d.Name(),
		Warnings:          warnings,
	}, nil
}

func (d *MonorepoDetector) makeMonorepoFn(name, path, subDir, manager string, confidence float64) DetectedFunction {
	return DetectedFunction{
		Name:         name,
		EntryPoint:   path,
		Runtime:      "monorepo",
		Confidence:   confidence,
		Strategy:     d.Name(),
		SubDirectory: subDir,
		Dependencies: &DependencyInfo{
			Manager: manager,
		},
	}
}

func (d *MonorepoDetector) detectSubPackages(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry, monorepoType string) []DetectedFunction {
	var pkgs []DetectedFunction
	seen := make(map[string]bool)

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		if base != "package.json" && base != "Cargo.toml" && base != "go.mod" {
			continue
		}

		dir := filepath.Dir(e.Path)
		if dir == "" || dir == "." {
			continue
		}

		if !strings.Contains(dir, "packages/") && !strings.Contains(dir, "apps/") && !strings.Contains(dir, "services/") {
			continue
		}

		if seen[dir] {
			continue
		}
		seen[dir] = true

		var runtime, manager string
		switch base {
		case "package.json":
			runtime = "node"
			manager = "npm"
		case "Cargo.toml":
			runtime = "rust"
			manager = "cargo"
		case "go.mod":
			runtime = "go"
			manager = "gomod"
		}

		pkgName := filepath.Base(dir)
		pkgs = append(pkgs, DetectedFunction{
			Name:         pkgName,
			EntryPoint:   e.Path,
			Runtime:      runtime,
			Confidence:   0.7,
			Strategy:     d.Name() + "-subpackage",
			SubDirectory: dir,
			Dependencies: &DependencyInfo{
				Manager: manager,
				Packages: []string{pkgName},
			},
			Manifest: map[string]interface{}{
				"sub_package":  true,
				"monorepo_type": monorepoType,
			},
		})
	}

	return pkgs
}

func hasFile(entries []GitHubTreeEntry, filename string) bool {
	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		if filepath.Base(e.Path) == filename || e.Path == filename {
			return true
		}
	}
	return false
}

// ──────────────────────────────────────────────
// AWSLambdaDetector
// ──────────────────────────────────────────────

type AWSLambdaDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *AWSLambdaDetector) Name() string { return "aws-lambda-detector" }
func (d *AWSLambdaDetector) Priority() int { return 85 }

func (d *AWSLambdaDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var serverlessYAML string
	var hasPowertools bool
	var zipPackages []string
	var fns []DetectedFunction

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		dir := filepath.Dir(e.Path)

		if base == "serverless.yml" || base == "serverless.yaml" {
			serverlessYAML = e.Path
		}

		if strings.Contains(e.Path, "aws-lambda-powertools") ||
			strings.Contains(e.Path, "@aws-lambda-powertools") {
			hasPowertools = true
		}

		if strings.HasSuffix(base, ".zip") {
			zipPackages = append(zipPackages, e.Path)
		}

		if base == "template.yaml" || base == "template.yml" {
			fns = append(fns, DetectedFunction{
				Name:         "sam-template",
				EntryPoint:   e.Path,
				Runtime:      "aws-sam",
				Confidence:   0.8,
				Strategy:     d.Name(),
				SubDirectory: dir,
			})
		}
	}

	if serverlessYAML != "" && !hasPowertools {
		raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, serverlessYAML, "")
		if err == nil {
			content := string(raw)
			if strings.Contains(content, "provider: aws") || strings.Contains(content, "provider:\n    name: aws") {
				lines := strings.Split(content, "\n")
				inFunctions := false
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed == "functions:" || strings.HasPrefix(trimmed, "functions:") {
						inFunctions = true
						continue
					}
					if inFunctions {
						if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" {
							break
						}
						if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "handler:") && !strings.Contains(trimmed, "runtime:") {
							fnName := strings.TrimSuffix(trimmed, ":")
							fnName = strings.TrimSpace(fnName)
							if fnName != "" && fnName != "functions" {
								fns = append(fns, DetectedFunction{
									Name:         fnName,
									EntryPoint:   fnName,
									Runtime:      "aws-lambda",
									Confidence:   0.85,
									Strategy:     d.Name(),
								})
							}
						}
					}
				}
			}
		}
	}

	if len(fns) > 0 {
		return &ScanResult{
			Functions:         fns,
			PrimaryRuntime:    "aws-lambda",
			OverallConfidence: 0.85,
			StrategyUsed:      d.Name(),
		}, nil
	}

	if hasPowertools || len(zipPackages) > 0 {
		return &ScanResult{
			Functions: []DetectedFunction{
				{
					Name:         "aws-lambda-function",
					EntryPoint:   "src/",
					Runtime:      "aws-lambda",
					Confidence:   0.7,
					Strategy:     d.Name(),
				},
			},
			PrimaryRuntime:    "aws-lambda",
			OverallConfidence: 0.7,
			StrategyUsed:      d.Name(),
		}, nil
	}

	return nil, nil
}

// ──────────────────────────────────────────────
// AzureFunctionsDetector
// ──────────────────────────────────────────────

type AzureFunctionsDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *AzureFunctionsDetector) Name() string { return "azure-functions-detector" }
func (d *AzureFunctionsDetector) Priority() int { return 85 }

func (d *AzureFunctionsDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hostJSONPath string
	var functionConfigs []string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		dir := filepath.Dir(e.Path)

		if base == "host.json" {
			hostJSONPath = e.Path
		}

		if base == "function.json" && (strings.Contains(dir, "TimerTrigger") ||
			strings.Contains(dir, "HttpTrigger") ||
			strings.Contains(dir, "QueueTrigger") ||
			strings.Contains(dir, "BlobTrigger")) {
			functionConfigs = append(functionConfigs, e.Path)
		}

		if strings.Contains(base, "TimerTrigger") && strings.HasSuffix(base, ".cs") ||
			strings.Contains(base, "TimerTrigger") && strings.HasSuffix(base, ".py") ||
			strings.Contains(base, "TimerTrigger") && strings.HasSuffix(base, ".js") ||
			strings.Contains(base, "HttpTrigger") && strings.HasSuffix(base, ".cs") ||
			strings.Contains(base, "HttpTrigger") && strings.HasSuffix(base, ".py") ||
			strings.Contains(base, "HttpTrigger") && strings.HasSuffix(base, ".js") {
			if !contains(functionConfigs, e.Path) {
				functionConfigs = append(functionConfigs, e.Path)
			}
		}
	}

	if hostJSONPath == "" && len(functionConfigs) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	seen := make(map[string]bool)

	for _, fc := range functionConfigs {
		dir := filepath.Dir(fc)
		name := filepath.Base(dir)
		key := name + ":" + dir
		if seen[key] {
			continue
		}
		seen[key] = true

		var triggerType string
		if strings.Contains(fc, "TimerTrigger") {
			triggerType = "TimerTrigger"
		} else if strings.Contains(fc, "HttpTrigger") {
			triggerType = "HttpTrigger"
		} else if strings.Contains(fc, "QueueTrigger") {
			triggerType = "QueueTrigger"
		} else if strings.Contains(fc, "BlobTrigger") {
			triggerType = "BlobTrigger"
		}

		runtime := "azure-functions"
		if strings.HasSuffix(fc, ".py") {
			runtime = "azure-functions-python"
		} else if strings.HasSuffix(fc, ".js") {
			runtime = "azure-functions-node"
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   fc,
			Runtime:      runtime,
			Confidence:   0.85,
			Strategy:     d.Name(),
			SubDirectory: dir,
			Manifest: map[string]interface{}{
				"trigger_type": triggerType,
			},
		}
		fns = append(fns, fn)
	}

	if len(fns) == 0 && hostJSONPath != "" {
		fns = append(fns, DetectedFunction{
			Name:         "azure-functions-app",
			EntryPoint:   hostJSONPath,
			Runtime:      "azure-functions",
			Confidence:   0.6,
			Strategy:     d.Name(),
			SubDirectory: filepath.Dir(hostJSONPath),
		})
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "azure-functions",
		OverallConfidence: 0.85,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// GCFFunctionsDetector
// ──────────────────────────────────────────────

type GCFFunctionsDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *GCFFunctionsDetector) Name() string { return "gcf-detector" }
func (d *GCFFunctionsDetector) Priority() int { return 85 }

func (d *GCFFunctionsDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hasFunctionsFramework bool
	var mainPyPath string
	var fns []DetectedFunction

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		dir := filepath.Dir(e.Path)

		if base == "package.json" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "@google-cloud/functions-framework") {
					hasFunctionsFramework = true
					fns = append(fns, DetectedFunction{
						Name:         "gcf-node-function",
						EntryPoint:   e.Path,
						Runtime:      "gcf-node",
						Confidence:   0.85,
						Strategy:     d.Name(),
						SubDirectory: dir,
					})
				}
			}
		}

		if base == "main.py" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "functions_framework") ||
					strings.Contains(content, "google.cloud.functions") {
					mainPyPath = e.Path
					fns = append(fns, DetectedFunction{
						Name:         "gcf-python-function",
						EntryPoint:   e.Path,
						Runtime:      "gcf-python",
						Confidence:   0.85,
						Strategy:     d.Name(),
						SubDirectory: dir,
					})
				}
			}
		}

		if base == "requirements.txt" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "functions-framework") ||
					strings.Contains(content, "google-cloud-functions") {
					hasFunctionsFramework = true
				}
			}
		}
	}

	if !hasFunctionsFramework && mainPyPath == "" {
		return nil, nil
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "gcf",
		OverallConfidence: 0.85,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// NestJSDetector
// ──────────────────────────────────────────────

type NestJSDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *NestJSDetector) Name() string { return "nestjs-detector" }
func (d *NestJSDetector) Priority() int { return 60 }

func (d *NestJSDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hasNestJSCore bool
	var hasPlatformExpress bool
	var packageJSONPath string
	var mainTsPath string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)

		if base == "package.json" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "@nestjs/core") {
					hasNestJSCore = true
				}
				if strings.Contains(content, "@nestjs/platform-express") {
					hasPlatformExpress = true
				}
				packageJSONPath = e.Path
			}
		}

		if base == "main.ts" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "@nestjs/core") ||
					strings.Contains(content, "NestFactory") {
					mainTsPath = e.Path
				}
			}
		}
	}

	if !hasNestJSCore || !hasPlatformExpress {
		return nil, nil
	}

	var fns []DetectedFunction
	if mainTsPath != "" {
		dir := filepath.Dir(mainTsPath)
		fns = append(fns, DetectedFunction{
			Name:         "nestjs-application",
			EntryPoint:   mainTsPath,
			Runtime:      "node18-typescript",
			Confidence:   0.85,
			Strategy:     d.Name(),
			SubDirectory: dir,
			Dependencies: &DependencyInfo{
				Manager: "npm",
			},
		})
	} else if packageJSONPath != "" {
		dir := filepath.Dir(packageJSONPath)
		fns = append(fns, DetectedFunction{
			Name:         "nestjs-application",
			EntryPoint:   packageJSONPath,
			Runtime:      "node18-typescript",
			Confidence:   0.7,
			Strategy:     d.Name(),
			SubDirectory: dir,
			Dependencies: &DependencyInfo{
				Manager: "npm",
			},
		})
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "node",
		OverallConfidence: 0.8,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// NextJSDetector
// ──────────────────────────────────────────────

type NextJSDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *NextJSDetector) Name() string { return "nextjs-detector" }
func (d *NextJSDetector) Priority() int { return 55 }

func (d *NextJSDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hasNextConfig bool
	var hasPagesAPI bool
	var hasAppAPI bool
	var fns []DetectedFunction

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		path := e.Path

		if base == "next.config.js" || base == "next.config.mjs" ||
			base == "next.config.ts" || base == "next.config.cjs" {
			hasNextConfig = true
		}

		if strings.HasPrefix(path, "pages/api/") && strings.HasSuffix(base, ".ts") ||
			strings.HasPrefix(path, "pages/api/") && strings.HasSuffix(base, ".js") ||
			strings.HasPrefix(path, "pages/api/") && strings.HasSuffix(base, ".tsx") ||
			strings.HasPrefix(path, "pages/api/") && strings.HasSuffix(base, ".jsx") {
			hasPagesAPI = true
			dir := filepath.Dir(path)
			name := strings.TrimPrefix(dir, "pages/")
			fns = append(fns, DetectedFunction{
				Name:         name,
				EntryPoint:   path,
				Runtime:      "nextjs",
				Confidence:   0.7,
				Strategy:     d.Name(),
				SubDirectory: dir,
			})
		}

		if strings.HasPrefix(path, "app/api/") && strings.HasSuffix(base, "/route.ts") ||
			strings.HasPrefix(path, "app/api/") && strings.HasSuffix(base, "/route.js") ||
			strings.HasPrefix(path, "app/api/") && strings.HasSuffix(base, "/route.tsx") ||
			strings.HasPrefix(path, "app/api/") && strings.HasSuffix(base, "/route.js") {
			hasAppAPI = true
			dir := filepath.Dir(path)
			name := "app" + strings.TrimPrefix(dir, "app")
			fns = append(fns, DetectedFunction{
				Name:         name,
				EntryPoint:   path,
				Runtime:      "nextjs",
				Confidence:   0.8,
				Strategy:     d.Name(),
				SubDirectory: dir,
			})
		}
	}

	if !hasNextConfig && !hasPagesAPI && !hasAppAPI {
		return nil, nil
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "nextjs",
		OverallConfidence: 0.75,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// AstroDetector
// ──────────────────────────────────────────────

type AstroDetector struct {
	logger *logrus.Logger
}

func (d *AstroDetector) Name() string { return "astro-detector" }
func (d *AstroDetector) Priority() int { return 45 }

func (d *AstroDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hasAstroConfig bool
	var hasPagesDir bool

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)

		if base == "astro.config.mjs" || base == "astro.config.js" ||
			base == "astro.config.ts" || base == "astro.config.cjs" {
			hasAstroConfig = true
		}

		if strings.HasPrefix(e.Path, "src/pages/") && e.Type == "blob" {
			hasPagesDir = true
		}
	}

	if !hasAstroConfig && !hasPagesDir {
		return nil, nil
	}

	fns := []DetectedFunction{
		{
			Name:         "astro-site",
			EntryPoint:   "src/pages/",
			Runtime:      "astro",
			Confidence:   0.7,
			Strategy:     d.Name(),
		},
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "astro",
		OverallConfidence: 0.7,
		StrategyUsed:      d.Name(),
		Warnings:          []string{"Astro is a static site generator; serverless functions may require adapter configuration"},
	}, nil
}

// ──────────────────────────────────────────────
// DenoDetector
// ──────────────────────────────────────────────

type DenoDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *DenoDetector) Name() string { return "deno-detector" }
func (d *DenoDetector) Priority() int { return 45 }

func (d *DenoDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hasDenoConfig bool
	var hasFreshEntry bool
	var hasDenoServe bool
	var fns []DetectedFunction

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		dir := filepath.Dir(e.Path)

		if base == "deno.json" || base == "deno.jsonc" {
			hasDenoConfig = true
			fns = append(fns, DetectedFunction{
				Name:         "deno-project",
				EntryPoint:   e.Path,
				Runtime:      "deno",
				Confidence:   0.7,
				Strategy:     d.Name(),
				SubDirectory: dir,
				Dependencies: &DependencyInfo{
					Manager: "deno",
				},
			})
		}

		if base == "fresh.ts" || base == "fresh.js" {
			hasFreshEntry = true
			fns = append(fns, DetectedFunction{
				Name:         "fresh-app",
				EntryPoint:   e.Path,
				Runtime:      "deno-fresh",
				Confidence:   0.85,
				Strategy:     d.Name(),
				SubDirectory: dir,
			})
		}

		if base == "main.ts" || base == "main.js" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "Deno.serve") ||
					strings.Contains(content, "serve(") {
					hasDenoServe = true
					fns = append(fns, DetectedFunction{
						Name:         "deno-server",
						EntryPoint:   e.Path,
						Runtime:      "deno",
						Confidence:   0.75,
						Strategy:     d.Name(),
						SubDirectory: dir,
					})
				}
			}
		}
	}

	if !hasDenoConfig && !hasFreshEntry && !hasDenoServe {
		return nil, nil
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "deno",
		OverallConfidence: 0.75,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// BunDetector
// ──────────────────────────────────────────────

type BunDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *BunDetector) Name() string { return "bun-detector" }
func (d *BunDetector) Priority() int { return 45 }

func (d *BunDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var hasBunLock bool
	var hasBunServe bool
	var fns []DetectedFunction

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)
		dir := filepath.Dir(e.Path)

		if base == "bun.lockb" {
			hasBunLock = true
			fns = append(fns, DetectedFunction{
				Name:         "bun-project",
				EntryPoint:   e.Path,
				Runtime:      "bun",
				Confidence:   0.6,
				Strategy:     d.Name(),
				SubDirectory: dir,
				Dependencies: &DependencyInfo{
					Manager: "bun",
				},
			})
		}

		if base == "index.ts" || base == "index.js" || base == "main.ts" || base == "main.js" {
			raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "Bun.serve") {
					hasBunServe = true
					fns = append(fns, DetectedFunction{
						Name:         "bun-server",
						EntryPoint:   e.Path,
						Runtime:      "bun",
						Confidence:   0.75,
						Strategy:     d.Name(),
						SubDirectory: dir,
					})
				}
			}
		}
	}

	if !hasBunLock && !hasBunServe {
		return nil, nil
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "bun",
		OverallConfidence: 0.7,
		StrategyUsed:      d.Name(),
	}, nil
}

// ──────────────────────────────────────────────
// ContainerDetector
// ──────────────────────────────────────────────

type ContainerDetector struct {
	logger *logrus.Logger
}

func (d *ContainerDetector) Name() string { return "container-detector" }
func (d *ContainerDetector) Priority() int { return 25 }

func (d *ContainerDetector) Detect(_ context.Context, _ *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var dockerfilePath string
	var dockerComposePath string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)

		if base == "Dockerfile" || base == "Dockerfile.dev" || base == "Dockerfile.prod" {
			dockerfilePath = e.Path
		}

		if base == "docker-compose.yml" || base == "docker-compose.yaml" {
			raw, err := readFileFromEntries(entries, e.Path)
			if err == nil {
				content := string(raw)
				if strings.Contains(content, "function") ||
					strings.Contains(content, "serverless") ||
					strings.Contains(content, "lambda") {
					dockerComposePath = e.Path
				}
			}
		}
	}

	if dockerfilePath == "" && dockerComposePath == "" {
		return nil, nil
	}

	var fns []DetectedFunction
	if dockerfilePath != "" {
		dir := filepath.Dir(dockerfilePath)
		fns = append(fns, DetectedFunction{
			Name:         "container-function",
			EntryPoint:   dockerfilePath,
			Runtime:      "container",
			Confidence:   0.5,
			Strategy:     d.Name(),
			SubDirectory: dir,
		})
	}
	if dockerComposePath != "" {
		dir := filepath.Dir(dockerComposePath)
		fns = append(fns, DetectedFunction{
			Name:         "container-compose",
			EntryPoint:   dockerComposePath,
			Runtime:      "container",
			Confidence:   0.5,
			Strategy:     d.Name(),
			SubDirectory: dir,
		})
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "container",
		OverallConfidence: 0.5,
		StrategyUsed:      d.Name(),
		Warnings:          []string{"Container-based functions require custom deployment configuration"},
	}, nil
}

// ──────────────────────────────────────────────
// TerraformDetector
// ──────────────────────────────────────────────

type TerraformDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *TerraformDetector) Name() string { return "terraform-detector" }
func (d *TerraformDetector) Priority() int { return 20 }

func (d *TerraformDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var tfFiles []string

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)

		if strings.HasSuffix(base, ".tf") {
			tfFiles = append(tfFiles, e.Path)
		}
	}

	if len(tfFiles) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	for _, tf := range tfFiles {
		raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, tf, "")
		if err != nil {
			continue
		}
		content := string(raw)
		if strings.Contains(content, `resource "aws_lambda_function"`) ||
			strings.Contains(content, "resource \"aws_lambda_function\"") ||
			strings.Contains(content, "google_cloudfunctions_function") ||
			strings.Contains(content, "azurerm_function_app") {

			dir := filepath.Dir(tf)
			name := filepath.Base(tf)
			fns = append(fns, DetectedFunction{
				Name:         name,
				EntryPoint:   tf,
				Runtime:      "terraform",
				Confidence:   0.8,
				Strategy:     d.Name(),
				SubDirectory: dir,
				Manifest: map[string]interface{}{
					"infrastructure": true,
				},
			})
		}
	}

	if len(fns) == 0 {
		return nil, nil
	}

	return &ScanResult{
		Functions:         fns,
		PrimaryRuntime:    "terraform",
		OverallConfidence: 0.8,
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func readFileFromEntries(entries []GitHubTreeEntry, path string) ([]byte, error) {
	for _, e := range entries {
		if e.Path == path && e.Type == "blob" {
			return []byte{}, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// ──────────────────────────────────────────────
// GitHubActionsDetector
// ──────────────────────────────────────────────

type GitHubActionsDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *GitHubActionsDetector) Name() string { return "github-actions-detector" }
func (d *GitHubActionsDetector) Priority() int { return 10 }

func (d *GitHubActionsDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var workflows []Workflow
	seen := make(map[string]bool)

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		path := e.Path
		if !strings.HasPrefix(path, ".github/workflows/") {
			continue
		}
		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			continue
		}
		if seen[path] {
			continue
		}
		seen[path] = true

		content, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, path, "")
		if err != nil {
			continue
		}

		wf := d.parseWorkflow(string(content), path)
		if wf.Name == "" {
			wf.Name = filepath.Base(filepath.Dir(path)) + "-" + filepath.Base(path)
		}
		wf.Path = path
		workflows = append(workflows, wf)
	}

	if len(workflows) == 0 {
		return nil, nil
	}

	var fns []DetectedFunction
	for _, wf := range workflows {
		manifest := map[string]interface{}{
			"workflow_name": wf.Name,
			"events":        wf.Events,
			"jobs":           wf.Jobs,
			"is_deploy":      wf.IsDeploy,
			"deploy_type":    wf.DeployType,
		}
		fns = append(fns, DetectedFunction{
			Name:         "workflow-" + wf.Name,
			EntryPoint:   wf.Path,
			Runtime:      "github-actions",
			Confidence:   0.6,
			Strategy:     d.Name(),
			SubDirectory: filepath.Dir(wf.Path),
			Manifest:     manifest,
		})
	}

	return &ScanResult{
		Functions:         fns,
		OverallConfidence: 0.6,
		StrategyUsed:      d.Name(),
		Warnings:          d.suggestAutoSync(workflows),
	}, nil
}

func (d *GitHubActionsDetector) parseWorkflow(content, path string) Workflow {
	wf := Workflow{
		Path:   path,
		Events: []string{},
		Jobs:   []WorkflowJob{},
	}

	lines := strings.Split(content, "\n")
	inJobs := false
	var currentJob WorkflowJob
	isDeployWorkflow := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "on:") || strings.HasPrefix(trimmed, "on ") {
			events := d.extractEvents(trimmed)
			wf.Events = append(wf.Events, events...)
		}

		if trimmed == "jobs:" {
			inJobs = true
			continue
		}

		if inJobs && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "steps:") {
			if currentJob.Name != "" {
				wf.Jobs = append(wf.Jobs, currentJob)
			}
			currentJob = WorkflowJob{
				Name:      strings.TrimSuffix(trimmed, ":"),
				StepNames: []string{},
			}
		}

		if inJobs && strings.Contains(trimmed, "- name:") {
			stepName := strings.TrimPrefix(trimmed, "- name:")
			stepName = strings.TrimSpace(stepName)
			currentJob.StepNames = append(currentJob.StepNames, stepName)

			lower := strings.ToLower(stepName)
			if strings.Contains(lower, "deploy") || strings.Contains(lower, "publish") ||
				strings.Contains(lower, "release") || strings.Contains(lower, "upload") {
				isDeployWorkflow = true
				if strings.Contains(lower, "lambda") || strings.Contains(lower, "serverless") {
					wf.DeployType = "lambda"
				} else if strings.Contains(lower, "azure") || strings.Contains(lower, "function") {
					wf.DeployType = "azure-functions"
				} else if strings.Contains(lower, "gcf") || strings.Contains(lower, "cloudfunctions") {
					wf.DeployType = "gcf"
				} else if strings.Contains(lower, "docker") || strings.Contains(lower, "container") {
					wf.DeployType = "container"
				} else {
					wf.DeployType = "generic"
				}
			}
		}
	}

	if currentJob.Name != "" {
		wf.Jobs = append(wf.Jobs, currentJob)
	}

	wf.IsDeploy = isDeployWorkflow
	return wf
}

func (d *GitHubActionsDetector) extractEvents(onLine string) []string {
	onLine = strings.TrimPrefix(onLine, "on:")
	onLine = strings.TrimSpace(onLine)
	if onLine == "" {
		return nil
	}

	events := []string{}
	parts := strings.Split(onLine, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			events = append(events, p)
		}
	}
	return events
}

func (d *GitHubActionsDetector) suggestAutoSync(workflows []Workflow) []string {
	var warnings []string
	for _, wf := range workflows {
		if wf.IsDeploy {
			warnings = append(warnings, fmt.Sprintf("workflow %q deploys via %s - consider auto-sync", wf.Name, wf.DeployType))
		}
	}
	return warnings
}

// ──────────────────────────────────────────────
// AIDetector
// ──────────────────────────────────────────────

type AIDetector struct {
	client      *Client
	logger      *logrus.Logger
	aiServiceURL string
}

func (d *AIDetector) Name() string { return "ai-detector" }
func (d *AIDetector) Priority() int { return 5 }

func (d *AIDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	if d.aiServiceURL == "" {
		return nil, nil
	}

	codeFiles := d.filterCodeFiles(entries, 20)
	if len(codeFiles) == 0 {
		return nil, nil
	}

	prompt := d.buildPrompt(repo, codeFiles)
	resp, err := d.callAIService(ctx, prompt)
	if err != nil {
		d.logger.WithError(err).Warn("AI detection failed, skipping")
		return nil, nil
	}

	return d.parseAIResponse(resp)
}

func (d *AIDetector) filterCodeFiles(entries []GitHubTreeEntry, limit int) []string {
	var files []string
	exts := map[string]bool{
		".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true,
		".py": true, ".go": true, ".rs": true, ".java": true,
	}

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		ext := filepath.Ext(e.Path)
		if !exts[ext] {
			continue
		}
		if strings.Contains(e.Path, "node_modules/") || strings.Contains(e.Path, ".git/") {
			continue
		}
		files = append(files, e.Path)
		if len(files) >= limit {
			break
		}
	}
	return files
}

func (d *AIDetector) buildPrompt(repo *GitHubRepo, files []string) string {
	return fmt.Sprintf(`Analyze this repository for serverless functions.
Repository: %s/%s
Files: %v

Identify:
1. Entry point functions (handlers)
2. Runtime/framework used
3. Event sources (HTTP, timer, queue, etc.)
4. Dependencies and their purpose

Return JSON with: {"functions": [{"name": "...", "entry_point": "...", "runtime": "...", "event_sources": [...], "confidence": 0.0-1.0}]}`, repo.Owner.Login, repo.Name, files)
}

func (d *AIDetector) callAIService(ctx context.Context, prompt string) ([]byte, error) {
	type request struct {
		Prompt string `json:"prompt"`
	}
	type response struct {
		Functions []DetectedFunction `json:"functions"`
	}

	reqBody, _ := json.Marshal(request{Prompt: prompt})
	req, err := http.NewRequestWithContext(ctx, "POST", d.aiServiceURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned %d", resp.StatusCode)
	}

	var aiResp response
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, err
	}

	return json.Marshal(aiResp)
}

func (d *AIDetector) parseAIResponse(data []byte) (*ScanResult, error) {
	type aiResponse struct {
		Functions []DetectedFunction `json:"functions"`
	}

	var resp aiResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Functions) == 0 {
		return nil, nil
	}

	for i := range resp.Functions {
		resp.Functions[i].Strategy = d.Name()
		if resp.Functions[i].Confidence == 0 {
			resp.Functions[i].Confidence = 0.6
		}
	}

	return &ScanResult{
		Functions:         resp.Functions,
		OverallConfidence: 0.6,
		StrategyUsed:      d.Name(),
		Warnings:          []string{"AI-detected functions - verify manually"},
	}, nil
}

// ──────────────────────────────────────────────
// SignatureDetector
// ──────────────────────────────────────────────

type SignatureDetector struct {
	client *Client
	logger *logrus.Logger
}

func (d *SignatureDetector) Name() string { return "signature-detector" }
func (d *SignatureDetector) Priority() int { return 15 }

func (d *SignatureDetector) Detect(ctx context.Context, repo *GitHubRepo, entries []GitHubTreeEntry) (*ScanResult, error) {
	var fns []DetectedFunction
	seen := make(map[string]bool)

	for _, e := range entries {
		if e.Type != "blob" {
			continue
		}
		base := filepath.Base(e.Path)

		if !d.isCodeFile(base) {
			continue
		}

		raw, err := d.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, e.Path, "")
		if err != nil {
			continue
		}

		content := string(raw)
		handlerName, runtime := d.AnalyzeContent(e.Path, []byte(content))
		if handlerName == "" {
			continue
		}

		key := e.Path + ":" + handlerName
		if seen[key] {
			continue
		}
		seen[key] = true

		dir := filepath.Dir(e.Path)
		name := handlerName
		if dir != "." && dir != "" {
			name = filepath.Base(dir) + "-" + handlerName
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   e.Path,
			Runtime:      runtime,
			Confidence:   0.7,
			Strategy:     d.Name(),
			SubDirectory: dir,
			Manifest: map[string]interface{}{
				"handler_name": handlerName,
			},
		}
		fns = append(fns, fn)
	}

	if len(fns) == 0 {
		return nil, nil
	}

	return &ScanResult{
		Functions:         fns,
		OverallConfidence: 0.7,
		StrategyUsed:      d.Name(),
	}, nil
}

func (d *SignatureDetector) isCodeFile(filename string) bool {
	exts := map[string]bool{
		".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true,
		".py": true, ".go": true, ".rs": true, ".java": true, ".rb": true,
		".php": true, ".cs": true, ".fs": true, ".c": true, ".cpp": true,
	}
	ext := filepath.Ext(filename)
	return exts[ext]
}

func (d *SignatureDetector) AnalyzeContent(path string, content []byte) (string, string) {
	str := string(content)

	switch {
	case d.containsAny(str, "export async function handler", "exports.handler", "module.exports.handler", "export default function handler"):
		return "handler", "nodejs"
	case d.containsAny(str, "def lambda_handler", "def handle", "async def lambda_handler"):
		return "lambda_handler", "python"
	case d.containsAny(str, "func HandleRequest", "func Handle", "func handle", "func Process"):
		if strings.Contains(str, "HandleRequest") {
			return "HandleRequest", "go"
		}
		return "Handle", "go"
	case d.containsAny(str, "pub async fn handle", "pub fn handle", "async fn handle", "#[tokio::main]"):
		return "handle", "rust"
	case d.containsAny(str, "export default async function handler", "export { handler", "export const handler"):
		return "handler", "nodejs"
	case d.containsAny(str, "public class Function", "public class Handler", "@FunctionalInterface"):
		return "handleRequest", "java"
	case d.containsAny(str, "module.exports = async", "module.exports = function", "exports.serve"):
		return "handler", "nodejs"
	case d.containsAny(str, "async def handler(event", "def handler(event", "async def handle(event"):
		return "handler", "python"
	case d.containsAny(str, "func(w http.ResponseWriter", "http.HandlerFunc", "func Handle(w http"):
		return "httpHandler", "go"
	case d.containsAny(str, "addEventListener", "onclick", "on('fetch'", "addEventListener('fetch'"):
		return "fetchHandler", "nodejs"
	case d.containsAny(str, "export { default }", "export default", "export class"):
		return "handler", "nodejs"
	}
	return "", ""
}

func (d *SignatureDetector) containsAny(str string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.Contains(str, p) {
			return true
		}
	}
	return false
}

// ──────────────────────────────────────────────
// Incremental Scanning
// ──────────────────────────────────────────────

type ChangedFile struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	OldSHA    string `json:"old_sha,omitempty"`
	NewSHA    string `json:"new_sha,omitempty"`
	Patch     string `json:"patch,omitempty"`
	OldObject string `json:"old_object,omitempty"`
	NewObject string `json:"new_object,omitempty"`
}

func (s *Scanner) IncrementalScan(ctx context.Context, repo *GitHubRepo, before, after string) (*ScanResult, error) {
	s.logger.WithFields(logrus.Fields{
		"repo":    repo.FullName,
		"before":  before,
		"after":   after,
	}).Info("starting incremental scan")

	diffResult, err := s.client.GetCompareDiff(ctx, repo.Owner.Login, repo.Name, before, after)
	if err != nil {
		s.logger.WithError(err).Warn("incremental scan failed, falling back to full scan")
		return s.ScanRepo(ctx, repo.Owner.Login, repo.Name, "")
	}

	changedFiles := s.parseChangedFilesFromDiff(diffResult)
	if len(changedFiles) == 0 {
		return &ScanResult{
			Functions:         []DetectedFunction{},
			FilesScanned:       0,
			ScanMode:           "incremental",
			OverallConfidence:  1.0,
			StrategyUsed:       "no-changes",
		}, nil
	}

	affectedFunctions, filesScanned := s.detectInChangedFiles(ctx, repo, changedFiles)

	result := &ScanResult{
		Functions:         affectedFunctions,
		FilesScanned:      filesScanned,
		ScanMode:          "incremental",
		OverallConfidence: 0.75,
		StrategyUsed:      "incremental",
	}

	if len(affectedFunctions) == 0 {
		result.Warnings = []string{"changed files detected but no function signatures found"}
	}

	return result, nil
}

func (s *Scanner) parseChangedFilesFromDiff(diffResult map[string]interface{}) []ChangedFile {
	var changed []ChangedFile

	if diffResult == nil {
		return changed
	}

	if commits, ok := diffResult["commits"].([]interface{}); ok {
		for _, c := range commits {
			if commit, ok := c.(map[string]interface{}); ok {
				if added, ok := commit["added"].([]interface{}); ok {
					for _, f := range added {
						if path, ok := f.(string); ok {
							changed = append(changed, ChangedFile{Path: path, Status: "added"})
						}
					}
				}
				if removed, ok := commit["removed"].([]interface{}); ok {
					for _, f := range removed {
						if path, ok := f.(string); ok {
							changed = append(changed, ChangedFile{Path: path, Status: "removed"})
						}
					}
				}
				if modified, ok := commit["modified"].([]interface{}); ok {
					for _, f := range modified {
						if path, ok := f.(string); ok {
							changed = append(changed, ChangedFile{Path: path, Status: "modified"})
						}
					}
				}
			}
		}
	}

	if files, ok := diffResult["files"].([]interface{}); ok {
		for _, f := range files {
			if file, ok := f.(map[string]interface{}); ok {
				cf := ChangedFile{}
				if filename, ok := file["filename"].(string); ok {
					cf.Path = filename
				}
				if status, ok := file["status"].(string); ok {
					cf.Status = status
				}
				if patch, ok := file["patch"].(string); ok {
					cf.Patch = patch
				}
				changed = append(changed, cf)
			}
		}
	}

	seen := make(map[string]bool)
	var unique []ChangedFile
	for _, f := range changed {
		if !seen[f.Path] {
			seen[f.Path] = true
			unique = append(unique, f)
		}
	}

	return unique
}

func (s *Scanner) detectInChangedFiles(ctx context.Context, repo *GitHubRepo, changedFiles []ChangedFile) ([]DetectedFunction, int) {
	var affected []DetectedFunction
	filesScanned := 0

	signatureDetector := &SignatureDetector{client: s.client, logger: s.logger}

	for _, cf := range changedFiles {
		if cf.Status == "removed" {
			continue
		}

		base := filepath.Base(cf.Path)
		if !signatureDetector.isCodeFile(base) {
			continue
		}

		filesScanned++

		raw, err := s.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, cf.Path, "")
		if err != nil {
			continue
		}

		handlerName, runtime := signatureDetector.AnalyzeContent(cf.Path, raw)
		if handlerName == "" {
			continue
		}

		dir := filepath.Dir(cf.Path)
		name := handlerName
		if dir != "." && dir != "" {
			name = filepath.Base(dir) + "-" + handlerName
		}

		fn := DetectedFunction{
			Name:         name,
			EntryPoint:   cf.Path,
			Runtime:      runtime,
			Confidence:   0.75,
			Strategy:     "incremental-signature",
			SubDirectory: dir,
			Manifest: map[string]interface{}{
				"handler_name": handlerName,
				"change_type":  cf.Status,
			},
		}
		affected = append(affected, fn)
	}

	return affected, filesScanned
}

// ──────────────────────────────────────────────
// Import Preview with Diff
// ──────────────────────────────────────────────

func (s *Scanner) PreviewImport(ctx context.Context, repo *GitHubRepo, currentFunctions []DetectedFunction, branch string) (*ImportPreview, error) {
	newResult, err := s.ScanRepo(ctx, repo.Owner.Login, repo.Name, branch)
	if err != nil {
		return nil, err
	}

	newFnMap := make(map[string]DetectedFunction)
	for _, fn := range newResult.Functions {
		newFnMap[fn.EntryPoint] = fn
	}

	currentFnMap := make(map[string]DetectedFunction)
	for _, fn := range currentFunctions {
		currentFnMap[fn.EntryPoint] = fn
	}

	var toAdd []DetectedFunction
	var toUpdate []FunctionUpdate
	var toDelete []string

	for entryPt, newFn := range newFnMap {
		if oldFn, exists := currentFnMap[entryPt]; exists {
			changes := s.detectFunctionChanges(oldFn, newFn)
			if len(changes) > 0 {
				toUpdate = append(toUpdate, FunctionUpdate{
					Function:    newFn,
					OldRuntime:  oldFn.Runtime,
					OldEntryPt:  oldFn.EntryPoint,
					Changes:     changes,
				})
			}
		} else {
			toAdd = append(toAdd, newFn)
		}
	}

	for entryPt := range currentFnMap {
		if _, exists := newFnMap[entryPt]; !exists {
			toDelete = append(toDelete, entryPt)
		}
	}

	breaking := s.detectBreakingChanges(toUpdate)
	depChanges := s.detectDependencyChanges(currentFunctions, newResult.Functions)

	return &ImportPreview{
		FunctionsToAdd:    toAdd,
		FunctionsToUpdate: toUpdate,
		FunctionsToDelete: toDelete,
		BreakingChanges:   breaking,
		DependencyChanges: depChanges,
		EstimatedCost:     s.estimatePreviewCost(toAdd, toUpdate, toDelete),
		ScanMode:          "preview",
	}, nil
}

func (s *Scanner) detectFunctionChanges(old, new DetectedFunction) []string {
	var changes []string
	if old.Runtime != new.Runtime {
		changes = append(changes, fmt.Sprintf("runtime: %s -> %s", old.Runtime, new.Runtime))
	}
	if old.EntryPoint != new.EntryPoint {
		changes = append(changes, fmt.Sprintf("entry_point: %s -> %s", old.EntryPoint, new.EntryPoint))
	}
	if old.Confidence != new.Confidence {
		changes = append(changes, fmt.Sprintf("confidence: %.2f -> %.2f", old.Confidence, new.Confidence))
	}
	return changes
}

func (s *Scanner) detectBreakingChanges(updates []FunctionUpdate) []BreakingChange {
	var breaking []BreakingChange
	for _, u := range updates {
		for _, change := range u.Changes {
			if strings.Contains(change, "runtime:") {
				breaking = append(breaking, BreakingChange{
					Function:    u.Function.Name,
					Description: fmt.Sprintf("Runtime change: %s", change),
					Severity:    "high",
				})
			}
		}
	}
	return breaking
}

func (s *Scanner) detectDependencyChanges(current, updated []DetectedFunction) []DependencyDelta {
	var deltas []DependencyDelta
	currMap := make(map[string]*DependencyInfo)
	for i := range current {
		currMap[current[i].EntryPoint] = current[i].Dependencies
	}
	for _, fn := range updated {
		if oldDep := currMap[fn.EntryPoint]; oldDep != nil && fn.Dependencies != nil {
			if oldDep.Manager != fn.Dependencies.Manager {
				deltas = append(deltas, DependencyDelta{
					Package: fn.Name,
					OldVer:  oldDep.Manager,
					NewVer:  fn.Dependencies.Manager,
					Change:  "manager_changed",
				})
			}
		}
	}
	return deltas
}

func (s *Scanner) estimatePreviewCost(toAdd []DetectedFunction, toUpdate []FunctionUpdate, toDelete []string) float64 {
	base := 0.01
	addCost := float64(len(toAdd)) * 0.005
	updateCost := float64(len(toUpdate)) * 0.002
	delCost := float64(len(toDelete)) * 0.001
	return base + addCost + updateCost + delCost
}

// ──────────────────────────────────────────────
// Conflict Detection
// ──────────────────────────────────────────────

func (s *Scanner) DetectConflicts(localFunctions []DetectedFunction, remoteFunctions []DetectedFunction) []Conflict {
	var conflicts []Conflict
	localMap := make(map[string]DetectedFunction)
	for _, fn := range localFunctions {
		localMap[fn.EntryPoint] = fn
	}
	remoteMap := make(map[string]DetectedFunction)
	for _, fn := range remoteFunctions {
		remoteMap[fn.EntryPoint] = fn
	}

	for entryPt, local := range localMap {
		if remote, exists := remoteMap[entryPt]; exists {
			if local.Runtime != remote.Runtime {
				conflicts = append(conflicts, Conflict{
					LocalFunction:   local.Name,
					RemoteFunction: remote.Name,
					ConflictType:    "runtime_mismatch",
					Resolution:      "use_remote",
				})
			}
		} else {
			conflicts = append(conflicts, Conflict{
				LocalFunction:   local.Name,
				RemoteFunction:  "",
				ConflictType:    "deleted_remotely",
				Resolution:      "keep_local",
			})
		}
	}

	for entryPt, remote := range remoteMap {
		if _, exists := localMap[entryPt]; !exists {
			conflicts = append(conflicts, Conflict{
				LocalFunction:   "",
				RemoteFunction:  remote.Name,
				ConflictType:    "new_remote",
				Resolution:      "use_remote",
			})
		}
	}

	return conflicts
}

// ──────────────────────────────────────────────
// Enhanced Function Metadata
// ──────────────────────────────────────────────

func (s *Scanner) ExtractEnhancedFunction(ctx context.Context, repo *GitHubRepo, fn DetectedFunction) (*EnhancedFunction, error) {
	ef := &EnhancedFunction{
		DetectedFunction: fn,
		EventSources:     s.extractEventSources(fn),
		EnvironmentVars:  []EnvVar{},
		SecretsUsed:      []string{},
	}

	content, err := s.client.GetFileContent(ctx, repo.Owner.Login, repo.Name, fn.EntryPoint, "")
	if err != nil {
		return ef, nil
	}

	ef.EnvironmentVars = s.extractEnvVars(string(content))
	ef.SecretsUsed = s.extractSecrets(ef.EnvironmentVars)
	ef.MemoryEstimateMB = s.estimateMemory(fn)
	ef.ColdStartHintS = s.estimateColdStart(fn.Runtime)
	ef.DocumentationURL = s.findDocumentationURL(repo, fn)
	ef.TestFile = s.findTestFile(fn)

	return ef, nil
}

func (s *Scanner) extractEventSources(fn DetectedFunction) []string {
	var sources []string
	if fn.Manifest == nil {
		return sources
	}

	if trigger, ok := fn.Manifest["trigger_type"].(string); ok {
		sources = append(sources, trigger)
	}

	switch fn.Runtime {
	case "nextjs":
		sources = append(sources, "http")
	case "aws-lambda":
		sources = append(sources, "aws:alexa", "aws:api", "aws:s3", "aws:sqs")
	case "azure-functions":
		sources = append(sources, "http", "timer", "queue", "blob")
	case "gcf", "gcf-python":
		sources = append(sources, "http", "pubsub", "storage", "schedule")
	}

	return sources
}

func (s *Scanner) extractEnvVars(content string) []EnvVar {
	var envVars []EnvVar
	patterns := []string{
		`process\.env\.(\w+)`,
		`os\.environ\["(\w+)"\]`,
		`os\.getenv\["(\w+)"\]`,
		`os\.environ\.get\("(\w+)"\)`,
		`\b(\w+)=.*\${\w+}`,
	}

	for _, pattern := range patterns {
		if matches := reFindAllString(content, pattern); matches != nil {
			for _, m := range matches {
				envVars = append(envVars, EnvVar{
					Name:       m,
					IsSecret:   s.looksLikeSecret(m),
					Referenced: true,
				})
			}
		}
	}

	return envVars
}

func (s *Scanner) looksLikeSecret(name string) bool {
	secretPatterns := []string{"KEY", "SECRET", "PASSWORD", "TOKEN", "CREDENTIAL", "PRIVATE"}
	nameUpper := strings.ToUpper(name)
	for _, p := range secretPatterns {
		if strings.Contains(nameUpper, p) {
			return true
		}
	}
	return false
}

func (s *Scanner) extractSecrets(envVars []EnvVar) []string {
	var secrets []string
	for _, e := range envVars {
		if e.IsSecret {
			secrets = append(secrets, e.Name)
		}
	}
	return secrets
}

func (s *Scanner) estimateMemory(fn DetectedFunction) int {
	switch {
	case strings.Contains(fn.Runtime, "node"):
		return 256
	case strings.Contains(fn.Runtime, "python"):
		return 512
	case strings.Contains(fn.Runtime, "go"):
		return 128
	case strings.Contains(fn.Runtime, "rust"):
		return 128
	case strings.Contains(fn.Runtime, "java"):
		return 512
	default:
		return 256
	}
}

func (s *Scanner) estimateColdStart(runtime string) float64 {
	switch {
	case strings.Contains(runtime, "java"):
		return 3.5
	case strings.Contains(runtime, "python"):
		return 1.0
	case strings.Contains(runtime, "node"):
		return 0.5
	case strings.Contains(runtime, "go"):
		return 0.2
	case strings.Contains(runtime, "rust"):
		return 0.3
	default:
		return 1.0
	}
}

func (s *Scanner) findDocumentationURL(repo *GitHubRepo, fn DetectedFunction) string {
	dir := filepath.Dir(fn.EntryPoint)
	base := filepath.Base(fn.EntryPoint)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	docFiles := []string{"README.md", "readme.md", "docs.md", "API.md", nameWithoutExt + ".md"}
	for _, doc := range docFiles {
		docPath := dir + "/" + doc
		if dir == "." || dir == "" {
			docPath = doc
		}
		return fmt.Sprintf("https://github.com/%s/%s/blob/main/%s", repo.Owner.Login, repo.Name, docPath)
	}

	return ""
}

func (s *Scanner) findTestFile(fn DetectedFunction) string {
	dir := filepath.Dir(fn.EntryPoint)
	base := filepath.Base(fn.EntryPoint)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	testPatterns := []string{
		dir + "/test/" + base,
		dir + "/test/" + nameWithoutExt + ".test" + ext,
		dir + "/__tests__/" + base,
		dir + "/__tests__/" + nameWithoutExt + ".test" + ext,
		dir + "/spec/" + base,
		dir + "/" + nameWithoutExt + ".test" + ext,
		dir + "/tests/" + base,
	}

	for _, p := range testPatterns {
		return p
	}

	return ""
}

func reFindAllString(content, pattern string) []string {
	var matches []string
	re := regexp.MustCompile(pattern)
	found := re.FindAllStringSubmatch(content, -1)
	for _, f := range found {
		if len(f) > 1 {
			matches = append(matches, f[1])
		}
	}
	return matches
}
