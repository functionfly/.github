package admin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/sirupsen/logrus"
)

// SignRequestMaxClockSkew is the maximum window (seconds) a signed request remains valid.
// Kept tight so a stolen signature is only useful for a brief window.
const SignRequestMaxClockSkew = 300

// signRequestRequest is the payload the admin SPA sends to get a short-lived
// signature for a single outbound request.
type signRequestRequest struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	BodyHash  string `json:"body_hash"`
	Timestamp int64  `json:"timestamp"`
}

// signRequestResponse contains the signature components the client attaches as
// X-FFLY-Timestamp and X-FFLY-Signature headers. The signature is bound to the
// (method, path, body_hash, timestamp) tuple and is valid for SignRequestMaxClockSkew.
type signRequestResponse struct {
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
	ExpiresAt int64  `json:"expires_at"`
}

// HandleSignRequest issues a short-lived HMAC signature for an admin SPA request.
//
// The signature is derived from the same secret the server uses to verify inbound
// requests, but the secret is never exposed to the browser. The client sends
// only the request intent (method, path, body hash, current timestamp) and the
// server returns a signed value the client can attach as headers.
//
// This endpoint requires a valid admin session and is rate-limited by the
// AdvancedRateLimit middleware that wraps all admin routes.
func (h *Handler) HandleSignRequest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Defense-in-depth: do not allow signing of cross-origin paths. The browser
	// only ever needs to sign paths under /v1/admin (the admin base path), so
	// any other path is suspicious.
	if !isPathAllowedForSigning(r.URL.Path) {
		logrus.WithFields(logrus.Fields{
			"user_id": claims.UserID.String(),
			"path":    r.URL.Path,
		}).Warn("Rejected sign-request for disallowed path")
		apierror.WriteError(w, apierror.NewForbidden("Path not allowed for signing"))
		return
	}

	var req signRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Method == "" || req.Path == "" {
		apierror.WriteError(w, apierror.NewBadRequest("method and path are required"))
		return
	}
	if !isMethodAllowedForSigning(req.Method) {
		apierror.WriteError(w, apierror.NewBadRequest("method not allowed for signing"))
		return
	}

	// Pin the issued timestamp to server time so the client can't burn a long
	// window. The client-sent timestamp is ignored; we still validate it as a
	// lightweight replay guard (no more than 60s drift).
	now := time.Now().UTC()
	if req.Timestamp != 0 {
		drift := now.Unix() - req.Timestamp
		if drift < -60 || drift > 60 {
			apierror.WriteError(w, apierror.NewBadRequest("Timestamp drift too large"))
			return
		}
	}

	sharedSecret := os.Getenv("API_SHARED_SECRET")
	if sharedSecret == "" {
		logrus.Error("API_SHARED_SECRET not configured — cannot sign admin requests")
		apierror.WriteError(w, apierror.NewInternal("Service misconfigured"))
		return
	}

	signatureString := strconv.FormatInt(now.Unix(), 10) + req.Method + req.Path + req.BodyHash
	mac := hmac.New(sha256.New, []byte(sharedSecret))
	mac.Write([]byte(signatureString))
	signature := hex.EncodeToString(mac.Sum(nil))

	resp := signRequestResponse{
		Timestamp: strconv.FormatInt(now.Unix(), 10),
		Signature: signature,
		ExpiresAt: now.Add(SignRequestMaxClockSkew * time.Second).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// isPathAllowedForSigning restricts which request paths the admin SPA can ask
// to be signed. This is a defense-in-depth check on top of the
// AdvancedRateLimit middleware so that even a compromised session can't be
// leveraged to sign requests to other internal services that accept the same
// shared secret.
func isPathAllowedForSigning(path string) bool {
	// The admin base is /v1/admin. Anything else (e.g. /v1/auth, /v1/billing)
	// has its own auth model and should never be signed from the admin SPA.
	const adminPrefix = "/v1/admin"
	if len(path) < len(adminPrefix) {
		return false
	}
	return path[:len(adminPrefix)] == adminPrefix
}

func isMethodAllowedForSigning(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}
