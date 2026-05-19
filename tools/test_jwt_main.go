package testjwt

import (
	"fmt"
	"log"

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
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZTNkYjA0MjAtZDFhZi00OTlkLTg4OTgtYWRkZDQ0Y2EwNjY4IiwiZW1haWwiOiJhZG1pbkBmdW5jdGlvbmZseS5kZXYiLCJ0ZW5hbnRfaWQiOiJmNGJkODQwMC1jMzM5LTQwOGQtOWI3My01NzBiOWQ3YjkxOWEiLCJyb2xlIjoic3VwZXJfYWRtaW4iLCJwZXJtaXNzaW9ucyI6WyJ0ZW5hbnRzLnJlYWQiLCJ0ZW5hbnRzLndyaXRlIiwidXNlcnMucmVhZCIsInVzZXJzLndyaXRlIiwiYmlsbGluZy5yZWFkIiwiYmlsbGluZy53cml0ZSIsImRlcGxveW1lbnRzLnJlYWQiLCJkZXBsb3ltZW50cy53cml0ZSIsImF1ZGl0LnJlYWQiLCJzeXN0ZW0ucmVhZCIsInN5c3RlbS53cml0ZSIsInN0YXRlLnJlYWQiLCJzdGF0ZS53cml0ZSIsInN0YXRlLmRlbGV0ZSIsInN0YXRlLmFkbWluIiwidHJpZ2dlcnMubWFuYWdlIiwic25hcHNob3RzLmNyZWF0ZSIsInNuYXBzaG90cy5yZXN0b3JlIiwicmVwbGF5LmFjY2VzcyIsIm1lbW9yeS5yZWFkIiwibWVtb3J5LndyaXRlIiwicmVnaXN0cnkucHVibGlzaCIsInJlZ2lzdHJ5LnZlcmlmeSIsInJlZ2lzdHJ5LmFwcHJvdmUiLCJyZWdpc3RyeS5zaWduIiwicmVnaXN0cnkubWFuYWdlIiwibW9uaXRvcmluZy5hbGVydHMiLCJtb25pdG9yaW5nLm1hbmFnZSIsIm1vbml0b3JpbmcubWV0cmljcyIsIm1vbml0b3JpbmcubWRtaW4iLCJtb25pdG9yaW5nLmhlYWx0aCIsInNlY3VyaXR5LmluY2lkZW50cyIsInNlY3VyaXR5LnNjYW5zIiwic2VjdXJpdHkuYXVkaXQiLCJzZWN1cml0eS5hZG1pbiIsImNvbnRlbnQuY3JlYXRlIiwiY29udGVudC5lZGl0IiwiY29udGVudC5wdWJsaXNoIiwiY29udGVudC5tYW5hZ2UiLCJjaGFuZ2Vsb2cubWFuYWdlIiwiYmxvZy5tYW5hZ2UiLCJ0ZWFtLm1lbWJlcnMubWFuYWdlIiwidGVhbS5yb2xlcy5hc3NpZ24iLCJ0ZWFtLnJlc291cmNlcy5zaGFyZSIsInZlcmlmaWNhdGlvbi5hcHByb3ZlIiwidmVyaWZpY2F0aW9uLnNpZ24iLCJ2ZXJpZmljYXRpb24ub3ZlcnJpZGUiLCJmZWVkYmFjay5tb2RlcmF0ZSIsImZlZWRiYWNrLmFuYWx5dGljcyJdLCJpc3MiOiJmdW5jdGlvbmZseSIsInN1YiI6ImUzZGIwNDIwLWQxYWYtNDk5ZC04ODk4LWFkZGQ0NGNhMDY2OCIsImV4cCI6MTc3MTc5MzA1OSwiaWF0IjoxNzcxNzA2NjU5fQ.Up4vD4kCx8bWYWxkEgc3PHh7tEMkE-r7Oqjtw5MzaUo"

	secrets := []string{
		"your-secret-key-here",
		"functionfly-secret",
		"super-secret-jwt-key",
		"jwt-secret-key",
	}

	for _, secret := range secrets {
		fmt.Printf("Testing secret: %s\n", secret)

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			fmt.Printf("  ✅ Valid token!\n")
			fmt.Printf("  User: %s (%s)\n", claims.Email, claims.Role)
			fmt.Printf("  Tenant: %s\n", claims.TenantID)
			return
		} else {
			fmt.Printf("  Invalid token\n")
		}
		fmt.Println()
	}

	log.Fatal("No valid secret found")
}