package middleware

import (
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ReservedUsernameChecker checks if a username is reserved
type ReservedUsernameChecker interface {
	IsReservedUsername(username string) bool
}

// reservedUsernames contains the list of reserved usernames that cannot be used
var reservedUsernames = map[string]bool{
	// Platform names
	"functionfly": true,
	"function":   true,
	"flypy":      true,
	"registry":   true,
	"api":        true,

	// System accounts
	"system":   true,
	"admin":    true,
	"support":  true,
	"root":     true,
	"nobody":   true,
	"moderator": true,

	// Dashboard routes
	"account":     true,
	"dashboard":   true,
	"billing":    true,
	"settings":   true,
	"login":      true,
	"logout":     true,
	"signup":     true,
	"register":   true,
	"auth":       true,
	"password":   true,
	"reset":      true,

	// Feature routes
	"run":        true,
	"play":       true,
	"docs":       true,
	"blog":       true,
	"market":     true,
	"marketplace": true,
	"enterprise": true,
	"security":   true,
	"trust":      true,
	"core":       true,
	"debug":      true,
	"status":     true,
	"health":     true,
	"metrics":    true,
	"monitoring": true,

	// API versions
	"v1":      true,
	"v2":      true,
	"v3":      true,
	"latest":  true,

	// Common reserved
	"www":      true,
	"mail":     true,
	"ftp":      true,
	"localhost": true,
	"static":   true,
	"assets":   true,
	"cdn":      true,
	"files":    true,
	"download": true,
	"upload":   true,

	// OAuth providers (reserved for future OAuth)
	"google":   true,
	"github":   true,
	"twitter":  true,
	"facebook": true,
	"microsoft": true,
	"apple":    true,
	"slack":    true,

	// Function execution paths
	"execute":   true,
	"run":       true,
	"playground": true,
	"sandbox":   true,
	"test":      true,

	// Special keywords
	"me":      true,
	"you":     true,
	"users":   true,
	"profile": true,
	"public":  true,
	"private": true,
	"search":  true,
	"find":    true,
	"help":    true,
	"about":   true,
	"contact": true,
	"terms":   true,
	"privacy": true,
}

// DefaultReservedUsernameChecker implements ReservedUsernameChecker using the default list
type DefaultReservedUsernameChecker struct{}

// NewDefaultReservedUsernameChecker creates a new default checker
func NewDefaultReservedUsernameChecker() *DefaultReservedUsernameChecker {
	return &DefaultReservedUsernameChecker{}
}

// IsReservedUsername checks if the username is in the reserved list
func (c *DefaultReservedUsernameChecker) IsReservedUsername(username string) bool {
	// Normalize to lowercase
	usernameLower := strings.ToLower(username)
	return reservedUsernames[usernameLower]
}

// DatabaseReservedUsernameChecker checks against a database table
type DatabaseReservedUsernameChecker struct {
	repo storage.Repository
}

// NewDatabaseReservedUsernameChecker creates a checker that also queries the database
func NewDatabaseReservedUsernameChecker(repo storage.Repository) *DatabaseReservedUsernameChecker {
	return &DatabaseReservedUsernameChecker{repo: repo}
}

// IsReservedUsername checks both the default list and database
func (c *DatabaseReservedUsernameChecker) IsReservedUsername(username string) bool {
	// Check default list first
	if reservedUsernames[strings.ToLower(username)] {
		return true
	}

	// Check database if available
	if c.repo != nil {
		isReserved, err := c.repo.IsUsernameReserved(username)
		if err == nil && isReserved {
			return true
		}
	}

	return false
}

// ValidateUsernameMiddleware validates that a username is not reserved
func ValidateUsernameMiddleware(checker ReservedUsernameChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			
			// Check username parameter
			if username, ok := vars["username"]; ok {
				if checker.IsReservedUsername(username) {
					logrus.Warnf("Attempted to use reserved username: %s", username)
					http.Error(w, "This username is reserved and cannot be used", http.StatusForbidden)
					return
				}
			}

			// Check author parameter (for function URLs)
			if author, ok := vars["author"]; ok {
				if checker.IsReservedUsername(author) {
					logrus.Warnf("Attempted to use reserved author name: %s", author)
					http.Error(w, "This author name is reserved and cannot be used", http.StatusForbidden)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// GetReservedUsernames returns the full list of reserved usernames
func GetReservedUsernames() []string {
	usernames := make([]string, 0, len(reservedUsernames))
	for k := range reservedUsernames {
		usernames = append(usernames, k)
	}
	return usernames
}

// IsValidUsernameFormat validates the username format
func IsValidUsernameFormat(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Must start with letter or number
	first := username[0]
	if !((first >= 'a' && first <= 'z') || (first >= '0' && first <= '9')) {
		return false
	}

	// Can only contain lowercase letters, numbers, hyphens, and underscores
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}

	return true
}
