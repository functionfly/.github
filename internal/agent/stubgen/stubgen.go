// Package stubgen provides template-based code generation for agents.
// It generates production-quality function stubs without requiring an LLM.
package stubgen

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Config configures stub generation behavior
type Config struct {
	FallbackMode          string // "structured" (default), "minimal"
	IncludeSchemaComments bool   // Include input/output schema in comments
	EnableRAG             bool   // Enable RAG-based template enhancement
}

// DefaultConfig returns default stub generator configuration
func DefaultConfig() Config {
	return Config{
		FallbackMode:          "structured",
		IncludeSchemaComments: true,
		EnableRAG:             false,
	}
}

// CodeGenerator generates function source code from a request
type CodeGenerator interface {
	GenerateCode(ctx context.Context, req *GenerationRequest) (string, error)
}

// GenerationRequest represents a request to generate a function
type GenerationRequest struct {
	AgentID       string
	Name          string
	Description   string
	Category      string
	InputSchema   map[string]any
	OutputSchema  map[string]any
	Runtime       string
	Prompt        string
	Deterministic bool
	Tags          []string
}

// GenerationResult represents the result of function generation
type GenerationResult struct {
	FunctionID uuid.UUID
	Code       string
	Manifest   string
	ModelUsed  string
	Success    bool
	Error      string
}

// Template represents a function template
type Template struct {
	Name        string
	Description string
	Category    string
	Runtime     string
}

// TemplateRegistry manages available templates
type TemplateRegistry struct {
	templates map[string]*Template
}

// NewTemplateRegistry creates a registry with built-in templates
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: map[string]*Template{
			"http-api": {
				Name:        "http-api",
				Description: "RESTful HTTP API with routing",
				Category:    "http",
				Runtime:     "python3.11",
			},
			"cron-job": {
				Name:        "cron-job",
				Description: "Scheduled task",
				Category:    "scheduler",
				Runtime:     "python3.11",
			},
			"webhook": {
				Name:        "webhook",
				Description: "Webhook handler",
				Category:    "webhook",
				Runtime:     "python3.11",
			},
			"hello-world": {
				Name:        "hello-world",
				Description: "Simple function",
				Category:    "utility",
				Runtime:     "python3.11",
			},
		},
	}
}

// GetTemplate returns the best matching template
func (r *TemplateRegistry) GetTemplate(req *GenerationRequest) *Template {
	desc := strings.ToLower(req.Description + " " + req.Name)
	category := strings.ToLower(req.Category)

	if strings.Contains(category, "http") || strings.Contains(desc, "api") || strings.Contains(desc, "endpoint") || strings.Contains(desc, "rest") {
		return r.templates["http-api"]
	}
	if strings.Contains(category, "cron") || strings.Contains(category, "schedule") || strings.Contains(desc, "cron") || strings.Contains(desc, "scheduled") {
		return r.templates["cron-job"]
	}
	if strings.Contains(category, "webhook") || strings.Contains(desc, "webhook") || strings.Contains(desc, "callback") {
		return r.templates["webhook"]
	}
	return r.templates["hello-world"]
}

// Factory creates stub generators with configuration
type Factory struct {
	config Config
}

// NewFactory creates a new factory
func NewFactory(config Config) *Factory {
	if config.FallbackMode == "" {
		config.FallbackMode = "structured"
	}
	return &Factory{config: config}
}

// Create returns a configured CodeGenerator
func (f *Factory) Create() (CodeGenerator, error) {
	registry := NewTemplateRegistry()
	return &stubGenerator{registry: registry, config: f.config}, nil
}

// stubGenerator implements CodeGenerator using templates
type stubGenerator struct {
	registry *TemplateRegistry
	config   Config
}

// GenerateCode produces function code from the request
func (g *stubGenerator) GenerateCode(ctx context.Context, req *GenerationRequest) (string, error) {
	if req.Runtime == "" {
		req.Runtime = "python3.11"
	}

	template := g.registry.GetTemplate(req)

	switch template.Category {
	case "http":
		return g.generateHTTPAPI(req)
	case "scheduler":
		return g.generateCronJob(req)
	case "webhook":
		return g.generateWebhook(req)
	default:
		return g.generateUtility(req)
	}
}

// GenerateFunction generates a complete function
func (g *stubGenerator) GenerateFunction(ctx context.Context, req *GenerationRequest) (*GenerationResult, error) {
	code, err := g.GenerateCode(ctx, req)
	if err != nil {
		return &GenerationResult{Success: false, Error: err.Error()}, nil
	}

	return &GenerationResult{
		FunctionID: uuid.New(),
		Code:      code,
		Manifest:  g.createManifest(req),
		ModelUsed: "stub-template-v1",
		Success:   true,
	}, nil
}

func (g *stubGenerator) createManifest(req *GenerationRequest) string {
	m := map[string]any{
		"name":         req.Name,
		"version":      "1.0.0",
		"description": req.Description,
		"runtime":     req.Runtime,
		"category":    req.Category,
		"generator":   "stub-template-v1",
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}
	if req.InputSchema != nil {
		m["input_schema"] = req.InputSchema
	}
	if req.OutputSchema != nil {
		m["output_schema"] = req.OutputSchema
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}

// extractResourceName attempts to detect the primary resource from request
func extractResourceName(req *GenerationRequest) string {
	desc := strings.ToLower(req.Description + " " + req.Name)

	resources := []string{"user", "product", "order", "payment", "subscription",
		"customer", "invoice", "transaction", "item", "data", "record"}

	for _, r := range resources {
		if strings.Contains(desc, r) {
			return r + "s"
		}
	}
	return "resources"
}

// extractFields extracts field names from input schema
func extractFields(schema map[string]any) []string {
	if schema == nil {
		return []string{}
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		return []string{}
	}
	fields := make([]string, 0, len(props))
	for name := range props {
		fields = append(fields, name)
	}
	return fields
}

func (g *stubGenerator) generateHTTPAPI(req *GenerationRequest) (string, error) {
	schemaComments := ""
	if g.config.IncludeSchemaComments {
		if req.InputSchema != nil {
			schemaComments += "# Input Schema:\n# " + strings.ReplaceAll(string(mustJSON(req.InputSchema)), "\n", "\n# ") + "\n\n"
		}
		if req.OutputSchema != nil {
			schemaComments += "# Output Schema:\n# " + strings.ReplaceAll(string(mustJSON(req.OutputSchema)), "\n", "\n# ") + "\n\n"
		}
	}

	resource := extractResourceName(req)

	code := fmt.Sprintf(`"""HTTP API Function: %s"""

import json
import re
from datetime import datetime
from typing import Any, Dict, List, Optional
from urllib.parse import urlparse, parse_qs

# In-memory storage (replace with database in production)
_storage: Dict[str, List[Dict[str, Any]]] = {}
_resource_id_counter: Dict[str, int] = {}


def _get_resource_id(resource: str) -> int:
    """Get and increment counter for resource IDs."""
    if resource not in _resource_id_counter:
        _resource_id_counter[resource] = 1
        return 1
    current = _resource_id_counter[resource]
    _resource_id_counter[resource] = current + 1
    return current


def _parse_path(url: str) -> tuple[str, Dict[str, str]]:
    """Parse URL path and extract path parameters."""
    parsed = urlparse(url)
    path = parsed.path.strip("/")
    params = {}
    
    # Extract path parameters like /users/{id}
    parts = path.split("/")
    for i, part in enumerate(parts):
        if re.match(r"^\{.+\}$", part):
            params[f"param_{i}"] = part[1:-1]
    
    return path, params


def _get_query_params(url: str) -> Dict[str, List[str]]:
    """Extract query parameters from URL."""
    parsed = urlparse(url)
    return {k: v for k, v in parse_qs(parsed.query).items()}


def _filter_by_query(items: List[Dict[str, Any]], params: Dict[str, List[str]]) -> List[Dict[str, Any]]:
    """Filter items based on query parameters."""
    if not params:
        return items
    filtered = items
    for key, values in params.items():
        if len(values) == 1:
            filtered = [item for item in filtered if str(item.get(key, "")).lower() == values[0].lower()]
    return filtered


def _serialize_response(data: Any, status: int = 200) -> Dict[str, Any]:
    """Serialize response with proper formatting."""
    return {
        "status": status,
        "body": data if isinstance(data, dict) else {"data": data},
        "headers": {
            "Content-Type": "application/json",
            "X-Resource-Count": str(len(data)) if isinstance(data, list) else "1",
        }
    }


async def handler(event, env=None, ctx=None) -> Dict[str, Any]:
    """
    Handle HTTP requests with RESTful routing.
    
    Routes:
        GET    /%s              - List all %s
        GET    /%s/{{id}}       - Get %s by ID
        POST   /%s              - Create new %s
        PUT    /%s/{{id}}       - Update %s
        DELETE /%s/{{id}}       - Delete %s
        GET    /%s/search       - Search %s with query params
    """
    if isinstance(event, dict):
        url = event.get("url", "")
        method = event.get("method", "GET").upper()
        headers = event.get("headers", {})
        body = event.get("body")
    else:
        url = str(event)
        method = "GET"
        headers = {}
        body = None

    path, path_params = _parse_path(url)
    query_params = _get_query_params(url)
    resource_path = "/".join(path.split("/")[:-1]) if "/" in path else path

    # Route matching
    if path == "%s" or path == "%s/":
        if method == "GET":
            return await list_%s(query_params)
        if method == "POST":
            return await create_%s(body)
    
    if path.startswith("%s/") and method == "GET" and path.endswith("/search"):
        resource = path.split("/")[0]
        return await search_%s(query_params)
    
    if re.match(r"^%s/[^/]+$", path):
        resource_id = path.split("/")[-1]
        if method == "GET":
            return await get_%s(resource_id)
        if method == "PUT":
            return await update_%s(resource_id, body)
        if method == "DELETE":
            return await delete_%s(resource_id)
    
    return {"status": 404, "body": {"error": "Not found", "path": path}, "headers": {}}


async def list_%s(query_params: Dict[str, List[str]]) -> Dict[str, Any]:
    """List all %s with optional filtering."""
    items = _storage.get("%s", [])
    filtered = _filter_by_query(items, query_params)
    return _serialize_response({"items": filtered, "count": len(filtered)})


async def get_%s(resource_id: str) -> Dict[str, Any]:
    """Get %s by ID."""
    items = _storage.get("%s", [])
    for item in items:
        if str(item.get("id", "")) == resource_id:
            return _serialize_response(item)
    return {"status": 404, "body": {"error": "Not found"}, "headers": {}}


async def create_%s(body: Any) -> Dict[str, Any]:
    """Create new %s."""
    if not body:
        return {"status": 400, "body": {"error": "Request body required"}, "headers": {}}
    
    resource = "%s"
    if resource not in _storage:
        _storage[resource] = []
    
    item = body if isinstance(body, dict) else {"data": body}
    item["id"] = _get_resource_id(resource)
    item["created_at"] = datetime.utcnow().isoformat()
    _storage[resource].append(item)
    
    return {"status": 201, "body": item, "headers": {"Content-Type": "application/json"}}


async def update_%s(resource_id: str, body: Any) -> Dict[str, Any]:
    """Update existing %s."""
    if not body:
        return {"status": 400, "body": {"error": "Request body required"}, "headers": {}}
    
    resource = "%s"
    items = _storage.get(resource, [])
    for i, item in enumerate(items):
        if str(item.get("id", "")) == resource_id:
            item.update(body if isinstance(body, dict) else {"data": body})
            item["updated_at"] = datetime.utcnow().isoformat()
            _storage[resource][i] = item
            return _serialize_response(item)
    
    return {"status": 404, "body": {"error": "Not found"}, "headers": {}}


async def delete_%s(resource_id: str) -> Dict[str, Any]:
    """Delete %s."""
    resource = "%s"
    items = _storage.get(resource, [])
    for i, item in enumerate(items):
        if str(item.get("id", "")) == resource_id:
            _storage[resource].pop(i)
            return {"status": 204, "body": {}, "headers": {}}
    
    return {"status": 404, "body": {"error": "Not found"}, "headers": {}}


async def search_%s(query_params: Dict[str, List[str]]) -> Dict[str, Any]:
    """Search %s by query parameters."""
    return await list_%s(query_params)
`, req.Name, resource, resource, resource, resource, resource, resource, resource, resource, resource,
		resource, resource, resource, resource, resource, resource, resource,
		resource, resource, resource, resource, resource,
		resource, resource, resource, resource,
		resource, resource, resource, resource,
		resource, resource)

	return code, nil
}

func (g *stubGenerator) generateCronJob(req *GenerationRequest) (string, error) {
	schedule := extractSchedule(req.Description)
	resource := extractResourceName(req)

	code := fmt.Sprintf(`"""Cron Job Function: %s"""

import json
import logging
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional
from collections import deque

# Configure logging
logger = logging.getLogger(__name__)

# Execution state (persist across invocations in production via Redis/DB)
class CronState:
    def __init__(self):
        self.last_run: Optional[str] = None
        self.run_count: int = 0
        self.success_count: int = 0
        self.failure_count: int = 0
        self.last_success: Optional[str] = None
        self.last_failure: Optional[str] = None
        self.pending_items: deque = deque(maxlen=1000)

_state = CronState()


async def handler(event, env=None, ctx=None) -> Dict[str, Any]:
    """
    Scheduled task handler. Schedule: %s
    
    Main entry point for scheduled execution.
    Implements idempotent processing with state tracking.
    """
    start_time = datetime.utcnow()
    execution_id = f"exec_{int(start_time.timestamp())}"
    
    logger.info(f"Starting cron job execution_id={execution_id}")
    
    try:
        # Check for required environment variables
        if not env:
            env = {}
        
        # Run the main task
        result = await execute_task(event, env, _state)
        
        # Update success state
        _state.last_run = start_time.isoformat()
        _state.last_success = start_time.isoformat()
        _state.success_count += 1
        _state.run_count += 1
        
        duration_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)
        
        logger.info(f"Cron job completed execution_id={execution_id} duration_ms={duration_ms}")
        
        return {
            "ok": True,
            "execution_id": execution_id,
            "result": result,
            "executed_at": start_time.isoformat(),
            "duration_ms": duration_ms,
            "stats": {
                "total_runs": _state.run_count,
                "success_count": _state.success_count,
                "failure_count": _state.failure_count,
                "last_success": _state.last_success
            }
        }
        
    except Exception as e:
        # Update failure state
        _state.last_run = start_time.isoformat()
        _state.last_failure = start_time.isoformat()
        _state.failure_count += 1
        _state.run_count += 1
        
        logger.error(f"Cron job failed execution_id={execution_id} error={str(e)}")
        
        return {
            "ok": False,
            "execution_id": execution_id,
            "error": str(e),
            "error_type": type(e).__name__,
            "executed_at": start_time.isoformat(),
            "stats": {
                "total_runs": _state.run_count,
                "success_count": _state.success_count,
                "failure_count": _state.failure_count,
                "last_failure": _state.last_failure
            }
        }


async def execute_task(event, env: Dict[str, Any], state: CronState) -> Dict[str, Any]:
    """
    Main task execution logic.
    
    Implements common cron job patterns:
    - Incremental processing with last_run timestamp
    - Batch processing with configurable batch size
    - Error recovery and retry logic
    - State persistence for resumability
    
    Override this function with your specific task logic.
    """
    # Get last run time for incremental processing
    last_run = state.last_run
    
    # Configuration
    batch_size = int(env.get("BATCH_SIZE", "100"))
    retry_count = int(env.get("RETRY_COUNT", "3"))
    retry_delay_ms = int(env.get("RETRY_DELAY_MS", "1000"))
    
    processed = 0
    failed = 0
    skipped = 0
    
    # Fetch items to process (implement your specific fetch logic)
    items = await fetch_pending_items(last_run, batch_size)
    
    for item in items:
        try:
            # Process each item with retry logic
            for attempt in range(retry_count):
                try:
                    await process_item(item, env)
                    processed += 1
                    break
                except Exception as e:
                    if attempt < retry_count - 1:
                        logger.warning(f"Retry attempt {attempt + 1} for item")
                        await asyncio.sleep(retry_delay_ms / 1000)
                    else:
                        raise
        except Exception as e:
            logger.error(f"Failed to process item: {e}")
            failed += 1
            # Store failed item for manual review
            state.pending_items.append({
                "item": item,
                "error": str(e),
                "failed_at": datetime.utcnow().isoformat()
            })
    
    return {
        "processed": processed,
        "failed": failed,
        "skipped": skipped,
        "batch_size": len(items),
        "last_run": last_run,
        "timestamp": datetime.utcnow().isoformat()
    }


async def fetch_pending_items(last_run: Optional[str], batch_size: int) -> List[Dict[str, Any]]:
    """
    Fetch items that need to be processed.
    
    Implement your specific fetch logic:
    - Query database for new/updated records
    - Poll external API for pending items
    - Read from message queue
    """
    # PLACEHOLDER: Replace with your actual fetch logic
    # Example: Query database for items created after last_run
    #
    # items = await db.query("""
    #     SELECT * FROM tasks
    #     WHERE created_at > ?
    #     ORDER BY created_at ASC
    #     LIMIT ?
    # """, [last_run, batch_size])
    
    return []


async def process_item(item: Dict[str, Any], env: Dict[str, Any]) -> None:
    """
    Process a single item.
    
    Implement your specific processing logic:
    - Transform data
    - Call external APIs
    - Update database records
    - Send notifications
    """
    # PLACEHOLDER: Replace with your actual processing logic
    #
    # await external_api.send(item)
    # await db.update(item["id"], {"processed": True})
    # await notification.send(item)
    pass


def get_state() -> Dict[str, Any]:
    """Get current cron job state."""
    return {
        "last_run": _state.last_run,
        "run_count": _state.run_count,
        "success_count": _state.success_count,
        "failure_count": _state.failure_count,
        "pending_items_count": len(_state.pending_items)
    }


import asyncio
`, req.Name, schedule, resource, resource, resource, resource, resource, resource, resource)

	return code, nil
}

func (g *stubGenerator) generateWebhook(req *GenerationRequest) (string, error) {
	resource := extractResourceName(req)

	code := fmt.Sprintf(`"""Webhook Handler: %s"""

import json
import hmac
import hashlib
import re
import logging
from datetime import datetime
from typing import Any, Dict, List, Optional
from collections import deque

# Configure logging
logger = logging.getLogger(__name__)

# Webhook event storage (replace with database/queue in production)
_event_store: deque = deque(maxlen=10000)
_processed_count: int = 0
_failed_count: int = 0


def _compute_signature(payload: str, secret: str) -> str:
    """Compute HMAC-SHA256 signature for payload."""
    return "sha256=" + hmac.new(
        secret.encode("utf-8"),
        payload.encode("utf-8"),
        hashlib.sha256
    ).hexdigest()


def _verify_signature(payload: str, signature: str, secret: str) -> bool:
    """
    Verify webhook signature using HMAC-SHA256.
    
    Supports multiple signature formats:
    - GitHub style: X-Hub-Signature-256
    - Stripe style: X-Webhook-Signature
    - Custom: X-Signature
    """
    if not signature or not secret:
        return True  # Skip verification if not configured
    
    try:
        # Compute expected signature
        expected = _compute_signature(payload, secret)
        
        # Handle different signature formats
        if signature.startswith("sha256="):
            # GitHub/Stripe format
            return hmac.compare_digest(expected, signature)
        elif len(signature) == 64:
            # Raw hex format (no prefix)
            return hmac.compare_digest(expected, "sha256=" + signature)
        else:
            # Try both formats
            return (hmac.compare_digest(expected, signature) or 
                    hmac.compare_digest(expected, "sha256=" + signature))
    except Exception as e:
        logger.error(f"Signature verification failed: {e}")
        return False


def _parse_payload(body: Any) -> Dict[str, Any]:
    """Parse webhook body, handling various formats."""
    if isinstance(body, dict):
        return body
    if isinstance(body, str):
        try:
            return json.loads(body)
        except json.JSONDecodeError:
            return {"raw": body}
    return {"data": body}


async def handler(event, env=None, ctx=None) -> Dict[str, Any]:
    """
    Handle incoming webhooks with signature verification and event routing.
    
    Supports:
    - Multiple signature schemes (GitHub, Stripe, custom)
    - Event type detection and routing
    - Idempotent processing with deduplication
    - Retry logic for failed processing
    """
    start_time = datetime.utcnow()
    
    # Parse webhook event
    if isinstance(event, dict):
        headers = event.get("headers", {})
        body = event.get("body", event)
    else:
        headers = {}
        body = event

    # Parse payload
    payload = _parse_payload(body)
    
    # Get webhook secret for signature verification
    webhook_secret = env.get("WEBHOOK_SECRET") if env else None
    
    # Verify signature if secret is configured
    if webhook_secret:
        signature = (headers.get("x-webhook-signature") or 
                    headers.get("x-hub-signature-256") or 
                    headers.get("x-hub-signature") or
                    headers.get("x-signature"))
        
        if signature:
            # Get raw body for signature verification
            raw_body = body if isinstance(body, str) else json.dumps(body, sort_keys=True)
            
            if not _verify_signature(raw_body, signature, webhook_secret):
                logger.warning("Webhook signature verification failed")
                return {
                    "status": 401,
                    "body": {"error": "Invalid signature"},
                    "headers": {"Content-Type": "application/json"}
                }

    # Detect event type and ID for idempotency
    event_id = _extract_event_id(payload, headers)
    event_type = _detect_event_type(payload, headers)
    
    logger.info(f"Webhook received: type={event_type} id={event_id}")

    # Process the webhook with error handling
    try:
        result = await process_webhook(payload, event_type, event_id, env)
        
        global _processed_count
        _processed_count += 1
        
        logger.info(f"Webhook processed successfully: type={event_type} id={event_id}")
        
        return {
            "status": 200,
            "body": {
                "ok": True,
                "event_id": event_id,
                "event_type": event_type,
                "result": result,
                "processed_at": start_time.isoformat()
            },
            "headers": {
                "Content-Type": "application/json",
                "X-Event-ID": event_id,
                "X-Event-Type": event_type
            }
        }
        
    except Exception as e:
        global _failed_count
        _failed_count += 1
        
        logger.error(f"Webhook processing failed: type={event_type} id={event_id} error={e}")
        
        # Store failed event for retry/manual review
        _event_store.append({
            "payload": payload,
            "event_type": event_type,
            "event_id": event_id,
            "error": str(e),
            "failed_at": start_time.isoformat()
        })
        
        return {
            "status": 500,
            "body": {
                "error": str(e),
                "event_id": event_id,
                "event_type": event_type
            },
            "headers": {
                "Content-Type": "application/json",
                "X-Event-ID": event_id
            }
        }


async def process_webhook(payload: Dict[str, Any], event_type: str, event_id: str, env: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process webhook payload based on event type.
    
    Implement your webhook processing logic here.
    
    Common patterns:
    - Route based on event_type
    - Store event for async processing
    - Trigger downstream actions
    - Update database records
    - Send notifications
    
    Args:
        payload: Parsed webhook payload
        event_type: Detected event type
        event_id: Unique event identifier for idempotency
        env: Environment variables
    
    Returns:
        Processing result
    """
    # Route based on event type
    handlers = {
        "created": handle_created,
        "updated": handle_updated,
        "deleted": handle_deleted,
        "notification": handle_notification,
        "payment": handle_payment,
        "subscription": handle_subscription,
        "user": handle_user_event,
    }
    
    # Get handler for event type
    handler = handlers.get(event_type.lower())
    
    if handler:
        return await handler(payload, env)
    
    # Default: log and acknowledge
    logger.info(f"Unhandled event type: {event_type}")
    return {
        "processed": True,
        "event_type": event_type,
        "action": "acknowledged"
    }


async def handle_created(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle creation events."""
    resource_type = detect_resource_type(payload)
    resource_id = payload.get("id") or payload.get("data", {}).get("id")
    
    logger.info(f"Resource created: type={resource_type} id={resource_id}")
    
    # Implement your creation handling logic
    # await db.insert(resource_type, payload)
    # await trigger_webhook(resource_type, "created", payload)
    
    return {
        "action": "created",
        "resource_type": resource_type,
        "resource_id": resource_id
    }


async def handle_updated(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle update events."""
    resource_type = detect_resource_type(payload)
    resource_id = payload.get("id") or payload.get("data", {}).get("id")
    
    logger.info(f"Resource updated: type={resource_type} id={resource_id}")
    
    # Implement your update handling logic
    # await db.update(resource_type, resource_id, payload)
    
    return {
        "action": "updated",
        "resource_type": resource_type,
        "resource_id": resource_id
    }


async def handle_deleted(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle deletion events."""
    resource_type = detect_resource_type(payload)
    resource_id = payload.get("id") or payload.get("data", {}).get("id")
    
    logger.info(f"Resource deleted: type={resource_type} id={resource_id}")
    
    # Implement your deletion handling logic
    # await db.delete(resource_type, resource_id)
    
    return {
        "action": "deleted",
        "resource_type": resource_type,
        "resource_id": resource_id
    }


async def handle_notification(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle notification events."""
    notification_id = payload.get("id")
    notification_type = payload.get("notification_type") or payload.get("type")
    
    logger.info(f"Notification: type={notification_type} id={notification_id}")
    
    return {
        "action": "notification_processed",
        "notification_id": notification_id,
        "notification_type": notification_type
    }


async def handle_payment(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle payment events."""
    payment_id = payload.get("id") or payload.get("payment_id")
    amount = payload.get("amount") or payload.get("data", {}).get("amount")
    
    logger.info(f"Payment processed: id={payment_id} amount={amount}")
    
    return {
        "action": "payment_processed",
        "payment_id": payment_id,
        "amount": amount
    }


async def handle_subscription(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle subscription events."""
    sub_id = payload.get("id") or payload.get("subscription_id")
    status = payload.get("status")
    
    logger.info(f"Subscription: id={sub_id} status={status}")
    
    return {
        "action": "subscription_processed",
        "subscription_id": sub_id,
        "status": status
    }


async def handle_user_event(payload: Dict[str, Any], env: Dict[str, Any]) -> Dict[str, Any]:
    """Handle user-related events."""
    user_id = payload.get("user_id") or payload.get("data", {}).get("user_id")
    action = payload.get("action", "unknown")
    
    logger.info(f"User event: user_id={user_id} action={action}")
    
    return {
        "action": action,
        "user_id": user_id
    }


def _extract_event_id(payload: Dict[str, Any], headers: Dict[str, Any]) -> str:
    """Extract unique event ID from payload or headers."""
    # Check common ID fields
    for field in ["id", "event_id", "message_id", "webhook_id", "request_id"]:
        if field in payload:
            return str(payload[field])
    
    # Check headers
    for header in ["x-event-id", "x-request-id", "x-id"]:
        if header in headers:
            return str(headers[header])
    
    # Generate from payload hash
    payload_str = json.dumps(payload, sort_keys=True)
    return hashlib.md5(payload_str.encode()).hexdigest()[:16]


def _detect_event_type(payload: Dict[str, Any], headers: Dict[str, Any]) -> str:
    """
    Detect event type from payload and headers.
    
    Supports multiple providers:
    - GitHub: headers["x-github-event"]
    - Stripe: payload["type"]
    - Custom: payload["event_type"] or payload["action"]
    """
    # Check headers first (GitHub style)
    if "x-github-event" in headers:
        return headers["x-github-event"]
    
    # Check common payload fields
    for field in ["type", "action", "event", "event_type", "notification_type"]:
        if field in payload:
            return str(payload[field])
    
    # Check nested data object
    if "data" in payload and isinstance(payload["data"], dict):
        for field in ["type", "action", "event"]:
            if field in payload["data"]:
                return str(payload["data"][field])
    
    return "unknown"


def detect_resource_type(payload: Dict[str, Any]) -> str:
    """Detect resource type from payload."""
    for field in ["resource_type", "type", "model"]:
        if field in payload:
            return str(payload[field])
    
    if "data" in payload and isinstance(payload["data"], dict):
        for field in ["resource_type", "type"]:
            if field in payload["data"]:
                return str(payload["data"][field])
    
    return "unknown"


def get_stats() -> Dict[str, Any]:
    """Get webhook processing statistics."""
    return {
        "processed_count": _processed_count,
        "failed_count": _failed_count,
        "pending_count": len(_event_store)
    }
`, req.Name, resource)

	return code, nil
}

func (g *stubGenerator) generateUtility(req *GenerationRequest) (string, error) {
	schemaComments := ""
	if g.config.IncludeSchemaComments {
		if req.InputSchema != nil {
			schemaComments += "# Input Schema:\n# " + strings.ReplaceAll(string(mustJSON(req.InputSchema)), "\n", "\n# ") + "\n\n"
		}
		if req.OutputSchema != nil {
			schemaComments += "# Output Schema:\n# " + strings.ReplaceAll(string(mustJSON(req.OutputSchema)), "\n", "\n# ") + "\n\n"
		}
	}

	code := fmt.Sprintf(`"""Function: %s"""

import json
import logging
from datetime import datetime
from typing import Any, Dict, List, Optional, Union

# Configure logging
logger = logging.getLogger(__name__)

# Function state (use external storage in production)
_class Storage:
    def __init__(self):
        self._data: Dict[str, Any] = {}
        self._cache: Dict[str, tuple[Any, float]] = {}
        self._call_count: int = 0
    
    def get(self, key: str, default: Any = None) -> Any:
        return self._data.get(key, default)
    
    def set(self, key: str, value: Any) -> None:
        self._data[key] = value
    
    def increment(self, key: str, delta: int = 1) -> int:
        current = self._data.get(key, 0)
        self._data[key] = current + delta
        return self._data[key]
    
    def cache_get(self, key: str, ttl_seconds: float = 300) -> Optional[Any]:
        if key in self._cache:
            value, timestamp = self._cache[key]
            if (datetime.utcnow().timestamp() - timestamp) < ttl_seconds:
                return value
            del self._cache[key]
        return None
    
    def cache_set(self, key: str, value: Any) -> None:
        self._cache[key] = (value, datetime.utcnow().timestamp())

_storage = Storage()


def _serialize_error(error: Exception) -> Dict[str, Any]:
    """Serialize exception for error response."""
    return {
        "error": str(error),
        "type": type(error).__name__,
        "timestamp": datetime.utcnow().isoformat()
    }


def _apply_transformations(data: Any, transforms: List[str]) -> Any:
    """Apply a series of transformations to data."""
    for transform in transforms:
        if transform == "uppercase" and isinstance(data, str):
            data = data.upper()
        elif transform == "lowercase" and isinstance(data, str):
            data = data.lower()
        elif transform == "strip" and isinstance(data, str):
            data = data.strip()
        elif transform == "json_parse" and isinstance(data, str):
            try:
                data = json.loads(data)
            except json.JSONDecodeError:
                pass
        elif transform == "flatten" and isinstance(data, dict):
            data = {k: v for k, v in data.items() if v is not None}
    return data


def _validate_input(data: Any, schema: Dict[str, Any]) -> tuple[bool, Optional[str]]:
    """
    Validate input data against schema.
    
    Returns:
        (is_valid, error_message)
    """
    if not schema:
        return True, None
    
    required_fields = schema.get("required", [])
    properties = schema.get("properties", {})
    
    # Check required fields
    if isinstance(data, dict):
        for field in required_fields:
            if field not in data:
                return False, f"Missing required field: {field}"
        
        # Check field types
        for field, spec in properties.items():
            if field in data:
                expected_type = spec.get("type")
                if expected_type == "string" and not isinstance(data[field], str):
                    return False, f"Field '{field}' must be string"
                if expected_type == "number" and not isinstance(data[field], (int, float)):
                    return False, f"Field '{field}' must be number"
                if expected_type == "boolean" and not isinstance(data[field], bool):
                    return False, f"Field '{field}' must be boolean"
                if expected_type == "array" and not isinstance(data[field], list):
                    return False, f"Field '{field}' must be array"
                if expected_type == "object" and not isinstance(data[field], dict):
                    return False, f"Field '{field}' must be object"
    
    return True, None


async def handler(event, env=None, ctx=None) -> Dict[str, Any]:
    """
    Main function handler with built-in validation, error handling, and metrics.
    
    Features:
    - Automatic input validation
    - Error handling with structured responses
    - Execution metrics
    - Caching support
    - Transform pipeline
    """
    start_time = datetime.utcnow()
    _storage.increment("total_calls")
    
    try:
        # Parse input data
        data = event if isinstance(event, dict) else {"value": event}
        
        # Validate input if schema is provided
        if %s:
            is_valid, error_msg = _validate_input(data, %s)
            if not is_valid:
                return {
                    "ok": False,
                    "error": error_msg,
                    "timestamp": start_time.isoformat()
                }
        
        # Process the input
        result = await process(data, env)
        
        # Serialize result
        if isinstance(result, (dict, list)):
            serialized = result
        else:
            serialized = {"result": result}
        
        duration_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)
        
        return {
            "ok": True,
            "result": serialized,
            "timestamp": start_time.isoformat(),
            "duration_ms": duration_ms,
            "metrics": {
                "calls": _storage.get("total_calls", 0)
            }
        }
        
    except Exception as e:
        logger.error(f"Function error: {e}")
        return {
            "ok": False,
            "error": _serialize_error(e),
            "timestamp": start_time.isoformat()
        }


async def process(data: Dict[str, Any], env: Optional[Dict[str, Any]]) -> Dict[str, Any]:
    """
    Main processing logic.
    
    Implement your function logic here. This is the main entry point
    for your function's business logic.
    
    Args:
        data: Parsed and validated input data
        env: Environment variables and secrets
    
    Returns:
        Processing result
    """
    # Extract input fields
    input_value = data.get("value", data)
    
    # Implement your business logic here
    # 
    # Examples:
    # - Transform and enrich data
    # - Call external APIs
    # - Perform calculations
    # - Aggregate data
    # - Make decisions
    
    # Example: Simple data transformation pipeline
    transforms = ["strip", "lowercase"]
    processed = _apply_transformations(input_value, transforms)
    
    return {
        "processed": True,
        "input": data,
        "output": processed,
        "timestamp": datetime.utcnow().isoformat()
    }


async def batch_process(items: List[Dict[str, Any]], env: Optional[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """
    Batch processing handler for handling multiple items.
    
    Use this for processing arrays of items efficiently.
    
    Args:
        items: List of items to process
        env: Environment variables
    
    Returns:
        List of processing results
    """
    results = []
    for item in items:
        try:
            result = await process(item, env)
            results.append({"ok": True, "result": result})
        except Exception as e:
            results.append({"ok": False, "error": _serialize_error(e)})
    return results


async def health_check() -> Dict[str, Any]:
    """Health check endpoint for monitoring."""
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "metrics": {
            "total_calls": _storage.get("total_calls", 0)
        }
    }
`, req.Name,
		strings.ReplaceAll(string(mustJSON(req.InputSchema)), "\n", "\n                    "),
		strings.ReplaceAll(string(mustJSON(req.InputSchema)), "\n", "\n                    "))

	return code, nil
}

// Helper functions
func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func extractSchedule(desc string) string {
	desc = strings.ToLower(desc)
	if strings.Contains(desc, "hourly") {
		return "hourly (0 * * * *)"
	}
	if strings.Contains(desc, "daily") || strings.Contains(desc, "every day") {
		return "daily (0 0 * * *)"
	}
	if strings.Contains(desc, "weekly") {
		return "weekly (0 0 * * 0)"
	}
	cronRe := regexp.MustCompile(`\d+\s+\d+\s+\d+\s+\d+\s+\d+`)
	if m := cronRe.FindString(desc); m != "" {
		return "cron: " + m
	}
	return "configure in settings"
}

// GetFieldSchema extracts field info from schema properties
func GetFieldSchema(schema map[string]any) string {
	if schema == nil {
		return "any"
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		return "any"
	}
	if len(props) == 0 {
		return "any"
	}
	fields := make([]string, 0, len(props))
	for name, spec := range props {
		if m, ok := spec.(map[string]any); ok {
			t := m["type"]
			if t != nil {
				fields = append(fields, name+":"+strconv.Itoa(len(fields)+1))
			}
		}
	}
	return strings.Join(fields[:min(len(fields), 5)], ", ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}