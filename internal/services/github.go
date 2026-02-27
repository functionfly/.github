package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	ID          int       `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

// GitHubService handles GitHub API operations
type GitHubService struct {
	client        *http.Client
	baseURL       string
	token         string
	owner         string
	repo          string
	releaseParser *ReleaseParser
}

// ReleaseParser parses GitHub release bodies into structured changelog data
type ReleaseParser struct {
	versionRegex *regexp.Regexp
	changeRegex  *regexp.Regexp
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(owner, repo, token string) *GitHubService {
	return &GitHubService{
		client: &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.github.com",
		token:   token,
		owner:   owner,
		repo:    repo,
		releaseParser: &ReleaseParser{
			versionRegex: regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`),
			changeRegex:  regexp.MustCompile(`(?m)^###?\s+(.+)$|^-\s+(.+)$|^•\s+(.+)$`),
		},
	}
}

// FetchReleases fetches releases from GitHub API
func (s *GitHubService) FetchReleases(ctx context.Context, perPage int) ([]GitHubRelease, error) {
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=%d", s.baseURL, s.owner, s.repo, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	return releases, nil
}

// SyncReleases syncs GitHub releases with changelog entries
func (s *GitHubService) SyncReleases(ctx context.Context, repo storage.Repository) error {
	logrus.Info("Starting GitHub releases sync")

	releases, err := s.FetchReleases(ctx, 50)
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	synced := 0
	for _, release := range releases {
		// Skip drafts and prereleases
		if release.Draft || release.Prerelease {
			continue
		}

		// Check if release already exists
		existing, err := repo.GetChangelogEntryByVersion(release.TagName)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			logrus.WithError(err).WithField("version", release.TagName).Warn("Failed to check existing changelog entry")
			continue
		}

		if existing != nil {
			// Update if needed
			if existing.GitHubID == nil || *existing.GitHubID != strconv.Itoa(release.ID) {
				updates := map[string]interface{}{
					"github_id":    strconv.Itoa(release.ID),
					"release_url": release.HTMLURL,
				}
				if _, err := repo.UpdateChangelogEntry(ctx, existing.ID, updates); err != nil {
					logrus.WithError(err).WithField("version", release.TagName).Warn("Failed to update changelog entry")
				}
			}
			continue
		}

		// Parse release into changelog entry
		entry, err := s.parseReleaseToChangelogEntry(release)
		if err != nil {
			logrus.WithError(err).WithField("version", release.TagName).Warn("Failed to parse release")
			continue
		}

		// Create new changelog entry
		if _, err := repo.CreateChangelogEntry(ctx, entry); err != nil {
			logrus.WithError(err).WithField("version", release.TagName).Warn("Failed to create changelog entry")
			continue
		}

		synced++
		logrus.WithField("version", release.TagName).Info("Synced GitHub release")
	}

	logrus.WithField("count", synced).Info("Completed GitHub releases sync")
	return nil
}

// parseReleaseToChangelogEntry parses a GitHub release into a changelog entry
func (s *GitHubService) parseReleaseToChangelogEntry(release GitHubRelease) (*storage.ChangelogEntry, error) {
	entry := &storage.ChangelogEntry{
		Version:     release.TagName,
		Date:        release.PublishedAt,
		Title:       release.Name,
		Description: s.extractDescription(release.Body),
		ReleaseURL:  &release.HTMLURL,
		GitHubID:    stringPtr(strconv.Itoa(release.ID)),
		IsPublished: true,
		Type:        s.determineReleaseType(release.TagName),
		Changes:     s.parseChanges(release.Body),
	}

	return entry, nil
}

// determineReleaseType determines the release type based on version
func (s *GitHubService) determineReleaseType(version string) string {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	matches := s.releaseParser.versionRegex.FindStringSubmatch(version)
	if len(matches) != 4 {
		return "minor" // default
	}

	major, _ := strconv.Atoi(matches[1])
	_, _ = strconv.Atoi(matches[2]) // minor
	patch, _ := strconv.Atoi(matches[3])

	// Simple logic: if major > 0 or minor >= 10, consider major
	// if patch > 0, consider patch, else minor
	if major > 0 {
		return "major"
	}
	if patch > 0 {
		return "patch"
	}
	return "minor"
}

// extractDescription extracts description from release body
func (s *GitHubService) extractDescription(body string) string {
	lines := strings.Split(body, "\n")
	var description []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Stop at first heading or list
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			break
		}

		description = append(description, line)
	}

	return strings.Join(description, " ")
}

// parseChanges parses release body into structured changes
func (s *GitHubService) parseChanges(body string) []storage.ChangelogChange {
	var changes []storage.ChangelogChange

	// Split by sections (### headings)
	sections := s.splitByHeadings(body)

	for _, section := range sections {
		change := s.parseSection(section)
		if change != nil {
			changes = append(changes, *change)
		}
	}

	// If no sections found, create a default "Features" section
	if len(changes) == 0 && strings.TrimSpace(body) != "" {
		items := s.extractListItems(body)
		if len(items) > 0 {
			changes = append(changes, storage.ChangelogChange{
				Category: "Features",
				Icon:     "Sparkles",
				Items:    items,
			})
		}
	}

	return changes
}

// splitByHeadings splits release body by ### headings
func (s *GitHubService) splitByHeadings(body string) []string {
	lines := strings.Split(body, "\n")
	var sections []string
	var currentSection []string

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "###") {
			// Save previous section
			if len(currentSection) > 0 {
				sections = append(sections, strings.Join(currentSection, "\n"))
				currentSection = []string{}
			}
		}
		currentSection = append(currentSection, line)
	}

	// Add last section
	if len(currentSection) > 0 {
		sections = append(sections, strings.Join(currentSection, "\n"))
	}

	return sections
}

// parseSection parses a single section into a ChangelogChange
func (s *GitHubService) parseSection(section string) *storage.ChangelogChange {
	lines := strings.Split(section, "\n")
	if len(lines) == 0 {
		return nil
	}

	// Find heading
	var heading string
	var contentLines []string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "###") {
			heading = strings.TrimSpace(strings.TrimPrefix(line, "###"))
			contentLines = lines[i+1:]
			break
		}
	}

	if heading == "" {
		return nil
	}

	// Extract list items from content
	content := strings.Join(contentLines, "\n")
	items := s.extractListItems(content)

	if len(items) == 0 {
		return nil
	}

	return &storage.ChangelogChange{
		Category: s.normalizeCategory(heading),
		Icon:     s.getIconForCategory(heading),
		Items:    items,
	}
}

// extractListItems extracts list items from text
func (s *GitHubService) extractListItems(text string) []string {
	var items []string

	matches := s.releaseParser.changeRegex.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		// Find the non-empty capture group
		for _, capture := range match[1:] {
			if strings.TrimSpace(capture) != "" {
				items = append(items, strings.TrimSpace(capture))
				break
			}
		}
	}

	return items
}

// normalizeCategory normalizes section headings to standard categories
func (s *GitHubService) normalizeCategory(heading string) string {
	heading = strings.ToLower(strings.TrimSpace(heading))

	switch {
	case strings.Contains(heading, "feature") || strings.Contains(heading, "new"):
		return "Features"
	case strings.Contains(heading, "bug") || strings.Contains(heading, "fix"):
		return "Bug Fixes"
	case strings.Contains(heading, "security"):
		return "Security"
	case strings.Contains(heading, "performance") || strings.Contains(heading, "perf"):
		return "Performance"
	case strings.Contains(heading, "breaking") || strings.Contains(heading, "break"):
		return "Breaking Changes"
	case strings.Contains(heading, "deprecat"):
		return "Deprecations"
	default:
		// Capitalize first letter
		return strings.Title(heading)
	}
}

// getIconForCategory returns appropriate icon for category
func (s *GitHubService) getIconForCategory(category string) string {
	category = strings.ToLower(strings.TrimSpace(category))

	switch {
	case strings.Contains(category, "feature") || strings.Contains(category, "new"):
		return "Sparkles"
	case strings.Contains(category, "bug") || strings.Contains(category, "fix"):
		return "Bug"
	case strings.Contains(category, "security"):
		return "Shield"
	case strings.Contains(category, "performance") || strings.Contains(category, "perf"):
		return "Zap"
	case strings.Contains(category, "breaking") || strings.Contains(category, "break"):
		return "AlertTriangle"
	default:
		return "CheckCircle"
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}