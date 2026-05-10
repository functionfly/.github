package github

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMonorepoDetector(t *testing.T) {
	logger := logrus.New()
	client := &Client{}

	tests := []struct {
		name        string
		entries     []GitHubTreeEntry
		wantNil     bool
		wantPackage bool
	}{
		{
			name: "pnpm workspace",
			entries: []GitHubTreeEntry{
				{Path: "pnpm-workspace.yaml", Type: "blob"},
				{Path: "packages/pkg1/package.json", Type: "blob"},
			},
			wantPackage: true,
		},
		{
			name: "lerna monorepo",
			entries: []GitHubTreeEntry{
				{Path: "lerna.json", Type: "blob"},
				{Path: "packages/pkg1/package.json", Type: "blob"},
			},
			wantPackage: true,
		},
		{
			name: "turbo monorepo",
			entries: []GitHubTreeEntry{
				{Path: "turbo.json", Type: "blob"},
				{Path: "apps/web/package.json", Type: "blob"},
			},
			wantPackage: true,
		},
		{
			name: "nx monorepo",
			entries: []GitHubTreeEntry{
				{Path: "nx.json", Type: "blob"},
				{Path: "packages/shared/package.json", Type: "blob"},
			},
			wantPackage: true,
		},
		{
			name: "manual monorepo with packages dir",
			entries: []GitHubTreeEntry{
				{Path: "package.json", Type: "blob"},
				{Path: "packages/pkg1/package.json", Type: "blob"},
			},
			wantPackage: true,
		},
		{
			name: "not a monorepo",
			entries: []GitHubTreeEntry{
				{Path: "package.json", Type: "blob"},
				{Path: "index.js", Type: "blob"},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &MonorepoDetector{client: client, logger: logger}
			result, err := d.Detect(context.Background(), &GitHubRepo{}, tt.entries)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result.Functions) == 0 {
				t.Error("expected at least one function")
			}
		})
	}
}

func TestSignatureDetector_AnalyzeContent(t *testing.T) {
	d := &SignatureDetector{}

	tests := []struct {
		name     string
		path     string
		content  string
		wantName string
		wantRt   string
	}{
		{
			name:     "node async handler",
			path:     "src/handler.ts",
			content:  "export async function handler(event) { return event }",
			wantName: "handler",
			wantRt:   "nodejs",
		},
		{
			name:     "node exports handler",
			path:     "index.js",
			content:  "exports.handler = async (event) => { return event }",
			wantName: "handler",
			wantRt:   "nodejs",
		},
		{
			name:     "module exports handler",
			path:     "handler.js",
			content:  "module.exports.handler = async (event) => { return event }",
			wantName: "handler",
			wantRt:   "nodejs",
		},
		{
			name:     "python lambda handler",
			path:     "lambda_function.py",
			content:  "def lambda_handler(event, context):\n    return event",
			wantName: "lambda_handler",
			wantRt:   "python",
		},
		{
			name:     "python async handler",
			path:     "handler.py",
			content:  "async def lambda_handler(event, context):\n    return event",
			wantName: "lambda_handler",
			wantRt:   "python",
		},
		{
			name:     "go handler",
			path:     "handler.go",
			content:  "func Handle(w http.ResponseWriter, r *http.Request) {\n    json.NewEncoder(w).Encode(map[string]string{})\n}",
			wantName: "Handle",
			wantRt:   "go",
		},
		{
			name:     "go HandleRequest",
			path:     "handler.go",
			content:  "func HandleRequest(w http.ResponseWriter, r *http.Request) { }",
			wantName: "HandleRequest",
			wantRt:   "go",
		},
		{
			name:     "rust handler with tokio",
			path:     "main.rs",
			content:  "#[tokio::main]\nasync fn handle(req: Request) -> Response { todo!() }",
			wantName: "handle",
			wantRt:   "rust",
		},
		{
			name:     "java handler",
			path:     "Function.java",
			content:  "public class Function implements RequestHandler {\n    public Object handleRequest(Object input, Context context) { return input; }",
			wantName: "handleRequest",
			wantRt:   "java",
		},
		{
			name:     "no handler signature",
			path:     "utils.js",
			content:  "const helper = () => { return true }",
			wantName: "",
			wantRt:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, rt := d.AnalyzeContent(tt.path, []byte(tt.content))
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if rt != tt.wantRt {
				t.Errorf("runtime = %q, want %q", rt, tt.wantRt)
			}
		})
	}
}

func TestSignatureDetector_IsCodeFile(t *testing.T) {
	d := &SignatureDetector{}

	tests := []struct {
		filename string
		want     bool
	}{
		{"handler.ts", true},
		{"handler.js", true},
		{"handler.jsx", true},
		{"handler.mjs", true},
		{"handler.py", true},
		{"handler.go", true},
		{"handler.rs", true},
		{"handler.java", true},
		{"handler.rb", true},
		{"handler.php", true},
		{"handler.cs", true},
		{"README.md", false},
		{"Dockerfile", false},
		{"config.yml", false},
		{"Makefile", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := d.isCodeFile(tt.filename)
			if got != tt.want {
				t.Errorf("isCodeFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestHasFile(t *testing.T) {
	entries := []GitHubTreeEntry{
		{Path: "package.json", Type: "blob"},
		{Path: "src/index.ts", Type: "blob"},
		{Path: "README.md", Type: "blob"},
	}

	tests := []struct {
		filename string
		want     bool
	}{
		{"package.json", true},
		{"README.md", true},
		{"pnpm-workspace.yaml", false},
		{"src/index.ts", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := hasFile(entries, tt.filename)
			if got != tt.want {
				t.Errorf("hasFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLangStatsMap_Dominant(t *testing.T) {
	tests := []struct {
		name   string
		stats  langStatsMap
		want   string
	}{
		{"typescript dominant", langStatsMap{"TypeScript": 60.0, "Python": 30.0}, "TypeScript"},
		{"python dominant", langStatsMap{"Python": 80.0, "Go": 20.0}, "Python"},
		{"empty", langStatsMap{}, ""},
		{"single", langStatsMap{"Go": 100.0}, "Go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stats.Dominant(); got != tt.want {
				t.Errorf("Dominant() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLangStatsMap_DefaultRuntime(t *testing.T) {
	tests := []struct {
		name  string
		stats langStatsMap
		want  string
	}{
		{"typescript", langStatsMap{"TypeScript": 100.0}, "node18"},
		{"javascript", langStatsMap{"JavaScript": 100.0}, "node18"},
		{"python", langStatsMap{"Python": 100.0}, "python3.11"},
		{"go", langStatsMap{"Go": 100.0}, "go1.22"},
		{"rust", langStatsMap{"Rust": 100.0}, "rust1.75"},
		{"java", langStatsMap{"Java": 100.0}, "java17"},
		{"ruby", langStatsMap{"Ruby": 100.0}, "ruby3.2"},
		{"php", langStatsMap{"PHP": 100.0}, "php8.2"},
		{"csharp", langStatsMap{"C#": 100.0}, "dotnet6"},
		{"zig", langStatsMap{"Zig": 100.0}, "Zig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stats.DefaultRuntime(); got != tt.want {
				t.Errorf("DefaultRuntime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "b") {
		t.Error("expected contains(slice, 'b') = true")
	}
	if contains(slice, "d") {
		t.Error("expected contains(slice, 'd') = false")
	}
	if contains([]string{}, "a") {
		t.Error("expected contains(empty, 'a') = false")
	}
}

func TestDetectConflicts(t *testing.T) {
	logger := logrus.New()
	s := &Scanner{logger: logger}

	local := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "node18"},
		{Name: "fn2", EntryPoint: "app.py", Runtime: "python3.11"},
	}

	remote := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "node20"},
		{Name: "fn3", EntryPoint: "main.go", Runtime: "go1.22"},
	}

	conflicts := s.DetectConflicts(local, remote)

	if len(conflicts) != 3 {
		t.Fatalf("expected 3 conflicts, got %d", len(conflicts))
	}

	found := make(map[string]bool)
	for _, c := range conflicts {
		found[c.ConflictType] = true
	}

	if !found["runtime_mismatch"] {
		t.Error("expected runtime_mismatch conflict")
	}
	if !found["deleted_remotely"] {
		t.Error("expected deleted_remotely conflict")
	}
	if !found["new_remote"] {
		t.Error("expected new_remote conflict")
	}
}

func TestDetectConflicts_NoConflicts(t *testing.T) {
	logger := logrus.New()
	s := &Scanner{logger: logger}

	local := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "node18"},
	}

	remote := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "node18"},
	}

	conflicts := s.DetectConflicts(local, remote)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestDetectConflicts_RuntimeMismatch(t *testing.T) {
	logger := logrus.New()
	s := &Scanner{logger: logger}

	local := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "node18"},
	}
	remote := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "python3.11"},
	}

	conflicts := s.DetectConflicts(local, remote)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].ConflictType != "runtime_mismatch" {
		t.Errorf("conflict type = %q, want %q", conflicts[0].ConflictType, "runtime_mismatch")
	}
	if conflicts[0].Resolution != "use_remote" {
		t.Errorf("resolution = %q, want %q", conflicts[0].Resolution, "use_remote")
	}
}

func TestEstimateCost(t *testing.T) {
	s := &Scanner{}

	result := &ScanResult{Functions: make([]DetectedFunction, 5)}
	cost := s.estimateCost(result)

	expectedMin := 0.01 + 5*0.005
	if cost < expectedMin {
		t.Errorf("estimateCost() = %v, want >= %v", cost, expectedMin)
	}
}

func TestEstimateImportTime(t *testing.T) {
	s := &Scanner{}

	result := &ScanResult{Functions: make([]DetectedFunction, 3)}
	timeS := s.estimateImportTime(result)

	expected := 10 + 3*5
	if timeS != expected {
		t.Errorf("estimateImportTime() = %d, want %d", timeS, expected)
	}
}

func TestResolveRuntimeFromStats(t *testing.T) {
	logger := logrus.New()
	s := &Scanner{logger: logger}

	detected := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "unknown"},
		{Name: "fn2", EntryPoint: "app.py", Runtime: "python3.11"},
		{Name: "fn3", EntryPoint: "main.go", Runtime: ""},
	}

	langStats := map[string]float64{"TypeScript": 70.0, "Go": 30.0}
	s.resolveRuntimeFromStats(&GitHubRepo{}, detected, langStats)

	if detected[0].Runtime != "node18" {
		t.Errorf("detected[0].Runtime = %q, want %q", detected[0].Runtime, "node18")
	}
	if detected[2].Runtime != "node18" {
		t.Errorf("detected[2].Runtime = %q, want %q", detected[2].Runtime, "node18")
	}
	if detected[1].Runtime != "python3.11" {
		t.Errorf("detected[1].Runtime = %q, want %q", detected[1].Runtime, "python3.11")
	}
}

func TestResolveRuntimeFromStats_EmptyDetected(t *testing.T) {
	logger := logrus.New()
	s := &Scanner{logger: logger}

	detected := []DetectedFunction{}
	langStats := map[string]float64{"Python": 100.0}

	s.resolveRuntimeFromStats(&GitHubRepo{}, detected, langStats)
}

func TestResolveRuntimeFromStats_EmptyStats(t *testing.T) {
	logger := logrus.New()
	s := &Scanner{logger: logger}

	detected := []DetectedFunction{
		{Name: "fn1", EntryPoint: "handler.js", Runtime: "unknown"},
	}
	langStats := map[string]float64{}

	s.resolveRuntimeFromStats(&GitHubRepo{}, detected, langStats)

	if detected[0].Runtime != "unknown" {
		t.Errorf("detected[0].Runtime = %q, want %q", detected[0].Runtime, "unknown")
	}
}