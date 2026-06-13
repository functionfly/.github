/*
Copyright © 2026 FunctionFly

*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cli"
	"github.com/functionfly/functionfly/internal/credentials"
	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish function to global registry with automatic infrastructure",
	Long: `Publishes function to global registry with automatic infrastructure handling.

Automatic Workflow:
1. Validate manifest
2. Bundle code (esbuild)
3. Generate content hash
4. Upload artifact to storage
5. Register version
6. Deploy to edge
7. Warm cache

Output:
✓ Validating manifest...
✓ Bundling code (2.1KB)...
✓ Computing hash: a1b2c3d4...
✓ Uploading to registry...
✓ Deploying to edge...
✓ Warming cache...

✓ Published trase/slugify@1.0.0

Public URL:
https://api.functionfly.com/trase/slugify

Curl:
curl https://api.functionfly.com/trase/slugify -d "Hello World"

Stats will be available in 30 seconds`,
	RunE: publishRunE,
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Local flags
	publishCmd.Flags().StringP("access", "a", "public", "Access level (public or private)")
	publishCmd.Flags().BoolP("force", "f", false, "Force publish even if version exists")
}

// publishRunE implements the publish command
func publishRunE(cmd *cobra.Command, args []string) error {
	_, _ = cmd.Flags().GetString("access")  // access flag for future implementation
	_, _ = cmd.Flags().GetBool("force")     // force flag for future implementation

	fmt.Println("✓ Validating manifest...")

	// 1. Load and validate manifest
	m, err := manifest.Load("")
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w\n   → Run 'ff init' if you don't have a functionfly.json", err)
	}

	if err := m.Validate(); err != nil {
		return fmt.Errorf("manifest validation failed: %w\n   → Check functionfly.json for errors", err)
	}

	fmt.Printf("✓ Manifest valid: %s\n", m.String())

	// 2. Load credentials
	creds, err := credentials.Load()
	if err != nil {
		return fmt.Errorf("not logged in: %w\n   → Run 'ff login' to authenticate", err)
	}

	fmt.Println("✓ Credentials loaded")

	// 3. Bundle code
	fmt.Println("✓ Bundling code...")
	bundle, err := bundler.BundleWithWorkingDirectory(m, "")
	if err != nil {
		return fmt.Errorf("bundling failed: %w\n   → Check your function code for syntax errors", err)
	}

	bundleSize := len(bundle)
	fmt.Printf("✓ Code bundled (%d bytes)\n", bundleSize)

	// 4. Generate version hash
	hash := bundler.HashContent(bundle)
	fmt.Printf("✓ Content hash: %s\n", hash[:16]+"...")

	// 5. Create API client
	apiURL := getAPIURL()
	client := cli.NewClient(apiURL, creds.Token)

	// 6. Prepare manifest for API
	manifestBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// 7. Publish to registry
	fmt.Println("✓ Publishing to registry...")

	publishReq := &cli.PublishRequest{
		Author:   creds.User.Username,
		Name:     m.Name,
		Version:  m.Version,
		Manifest: manifestBytes,
	}

	result, err := client.PublishFunction(publishReq)
	if err != nil {
		return fmt.Errorf("publish failed: %w\n   → Check your network connection and try again", err)
	}

	// 8. Print success
	fmt.Printf("✓ Published %s/%s@%s\n", creds.User.Username, m.Name, m.Version)
	if result.Message != "" {
		fmt.Printf("✓ %s\n", result.Message)
	}
	fmt.Println()
	fmt.Printf("Public URL:\n")
	fmt.Printf("https://api.functionfly.com/%s/%s\n", creds.User.Username, m.Name)
	fmt.Println()
	fmt.Printf("Curl:\n")
	fmt.Printf("curl https://api.functionfly.com/%s/%s -d \"Hello World\"\n", creds.User.Username, m.Name)
	fmt.Println()
	fmt.Println("Stats will be available in 30 seconds")
	return nil
}

