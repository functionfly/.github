// Publish the state reports as blog posts by inserting into blog_posts
// directly. Uses the same Markdown files produced by generate_state_report
// (default location: web/site/src/content/reports/YYYY-MM.md).
//
// The body column uses the blog app's portable-text format: each block is
// a single line, lines are joined by `\` separators. We do a minimal
// conversion (every paragraph / heading / list item / table row becomes
// one block) and rely on the blog app's renderer to handle line breaks
// within paragraphs as ` ` (space).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const categoryReports = "b6508c05-6d18-4949-9d8e-da39e81c4d9b" // "Product Deep Dives"

func main() {
	dir := flag.String("dir", "web/site/src/content/reports", "directory with .md reports")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "postgres DSN")
	authorID := flag.String("author", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "blog author ID")
	flag.Parse()
	if *dsn == "" {
		*dsn = "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=require"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	files, err := filepath.Glob(filepath.Join(*dir, "*.md"))
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	sort.Strings(files)
	posted := 0
	for _, f := range files {
		body, err := os.ReadFile(f)
		if err != nil {
			log.Printf("read %s: %v", f, err)
			continue
		}
		slug := "state-of-ai-builders-" + strings.TrimSuffix(filepath.Base(f), ".md")
		title, excerpt := parseTitleExcerpt(string(body))
		portable := mdToPortable(string(body))
		publishedAt := publishedAtFor(filepath.Base(f))
		tags := `["State of AI Builders","City Rankings","University Rankings","Ambassadors","monthly-report"]`
		_, err = pool.Exec(ctx, `
			INSERT INTO blog_posts (title, slug, description, body, author_id, category_id, tags, status, published_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, 'published', $8, NOW(), NOW())
			ON CONFLICT (slug) DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				body = EXCLUDED.body,
				category_id = EXCLUDED.category_id,
				tags = EXCLUDED.tags,
				published_at = EXCLUDED.published_at,
				updated_at = NOW()
		`, title, slug, excerpt, portable, *authorID, categoryReports, tags, publishedAt)
		if err != nil {
			log.Printf("insert %s: %v", f, err)
			continue
		}
		fmt.Printf("posted %s → %s\n", filepath.Base(f), slug)
		posted++
	}
	fmt.Printf("\n%d reports posted\n", posted)
}

// parseTitleExcerpt pulls the H1 title and the first TL;DR bullet for
// the excerpt.
func parseTitleExcerpt(md string) (title, excerpt string) {
	lines := strings.Split(md, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "# ") {
			title = strings.TrimPrefix(l, "# ")
			break
		}
	}
	// First bullet after "## TL;DR" becomes the excerpt.
	inTLDR := false
	for _, l := range lines {
		if strings.HasPrefix(l, "## TL;DR") {
			inTLDR = true
			continue
		}
		if inTLDR && strings.HasPrefix(l, "- ") {
			excerpt = strings.TrimPrefix(l, "- ")
			// Strip markdown bold markers for the excerpt.
			excerpt = strings.ReplaceAll(excerpt, "**", "")
			if len(excerpt) > 250 {
				excerpt = excerpt[:250] + "…"
			}
			break
		}
	}
	if title == "" {
		title = "State of AI Builders"
	}
	return
}

// publishedAtFor returns the 1st of the month after the report's month.
// e.g. 2026-06.md → 2026-07-01T09:00:00Z.
func publishedAtFor(filename string) time.Time {
	base := strings.TrimSuffix(filename, ".md") // "2026-06"
	t, err := time.Parse("2006-01", base)
	if err != nil {
		return time.Now()
	}
	return time.Date(t.Year(), t.Month()+1, 1, 9, 0, 0, 0, time.UTC)
}

// mdToPortable converts the report Markdown to the blog app's portable-text
// body format. Each block becomes one line; lines are joined with `\`.
// The blog app's renderer treats `\` as the block separator, so any line
// in the report (heading, paragraph, table row, list item) becomes a
// separate block. Long paragraphs are kept as one block — the blog app
// renders the line as a paragraph.
func mdToPortable(md string) string {
	lines := strings.Split(md, "\n")
	var blocks []string
	var current []string

	flush := func() {
		if len(current) == 0 {
			return
		}
		text := strings.Join(current, " ")
		text = strings.TrimSpace(text)
		if text != "" {
			blocks = append(blocks, text)
		}
		current = nil
	}

	for _, line := range lines {
		raw := strings.TrimRight(line, " \t\r")
		if strings.TrimSpace(raw) == "" {
			flush()
			continue
		}
		if strings.HasPrefix(raw, "#") {
			flush()
			blocks = append(blocks, raw)
			continue
		}
		if strings.HasPrefix(raw, "-") || strings.HasPrefix(raw, "*") || strings.HasPrefix(raw, ">") {
			flush()
			blocks = append(blocks, raw)
			continue
		}
		if strings.HasPrefix(raw, "|") {
			flush()
			blocks = append(blocks, raw)
			continue
		}
		if raw == "---" {
			flush()
			blocks = append(blocks, raw)
			continue
		}
		// Plain paragraph line.
		current = append(current, raw)
	}
	flush()

	// Join with `\` separator (the blog app's portable-text block delimiter).
	var b strings.Builder
	for i, blk := range blocks {
		if i > 0 {
			b.WriteString("\\\n")
		}
		b.WriteString(blk)
		b.WriteString("\\")
	}
	return b.String()
}
