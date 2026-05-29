package agent

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const (
	maxMarketplaceRefLen = 128
)

var marketplaceRefPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]{0,127}$`)

func validateMarketplaceFunctionRef(author, name string) error {
	author = strings.TrimSpace(author)
	name = strings.TrimSpace(name)
	if author == "" || name == "" {
		return fmt.Errorf("function_author and function_name are required")
	}
	if len(author) > maxMarketplaceRefLen || len(name) > maxMarketplaceRefLen {
		return fmt.Errorf("function_author and function_name must be at most %d characters", maxMarketplaceRefLen)
	}
	if !marketplaceRefPattern.MatchString(author) || !marketplaceRefPattern.MatchString(name) {
		return fmt.Errorf("function_author and function_name contain invalid characters")
	}
	return nil
}

func clientIPFromRequest(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return strings.TrimSpace(r.RemoteAddr)
}
