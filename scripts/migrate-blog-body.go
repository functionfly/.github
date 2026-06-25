// Script: migrate-blog-body
// One-time migration: convert blog post body from TipTap JSON to Markdown
//
// Usage: go run scripts/migrate-blog-body.go
//
// This script reads the TipTap JSON from blog_posts.body, converts it to Markdown,
// and writes it back. The column type is changed by the SQL migration
// 20260620150931_blog_body_to_markdown.up.sql.

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type TipTapNode struct {
	Type     string       `json:"type,omitempty"`
	Text     string       `json:"text,omitempty"`
	Children []TipTapNode `json:"children,omitempty"`
	Content  []TipTapNode `json:"content,omitempty"`
	Attrs    TipTapAttrs  `json:"attrs,omitempty"`
	Marks    []TipTapMark `json:"marks,omitempty"`
}

type TipTapAttrs struct {
	Level int `json:"level,omitempty"`
}

type TipTapMark struct {
	Type string `json:"type"`
	Href string `json:"href,omitempty"`
}

func nodeToMarkdown(node TipTapNode) string {
	switch node.Type {
	case "text":
		text := node.Text
		for _, mark := range node.Marks {
			switch mark.Type {
			case "bold":
				text = "**" + text + "**"
			case "italic":
				text = "*" + text + "*"
			case "code":
				text = "`" + text + "`"
			case "strike":
				text = "~~" + text + "~~"
			case "link":
				text = "[" + text + "](" + mark.Href + ")"
			}
		}
		return text
	case "paragraph":
		children := node.Children
		if len(children) == 0 {
			children = node.Content
		}
		inner := childrenToMarkdown(children)
		if inner == "" {
			return "\n\n"
		}
		return "\n\n" + inner + "\n\n"
	case "heading":
		level := node.Attrs.Level
		if level == 0 {
			level = 2
		}
		children := node.Children
		if len(children) == 0 {
			children = node.Content
		}
		inner := childrenToMarkdown(children)
		prefix := strings.Repeat("#", level)
		return "\n\n" + prefix + " " + inner + "\n\n"
	case "bulleted-list", "bulletList":
		return listToMarkdown(node, "- ", false)
	case "ordered-list", "orderedList":
		return listToMarkdown(node, "", true)
	case "list-item", "listItem":
		children := node.Children
		if len(children) == 0 {
			children = node.Content
		}
		return childrenToMarkdown(children)
	case "code-block", "codeBlock":
		children := node.Children
		if len(children) == 0 {
			children = node.Content
		}
		inner := childrenToMarkdown(children)
		return "\n\n```\n" + inner + "\n```\n\n"
	case "blockquote", "quote":
		children := node.Children
		if len(children) == 0 {
			children = node.Content
		}
		inner := childrenToMarkdown(children)
		return "\n\n> " + strings.ReplaceAll(inner, "\n", "\n> ") + "\n\n"
	case "horizontalRule":
		return "\n\n---\n\n"
	case "doc":
		return childrenToMarkdown(node.Content)
	}

	children := node.Children
	if len(children) == 0 {
		children = node.Content
	}
	return childrenToMarkdown(children)
}

func listToMarkdown(node TipTapNode, prefix string, numbered bool) string {
	children := node.Children
	if len(children) == 0 {
		children = node.Content
	}
	result := "\n"
	for i, item := range children {
		var innerChildren []TipTapNode
		if item.Type == "list-item" || item.Type == "listItem" {
			innerChildren = item.Children
			if len(innerChildren) == 0 {
				innerChildren = item.Content
			}
		} else {
			innerChildren = []TipTapNode{item}
		}
		inner := childrenToMarkdown(innerChildren)
		inner = strings.TrimSpace(inner)
		if numbered {
			result += fmt.Sprintf("%d. %s\n", i+1, inner)
		} else {
			result += prefix + inner + "\n"
		}
	}
	return result
}

func childrenToMarkdown(children []TipTapNode) string {
	result := ""
	for _, child := range children {
		result += nodeToMarkdown(child)
	}
	return result
}

func unescapeString(s string) string {
	result := strings.Builder{}
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
				i += 2
				continue
			case 't':
				result.WriteByte('\t')
				i += 2
				continue
			case 'r':
				result.WriteByte('\r')
				i += 2
				continue
			case '"':
				result.WriteByte('"')
				i += 2
				continue
			case '\\':
				result.WriteByte('\\')
				i += 2
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "functionfly"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, body FROM blog_posts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type Post struct {
		ID   string
		Body string
	}
	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Body); err != nil {
			log.Fatal(err)
		}
		posts = append(posts, p)
	}

	for _, p := range posts {
		body := p.Body
		// Unescape if it's a JSON string
		if len(body) > 0 && body[0] == '"' {
			body = unescapeString(body)
		}

		var arrayNode []TipTapNode
		if err := json.Unmarshal([]byte(body), &arrayNode); err == nil {
			markdown := childrenToMarkdown(arrayNode)
			markdown = strings.TrimSpace(markdown)
			if _, err := db.Exec("UPDATE blog_posts SET body = $1 WHERE id = $2", markdown, p.ID); err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			fmt.Printf("Updated %s (array): %d chars\n", p.ID, len(markdown))
			continue
		}

		var docNode TipTapNode
		if err := json.Unmarshal([]byte(body), &docNode); err == nil {
			markdown := nodeToMarkdown(docNode)
			markdown = strings.TrimSpace(markdown)
			if _, err := db.Exec("UPDATE blog_posts SET body = $1 WHERE id = $2", markdown, p.ID); err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			fmt.Printf("Updated %s (doc): %d chars\n", p.ID, len(markdown))
			continue
		}

		// Already plain text/markdown
		cleaned := strings.TrimSpace(body)
		if _, err := db.Exec("UPDATE blog_posts SET body = $1 WHERE id = $2", cleaned, p.ID); err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("Updated %s (text): %d chars\n", p.ID, len(cleaned))
	}
}
