package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var template string
	var force bool
	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Scaffold a new function project",
		Long:  "Create a new FunctionFly function project with all required files.",
		Example: "  fly init slugify\n  fly init --template typescript my-function\n  fly init --template python data-processor",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runInit(name, template, force)
		},
	}
	cmd.Flags().StringVarP(&template, "template", "t", "javascript", "Template (javascript, typescript, python)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	return cmd
}

func runInit(name, template string, force bool) error {
	if name == "" {
		if IsInteractive() {
			name = Prompt("Function name", "my-function")
		} else {
			return fmt.Errorf("function name is required\n   → Usage: fly init <name>")
		}
	}
	if !isValidFunctionName(name) {
		return fmt.Errorf("invalid function name: %q\n   → Names must be lowercase letters, numbers, and hyphens only", name)
	}
	if template == "javascript" && IsInteractive() {
		template = PromptSelect("Choose a template:", []string{"javascript", "typescript", "python"}, "javascript")
	}
	projectDir := filepath.Join(".", name)
	if _, err := os.Stat(projectDir); err == nil && !force {
		return fmt.Errorf("directory %q already exists\n   → Use --force to overwrite", projectDir)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	files, err := generateTemplateFiles(name, template)
	if err != nil {
		return err
	}
	for filename, content := range files {
		path := filepath.Join(projectDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", filename, err)
		}
		fmt.Printf("  ✓ %s\n", filepath.Join(name, filename))
	}
	fmt.Printf("\n✅ Created %s/\n\n", name)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  fly dev          # Run locally")
	fmt.Println("  fly publish      # Publish to the registry")
	return nil
}

func isValidFunctionName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func generateTemplateFiles(name, template string) (map[string]string, error) {
	files := map[string]string{}
	files["functionfly.jsonc"] = fmt.Sprintf(`{
  "$schema": "https://functionfly.com/schemas/functionfly.json",
  "name": "%s",
  "version": "1.0.0",
  "runtime": "%s",
  "public": true,
  "deterministic": true,
  "cache_ttl": 86400,
  "timeout_ms": 5000,
  "memory_mb": 128,
  "description": "A FunctionFly function"
}
`, name, runtimeForTemplate(template))
	switch template {
	case "typescript":
		files["index.ts"] = fmt.Sprintf("/**\n * %s - A FunctionFly function\n */\nexport default async function handler(input: string): Promise<string> {\n  return input;\n}\n", name)
	case "python":
		pyName := strings.ReplaceAll(name, "-", "_")
		_ = pyName
		files["main.py"] = fmt.Sprintf("\"\"\"\n%s - A FunctionFly function\n\"\"\"\n\nasync def handler(input: str) -> str:\n    return input\n", name)
	default:
		files["index.js"] = fmt.Sprintf("/**\n * %s - A FunctionFly function\n */\nexport default async function handler(input) {\n  return input;\n}\n", name)
	}
	files["test.http"] = fmt.Sprintf("### Test %s locally\nPOST http://localhost:8787\nContent-Type: application/json\n\n\"hello world\"\n", name)
	return files, nil
}

func runtimeForTemplate(template string) string {
	switch template {
	case "typescript":
		return "node20"
	case "python":
		return "python3.11"
	default:
		return "node20"
	}
}
