/*
Copyright © 2026 FunctionFly

*/
package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/functionfly/functionfly/internal/credentials"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub or Google OAuth",
	Long: `Creates developer identity in 5 seconds.

Flow:
1. CLI opens browser → OAuth provider (GitHub/Google)
2. User authorizes application
3. Callback received with auth code
4. Exchange code for JWT token
5. Store token in ~/.functionfly/credentials.json

Namespace: After login, developer gets global namespace:
- fx://username/* (e.g., fx://trase/slugify)`,
	Run: loginRun,
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Local flags
	loginCmd.Flags().StringP("provider", "p", "github", "OAuth provider (github or google)")
	loginCmd.Flags().BoolP("browser", "b", true, "Open browser automatically")
}

// loginRun implements the login command
func loginRun(cmd *cobra.Command, args []string) {
	provider, _ := cmd.Flags().GetString("provider")
	_, _ = cmd.Flags().GetBool("browser") // browser flag for future OAuth implementation

	// Validate provider
	if provider != "github" && provider != "google" {
		log.Fatalf("Invalid provider '%s'. Supported providers: github, google", provider)
	}

	fmt.Printf("Logging in with %s...\n", provider)

	// For development/demo purposes, use mock OAuth flow
	// In production, this would call the API to get OAuth URL
	fmt.Println("Note: Using mock OAuth flow for development")

	// Simulate OAuth flow - in real implementation, this would:
	// 1. Call API to get OAuth URL: GET /v1/auth/oauth/{provider}
	// 2. Open browser to that URL
	// 3. Handle callback locally and call API callback endpoint

	// Generate mock credentials for development
	username := "developer"
	if provider == "github" {
		username = "githubuser"
	}

	creds := &credentials.Credentials{
		Version:   "1.0.0",
		User: credentials.User{
			ID:       "usr_mock123",
			Username: username,
			Email:    fmt.Sprintf("%s@example.com", username),
			Provider: provider,
		},
		Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.mock_token_for_development",
		TokenType: "Bearer",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Save credentials
	if err := credentials.Save(creds); err != nil {
		log.Fatalf("Failed to save credentials: %v", err)
	}

	fmt.Printf("✓ Successfully logged in as %s (%s)\n", creds.User.Username, creds.User.Email)
	fmt.Printf("Namespace: fx://%s/*\n", creds.User.Username)

	// Note about real implementation
	fmt.Println("\nNote: This is a mock implementation for development.")
	fmt.Println("Real OAuth flow would integrate with GitHub/Google OAuth providers.")
}
