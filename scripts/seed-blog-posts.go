package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type CategoryData struct {
	Title       string
	Slug        string
	Description string
	Color       string
	Icon        string
	Order       int
}

type BlogPostData struct {
	Title        string
	Slug         string
	Description  string
	Body         interface{} // Rich text JSON structure
	CategorySlug string
	Tags         []string
	PublishedAt  string
}

// Categories to create
var categories = []CategoryData{
	{
		Title:       "Platform",
		Slug:        "platform",
		Description: "Core platform features and architecture",
		Color:       "#3B82F6",
		Icon:        "cpu",
		Order:       1,
	},
	{
		Title:       "Security",
		Slug:        "security",
		Description: "Security features and best practices",
		Color:       "#EF4444",
		Icon:        "shield",
		Order:       2,
	},
	{
		Title:       "AI & ML",
		Slug:        "ai-ml",
		Description: "AI and machine learning capabilities",
		Color:       "#8B5CF6",
		Icon:        "brain",
		Order:       3,
	},
	{
		Title:       "Tutorials",
		Slug:        "tutorials",
		Description: "Step-by-step guides and tutorials",
		Color:       "#10B981",
		Icon:        "book-open",
		Order:       4,
	},
	{
		Title:       "Case Studies",
		Slug:        "case-studies",
		Description: "Real-world success stories and examples",
		Color:       "#F59E0B",
		Icon:        "users",
		Order:       5,
	},
}

// Blog posts with rich content and categories
var blogPosts = []BlogPostData{
	{
		Title:       "Welcome to FunctionFly: Serverless Infrastructure for the AI Era",
		Slug:        "welcome-to-functionfly",
		Description: "Introducing FunctionFly, the serverless platform purpose-built for AI applications with Flywheel Network, zero-knowledge secrets, and AI-first architecture.",
		Body: []map[string]interface{}{
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Welcome to FunctionFly, the serverless platform designed for the AI era. We're building the infrastructure that enables developers to create, deploy, and monetize AI-powered applications with unprecedented ease and security."},
				},
			},
			{
				"type":  "heading",
				"level": 2,
				"children": []map[string]interface{}{
					{"text": "What Makes FunctionFly Different?"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "FunctionFly isn't just another serverless platform—it's purpose-built for the AI-native world. Our core innovations address the fundamental challenges of building AI applications at scale:"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "🔄 **Flywheel Network™**: A proof-of-execution knowledge network where every function execution becomes verifiable, composable knowledge. Problems are structured, solutions are executable, and AI agents collaborate in open debates."},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "🔐 **Zero-Knowledge Secrets Vault**: Client-side encrypted secrets that scale from free to enterprise-grade without compromising security. Your data never touches our servers in plaintext."},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "🧠 **AI-First Architecture**: Built for AI agents, RAG systems, and autonomous workflows with features like State Fabric for durable memory and CCP (Compute Capsules Protocol) for verifiable execution."},
				},
			},
		},
		CategorySlug: "platform",
		Tags:         []string{"functionfly", "serverless", "ai", "platform", "introduction", "flywheel-network", "secrets-vault"},
		PublishedAt:  "2024-01-10T09:00:00Z",
	},
	{
		Title:       "Introducing Flywheel Network™: Proof-of-Execution Knowledge Network",
		Slug:        "introducing-flywheel-network",
		Description: "Flywheel Network brings proof-of-execution to function calls, creating a decentralized knowledge network that rewards computational contributions.",
		Body: []map[string]interface{}{
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Flywheel Network™ is FunctionFly's revolutionary proof-of-execution knowledge network. It transforms every function execution into verifiable, composable knowledge—creating a self-reinforcing ecosystem where problems are structured, solutions are executable, and AI agents collaborate in open debates."},
				},
			},
			{
				"type":  "heading",
				"level": 2,
				"children": []map[string]interface{}{
					{"text": "The Problem with Traditional Development"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Today's development is fragmented. Code repositories store static snapshots. Documentation becomes outdated. Knowledge lives in tribal repositories—Slack threads, Notion docs, and institutional memory. When AI agents try to collaborate, they lack shared context and verifiable execution history."},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Flywheel Network solves this by treating execution as knowledge. Every function run becomes a data point. Every solution becomes composable. Every agent interaction becomes part of a growing knowledge graph."},
				},
			},
		},
		CategorySlug: "platform",
		Tags:         []string{"flywheel", "decentralized", "proof-of-execution", "knowledge-network", "ai-collaboration"},
		PublishedAt:  "2024-01-20T11:00:00Z",
	},
	{
		Title:       "Introducing Secrets Vault: Zero-Knowledge Secrets That Scale",
		Slug:        "introducing-secrets-vault",
		Description: "Secrets Vault provides cryptographic protection for sensitive data with zero-knowledge architecture that scales across distributed environments.",
		Body: []map[string]interface{}{
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Traditional secrets management solutions struggle with scale and security in distributed environments. Secrets Vault changes this paradigm by implementing zero-knowledge cryptography that protects your data from the moment it's created until it's used."},
				},
			},
			{
				"type":  "heading",
				"level": 2,
				"children": []map[string]interface{}{
					{"text": "Zero-Knowledge Architecture"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Your secrets are encrypted client-side before they ever reach our servers. The encryption keys never leave your device, ensuring that even if our infrastructure is compromised, your sensitive data remains secure."},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "This approach provides enterprise-grade security while maintaining the simplicity and scalability you need for modern applications."},
				},
			},
		},
		CategorySlug: "security",
		Tags:         []string{"security", "secrets", "cryptography", "zero-knowledge", "encryption"},
		PublishedAt:  "2024-01-25T14:00:00Z",
	},
	{
		Title:       "AI Agent Integration: Building for the AI Era",
		Slug:        "ai-agent-integration-building-for-the-ai-era",
		Description: "FunctionFly provides native AI agent integration capabilities that enable seamless interaction between AI systems and serverless functions.",
		Body: []map[string]interface{}{
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "The AI revolution requires infrastructure that can keep pace with intelligent systems. FunctionFly's AI agent integration provides this foundation with native support for agent-to-agent communication, persistent memory, and verifiable execution."},
				},
			},
			{
				"type":  "heading",
				"level": 2,
				"children": []map[string]interface{}{
					{"text": "Agent-First Design"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "FunctionFly is designed from the ground up for AI agents. Our platform understands agent communication patterns, provides persistent state management through State Fabric, and ensures deterministic execution for reliable agent behavior."},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Build autonomous workflows, RAG systems, and multi-agent applications with confidence."},
				},
			},
		},
		CategorySlug: "ai-ml",
		Tags:         []string{"ai", "agents", "integration", "autonomous", "rag", "state-fabric"},
		PublishedAt:  "2024-02-05T10:30:00Z",
	},
	{
		Title:       "Getting Started: Deploy Your First FunctionFly Function",
		Slug:        "getting-started-deploy-your-first-functionfly-function",
		Description: "A comprehensive guide to deploying your first function on FunctionFly, from setup to production deployment.",
		Body: []map[string]interface{}{
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Getting started with FunctionFly is designed to be as simple as possible while providing the power and flexibility you need. This guide will walk you through deploying your first function from development to production."},
				},
			},
			{
				"type":  "heading",
				"level": 2,
				"children": []map[string]interface{}{
					{"text": "Prerequisites"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Before we begin, make sure you have:"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "• A FunctionFly account"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "• Node.js 18+ installed"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "• Basic familiarity with JavaScript/TypeScript"},
				},
			},
		},
		CategorySlug: "tutorials",
		Tags:         []string{"tutorial", "getting-started", "deployment", "javascript", "guide"},
		PublishedAt:  "2024-02-10T08:00:00Z",
	},
	{
		Title:       "From Side Project to SaaS: How Alex Built an AI Content Moderator on FunctionFly",
		Slug:        "from-side-project-to-saas-how-alex-built-ai-content-moderator",
		Description: "Real-world success story: How one developer transformed a side project into a successful SaaS using FunctionFly's AI capabilities.",
		Body: []map[string]interface{}{
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Alex started with a simple idea: an AI-powered content moderation service. What started as a weekend project became a successful SaaS business built entirely on FunctionFly."},
				},
			},
			{
				"type":  "heading",
				"level": 2,
				"children": []map[string]interface{}{
					{"text": "The Spark"},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "Like many successful projects, this one started with a personal problem. Alex was building a social platform and struggling with content moderation at scale. Existing solutions were either too expensive or not sophisticated enough."},
				},
			},
			{
				"type": "paragraph",
				"children": []map[string]interface{}{
					{"text": "\"I built the first version in a weekend using FunctionFly's AI integration features,\" Alex recalls. \"The zero-knowledge secrets made it easy to integrate with multiple AI providers without compromising security.\""},
				},
			},
		},
		CategorySlug: "case-studies",
		Tags:         []string{"case-study", "success-story", "ai", "saas", "content-moderation"},
		PublishedAt:  "2024-02-15T12:00:00Z",
	},
}

func run() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost:5434/functionfly?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🌱 Starting blog post seeding for main database...")

	now := time.Now()

	// Create categories first
	fmt.Println("📂 Creating categories...")
	categoryMap := make(map[string]uuid.UUID)

	for _, catData := range categories {
		// Check if category exists
		var existingID uuid.UUID
		err := db.QueryRow("SELECT id FROM blog_categories WHERE slug = $1", catData.Slug).Scan(&existingID)
		if err == nil {
			fmt.Printf("Category '%s' already exists, using existing.\n", catData.Title)
			categoryMap[catData.Slug] = existingID
			continue
		}

		// Create category
		categoryID := uuid.New()
		_, err = db.Exec(`
			INSERT INTO blog_categories (id, title, slug, description, color, icon, "order", created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			categoryID, catData.Title, catData.Slug, catData.Description,
			catData.Color, catData.Icon, catData.Order, now, now)

		if err != nil {
			log.Printf("Error creating category '%s': %v", catData.Title, err)
			continue
		}

		fmt.Printf("✓ Created category: %s\n", catData.Title)
		categoryMap[catData.Slug] = categoryID
	}

	// Create blog posts
	fmt.Println("📝 Creating blog posts...")
	insertedCount := 0

	for _, postData := range blogPosts {
		// Check if post already exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM blog_posts WHERE slug = $1)", postData.Slug).Scan(&exists)
		if err != nil {
			log.Printf("Error checking if post exists: %v", err)
			continue
		}

		if exists {
			fmt.Printf("Post with slug '%s' already exists, skipping.\n", postData.Slug)
			continue
		}

		// Parse published date
		publishedAt := now
		if postData.PublishedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, postData.PublishedAt); err == nil {
				publishedAt = parsed
			}
		}

		// Convert body to JSON
		bodyJSON, err := json.Marshal(postData.Body)
		if err != nil {
			log.Printf("Error marshaling body for post '%s': %v", postData.Title, err)
			continue
		}

		// Insert post
		_, err = db.Exec(`
			INSERT INTO blog_posts (
				id, title, slug, content, excerpt, author, tags,
				is_published, published_at, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
			)`,
			uuid.New(),
			postData.Title,
			postData.Slug,
			string(bodyJSON),        // content
			postData.Description,    // excerpt
			"FunctionFly Team",      // Default author
			pq.Array(postData.Tags), // tags as TEXT[] array
			true,                    // is_published
			publishedAt,
			now,
			now,
		)

		if err != nil {
			log.Printf("Error inserting post '%s': %v", postData.Title, err)
			continue
		}

		fmt.Printf("✓ Inserted blog post: %s\n", postData.Title)
		insertedCount++
	}

	fmt.Printf("✅ Seeding complete! Created %d categories and inserted %d new blog posts into main database.\n", len(categoryMap), insertedCount)
}

func main() {
	run()
}
