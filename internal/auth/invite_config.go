package auth

import (
	"os"
	"strings"
)

// SignupInviteRequired returns true when SIGNUP_REQUIRE_INVITE_CODE is set to true or 1.
func SignupInviteRequired() bool {
	v := strings.TrimSpace(os.Getenv("SIGNUP_REQUIRE_INVITE_CODE"))
	return strings.EqualFold(v, "true") || v == "1"
}
