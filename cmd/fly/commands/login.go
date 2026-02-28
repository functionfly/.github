package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

func NewLoginCmd() *cobra.Command {
	var provider string
	var noBrowser bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with FunctionFly",
		Long:  "Authenticate with FunctionFly using OAuth.\n\nOpens your browser to complete authentication.",
		Example: "  fly login\n  fly login --provider github\n  fly login --provider google\n  fly login --no-browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(provider, noBrowser)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "github", "OAuth provider (github, google)")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the auth URL instead of opening a browser")
	return cmd
}

func runLogin(provider string, noBrowser bool) error {
	cfg, _ := LoadConfig()
	baseURL := "https://api.functionfly.com"
	if cfg != nil && cfg.API.URL != "" {
		baseURL = cfg.API.URL
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("could not start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	authURL := fmt.Sprintf("%s/v1/auth/oauth/%s?redirect_uri=%s", baseURL, provider, callbackURL)
	fmt.Printf("🔐 Authenticating with %s...\n", provider)
	if noBrowser {
		fmt.Printf("\nOpen this URL in your browser:\n%s\n\n", authURL)
	} else {
		fmt.Printf("Opening browser...\n")
		if err := openBrowser(authURL); err != nil {
			fmt.Printf("Could not open browser automatically.\nOpen this URL manually:\n%s\n\n", authURL)
		}
	}
	fmt.Printf("Waiting for authentication (Ctrl+C to cancel)...\n")
	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "no token received"
			}
			http.Error(w, "Authentication failed: "+errMsg, http.StatusBadRequest)
			errCh <- fmt.Errorf("authentication failed: %s", errMsg)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h2>✅ Authentication successful!</h2><p>You can close this tab.</p></body></html>`)
		tokenCh <- token
	})
	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	var token string
	select {
	case token = <-tokenCh:
	case err := <-errCh:
		server.Close()
		return err
	case <-ctx.Done():
		server.Close()
		return fmt.Errorf("authentication timed out after 5 minutes")
	}
	server.Close()
	client := NewAPIClientWithToken(token)
	var userResp struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Provider  string `json:"provider"`
		AvatarURL string `json:"avatar_url"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := client.Get("/v1/auth/me", &userResp); err != nil {
		fmt.Printf("⚠️  Could not fetch user info: %v\n", err)
	}
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	if userResp.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, userResp.ExpiresAt); err == nil {
			expiresAt = t
		}
	}
	creds := &Credentials{
		Version:   "1.0.0",
		User:      UserInfo{ID: userResp.ID, Username: userResp.Username, Email: userResp.Email, Provider: provider, AvatarURL: userResp.AvatarURL},
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := SaveCredentials(creds); err != nil {
		return fmt.Errorf("could not save credentials: %w", err)
	}
	username := userResp.Username
	if username == "" {
		username = "unknown"
	}
	fmt.Printf("\n✅ Logged in as %s\n", username)
	if userResp.Email != "" {
		fmt.Printf("   Email: %s\n", userResp.Email)
	}
	fmt.Printf("   Provider: %s\n", provider)
	fmt.Printf("\nYour namespace: fx://%s/*\n", username)
	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}
