package github

import (
	"time"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

type GitHubRepo struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	FullName        string     `json:"full_name"`
	Owner           GitHubUser `json:"owner"`
	Description     string     `json:"description"`
	DefaultBranch   string     `json:"default_branch"`
	Language        string     `json:"language"`
	Private         bool       `json:"private"`
	Fork            bool       `json:"fork"`
	Archived        bool       `json:"archived"`
	Topics          []string   `json:"topics"`
	StargazersCount int        `json:"stargazers_count"`
	ForksCount      int        `json:"forks_count"`
	Size            int        `json:"size"`
	PushedAt        *time.Time `json:"pushed_at"`
	HTMLURL         string     `json:"html_url"`
	CloneURL        string     `json:"clone_url"`
	SSHURL          string     `json:"ssh_url"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type GitHubBranch struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
	Protected bool `json:"protected"`
}

type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

type GitHubTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int    `json:"size"`
}

type GitHubTree struct {
	SHA       string            `json:"sha"`
	Tree      []GitHubTreeEntry `json:"tree"`
	Truncated bool              `json:"truncated"`
}

type GitHubWebhookRequest struct {
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
	Config struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Secret      string `json:"secret"`
	} `json:"config"`
}

type GitHubWebhookResponse struct {
	ID     int64    `json:"id"`
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Events []string `json:"events"`
	Config struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
	} `json:"config"`
}

type CommitStatusRequest struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url,omitempty"`
	Description string `json:"description,omitempty"`
	Context     string `json:"context"`
}

type ListReposOptions struct {
	Page      int
	PerPage   int
	Sort      string
	Direction string
	Type      string
}

type Languages map[string]float64

type DetectedFunction struct {
	Name         string                 `json:"name"`
	EntryPoint   string                 `json:"entry_point"`
	Runtime      string                 `json:"runtime"`
	SubDirectory string                 `json:"sub_directory,omitempty"`
	Confidence   float64                `json:"confidence"`
	Strategy     string                 `json:"strategy"`
	Manifest     map[string]interface{} `json:"manifest,omitempty"`
	Dependencies *DependencyInfo        `json:"dependencies,omitempty"`
}

type DependencyInfo struct {
	Manager  string   `json:"manager"`
	Lockfile string   `json:"lockfile,omitempty"`
	Packages []string `json:"packages,omitempty"`
}

type ScanResult struct {
	Functions            []DetectedFunction `json:"functions"`
	PrimaryRuntime       string             `json:"primary_runtime"`
	OverallConfidence    float64            `json:"overall_confidence"`
	StrategyUsed         string             `json:"strategy_used"`
	Warnings             []string           `json:"warnings"`
	EstimatedImportTimeS int                `json:"estimated_import_time_seconds"`
	EstimatedCostUSD     float64            `json:"estimated_cost_usd"`
}

type WebhookPushEvent struct {
	Ref    string `json:"ref"`
	Before string `json:"before"`
	After  string `json:"after"`
	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	Commits []struct {
		ID       string   `json:"id"`
		Message  string   `json:"message"`
		Added    []string `json:"added"`
		Removed  []string `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"commits"`
}

type WebhookPROvent struct {
	Action string `json:"action"`
	Number int    `json:"number"`
	PullRequest struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Head   struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"base"`
		Merged bool `json:"merged"`
	} `json:"pull_request"`
	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}
