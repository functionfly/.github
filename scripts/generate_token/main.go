package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	TenantID    string   `json:"tenant_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func main() {
	// Use JWT_SECRET so server and token stay in sync (e.g. from .env)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "ERROR: JWT_SECRET environment variable is required")
		os.Exit(1)
	}

	permissions := []string{
		"tenants.read", "tenants.write", "users.read", "users.write",
		"billing.read", "billing.write", "deployments.read", "deployments.write",
		"audit.read", "system.read", "system.write", "state.read", "state.write",
		"state.delete", "state.admin", "triggers.manage", "snapshots.create",
		"snapshots.restore", "replay.access", "memory.read", "memory.write",
		"registry.publish", "registry.verify", "registry.approve", "registry.sign",
		"registry.manage", "monitoring.alerts", "monitoring.manage", "monitoring.metrics",
		"monitoring.admin", "monitoring.health", "security.incidents", "security.scans",
		"security.audit", "security.admin", "content.create", "content.edit",
		"content.publish", "content.manage", "changelog.manage", "blog.manage",
		"team.members.manage", "team.roles.assign", "team.resources.share",
		"verification.approve", "verification.sign", "verification.override",
		"feedback.moderate", "feedback.analytics",
	}

	claims := Claims{
		UserID:      "e3db0420-d1af-499d-8898-addd44ca0668",
		Email:       "admin@functionfly.dev",
		TenantID:    "f4bd8400-c339-408d-9b73-570b9d7b919a",
		Role:        "super_admin",
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "functionfly",
			Subject:   "e3db0420-d1af-499d-8898-addd44ca0668",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	fmt.Println(tokenString)
}
