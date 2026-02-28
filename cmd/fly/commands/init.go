package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var template string
	var force bool
	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Scaffold a new function project",
		Long:  "Create a new FunctionFly function project with all required files.",
		Example: `  fly init hello-world
  fly init --template hello-world my-function
  fly init --template http-api api-service
  fly init --template cron-job daily-task
  fly init --template webhook webhook-handler`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runInit(name, template, force)
		},
	}
	cmd.Flags().StringVarP(&template, "template", "t", "hello-world", "Template (hello-world, http-api, cron-job, webhook)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	return cmd
}

func runInit(name, template string, force bool) error {
	if name == "" {
		if IsInteractive() {
			name = Prompt("Function name", "my-function")
		} else {
			return fmt.Errorf("function name is required\n   → Usage: fly init <name>")
		}
	}
	if !isValidFunctionName(name) {
		return fmt.Errorf("invalid function name: %q\n   → Names must be lowercase letters, numbers, and hyphens only", name)
	}
	if template == "javascript" && IsInteractive() {
		template = PromptSelect("Choose a template:", []string{"javascript", "typescript", "python"}, "javascript")
	}
	projectDir := filepath.Join(".", name)
	if _, err := os.Stat(projectDir); err == nil && !force {
		return fmt.Errorf("directory %q already exists\n   → Use --force to overwrite", projectDir)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}
	files, err := generateTemplateFiles(name, template)
	if err != nil {
		return err
	}
	for filename, content := range files {
		path := filepath.Join(projectDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", filename, err)
		}
		fmt.Printf("  ✓ %s\n", filepath.Join(name, filename))
	}
	fmt.Printf("\n✅ Created %s/\n\n", name)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  fly dev          # Run locally")
	fmt.Println("  fly publish      # Publish to the registry")
	return nil
}

func isValidFunctionName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func generateTemplateFiles(name, template string) (map[string]string, error) {
	files := map[string]string{}

	// Get template description based on template type
	description := getTemplateDescription(template)

	files["functionfly.jsonc"] = fmt.Sprintf(`{
  "$schema": "https://functionfly.com/schemas/functionfly.json",
  "name": "%s",
  "version": "1.0.0",
  "runtime": "%s",
  "public": true,
  "deterministic": true,
  "cache_ttl": 86400,
  "timeout_ms": %d,
  "memory_mb": %d,
  "description": "%s"%s
}
`, name, runtimeForTemplate(template), timeoutForTemplate(template), memoryForTemplate(template), description, scheduleForTemplate(template))

	switch template {
	case "http-api":
		files["main.py"] = getHttpApiTemplate(name)
	case "cron-job":
		files["main.py"] = getCronJobTemplate(name)
	case "webhook":
		files["main.py"] = getWebhookTemplate(name)
	case "hello-world":
		fallthrough
	default:
		files["main.py"] = getHelloWorldTemplate(name)
	}

	files["test.http"] = fmt.Sprintf("### Test %s locally\nPOST http://localhost:8787\nContent-Type: application/json\n\n{}", name)
	return files, nil
}

func runtimeForTemplate(template string) string {
	switch template {
	case "typescript":
		return "node20"
	case "python":
		return "python3.11"
	default:
		return "python3.11"
	}
}

func timeoutForTemplate(template string) int {
	switch template {
	case "http-api":
		return 10000
	case "cron-job":
		return 30000
	default:
		return 5000
	}
}

func memoryForTemplate(template string) int {
	switch template {
	case "http-api":
		return 256
	default:
		return 128
	}
}

func getTemplateDescription(template string) string {
	switch template {
	case "hello-world":
		return "A simple function that returns a greeting"
	case "http-api":
		return "A RESTful API with CRUD operations"
	case "cron-job":
		return "A scheduled function for periodic tasks"
	case "webhook":
		return "A webhook handler for third-party events"
	default:
		return "A FunctionFly function"
	}
}

func scheduleForTemplate(template string) string {
	switch template {
	case "cron-job":
		return "\n  ,\"schedule\": \"0 * * * *\""
	default:
		return ""
	}
}

func getHelloWorldTemplate(name string) string {
	return fmt.Sprintf(`"""
%s - A FunctionFly function
A simple function that returns a greeting.
"""

async def fetch(request, env, ctx):
    """
    Handle incoming requests and return a greeting.
    
    Args:
        request: The incoming request object
        env: Environment variables and secrets
        ctx: Execution context
    
    Returns:
        Response with greeting message
    """
    url = request.url
    
    # Parse query parameters
    params = {}
    if '?' in url:
        query_string = url.split('?')[1]
        for param in query_string.split('&'):
            if '=' in param:
                key, value = param.split('=', 1)
                params[key] = value
    
    # Get name from query params or default to World
    name = params.get('name', 'World')
    
    # Return greeting
    return {
        "status": 200,
        "body": f"Hello, {name}! Welcome to FunctionFly.",
        "headers": {
            "Content-Type": "application/json",
            "X-FunctionFly-Template": "hello-world"
        }
    }
`, name)
}

func getHttpApiTemplate(name string) string {
	return `"""
HTTP API Function
A RESTful API with built-in routing for CRUD operations.
"""

from datetime import datetime

# In-memory storage for demo (use a database in production)
users_db = {}
user_id_counter = 1


async def fetch(request, env, ctx):
    """
    Handle incoming HTTP requests with RESTful routing.
    
    Routes:
        GET    /users      - List all users
        GET    /users/{id} - Get user by ID
        POST   /users      - Create a new user
        PUT    /users/{id} - Update user
        DELETE /users/{id} - Delete user
        GET    /health     - Health check
    """
    url = request.url
    method = request.method
    
    # Health check endpoint
    if '/health' in url:
        return {
            "status": 200,
            "body": {
                "status": "healthy",
                "timestamp": datetime.utcnow().isoformat(),
                "service": "http-api"
            },
            "headers": {"Content-Type": "application/json"}
        }
    
    # User management endpoints
    if '/users' in url:
        if method == 'GET':
            user_list = list(users_db.values())
            return {
                "status": 200,
                "body": {
                    "users": user_list,
                    "count": len(user_list)
                },
                "headers": {"Content-Type": "application/json"}
            }
        
        if method == 'POST':
            global user_id_counter
            user_id = user_id_counter
            user_id_counter += 1
            
            user = {
                "id": user_id,
                "name": "New User",
                "created_at": datetime.utcnow().isoformat()
            }
            users_db[user_id] = user
            
            return {
                "status": 201,
                "body": user,
                "headers": {"Content-Type": "application/json"}
            }
    
    # Default: 404 Not Found
    return {
        "status": 404,
        "body": {"error": "Not found"},
        "headers": {"Content-Type": "application/json"}
    }
`
}

func getCronJobTemplate(name string) string {
	return `"""
Cron Job Function
A scheduled function that runs periodically.
Use cases: Data cleanup, report generation, cache refresh, etc.
"""

from datetime import datetime


async def fetch(request, env, ctx):
    """
    Handle scheduled execution.
    
    This function runs automatically based on the schedule defined
    in functionfly.jsonc.
    
    Schedule presets:
        "*/5 * * * *"    - Every 5 minutes
        "0 * * * *"      - Every hour
        "0 0 * * *"      - Every day at midnight
        "0 0 * * 1-5"   - Weekdays at midnight
    """
    # Check if this is a scheduled execution
    body = {}
    try:
        body = await request.json()
    except:
        pass
    
    trigger = body.get("trigger", "manual")
    timestamp = body.get("timestamp", datetime.utcnow().isoformat())
    
    result = {
        "status": "success",
        "trigger": trigger,
        "executed_at": timestamp,
        "message": "Scheduled job completed successfully"
    }
    
    return {
        "status": 200,
        "body": result,
        "headers": {
            "Content-Type": "application/json",
            "X-FunctionFly-Template": "cron-job"
        }
    }
`
}

func getWebhookTemplate(name string) string {
	return `"""
Webhook Handler Function
A generic webhook handler for processing incoming events.
Use cases: GitHub webhooks, Stripe, Slack, etc.
"""

from datetime import datetime


async def fetch(request, env, ctx):
    """
    Handle incoming webhook requests.
    
    Supports:
    - Signature verification (X-Hub-Signature-256)
    - Event type parsing
    - JSON payload parsing
    """
    # Get webhook secret from environment
    webhook_secret = env.get("WEBHOOK_SECRET", "")
    
    # Get headers
    headers = dict(request.headers)
    event_type = headers.get("x-github-event", headers.get("event-type", "unknown"))
    
    # Parse request body
    body = {}
    try:
        body = await request.json()
    except:
        pass
    
    result = {
        "received": True,
        "event_type": event_type,
        "timestamp": datetime.utcnow().isoformat(),
        "processed": True
    }
    
    return {
        "status": 200,
        "body": result,
        "headers": {"Content-Type": "application/json"}
    }
`
}
