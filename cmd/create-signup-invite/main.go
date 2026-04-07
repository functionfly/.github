package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

const opTimeout = 20 * time.Second

func main() {
	var (
		label     = flag.String("label", "", "Label for the invite (required)")
		maxUses   = flag.Int("max-uses", 0, "Max uses (0 = unlimited)")
		expiresIn = flag.String("expires-in", "", "Duration until expiry (e.g. 720h for 30 days)")
	)
	flag.Parse()

	if *label == "" {
		fmt.Println("Usage: create-signup-invite -label <label> [-max-uses N] [-expires-in 720h]")
		fmt.Println("Example: create-signup-invite -label 'thefunctionfly@gmail.com'")
		fmt.Println("Requires DATABASE_URL env var.")
		os.Exit(1)
	}

	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	// Swap pooler endpoint to direct so pgx simple-protocol queries work.
	if raw := strings.TrimSpace(os.Getenv("DATABASE_URL")); strings.Contains(raw, "-pooler.") {
		_ = os.Setenv("DATABASE_URL", strings.Replace(raw, "-pooler.", ".", 1))
	}

	pg, err := storage.NewPostgresDBWithOptions(true)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()

	var expiresAt *time.Time
	if *expiresIn != "" {
		dur, err := time.ParseDuration(*expiresIn)
		if err != nil {
			log.Fatalf("Invalid expires-in duration: %v", err)
		}
		t := time.Now().UTC().Add(dur)
		expiresAt = &t
	}

	var maxUsesPtr *int
	if *maxUses > 0 {
		maxUsesPtr = maxUses
	}

	id, plainCode, err := pg.CreateSignupInvite(ctx, *label, maxUsesPtr, expiresAt, nil)
	if err != nil {
		log.Fatalf("Failed to create invite: %v", err)
	}

	fmt.Println()
	fmt.Println("Signup invite created successfully!")
	fmt.Println()
	fmt.Printf("  Code:      %s\n", plainCode)
	fmt.Printf("  Label:     %s\n", *label)
	fmt.Printf("  ID:        %s\n", id)
	if maxUsesPtr != nil {
		fmt.Printf("  Max Uses:  %d\n", *maxUsesPtr)
	} else {
		fmt.Printf("  Max Uses:  unlimited\n")
	}
	if expiresAt != nil {
		fmt.Printf("  Expires:   %s\n", expiresAt.Format(time.RFC3339))
	} else {
		fmt.Printf("  Expires:   never\n")
	}
	fmt.Println()
	fmt.Println("Share this code with the invitee. It will not be shown again.")
}
