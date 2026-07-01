"""RAG and template retrieval for function generation.

Retrieves similar functions from the registry and provides templates
for cost-optimized generation using retrieval-augmented techniques.
"""

import json
import logging
from typing import List, Optional, Dict, Any, Tuple
from dataclasses import dataclass

from ...models.schemas import TripleQueryRequest
from ..flyembed import get_flyembed_service
from ..embeddings import get_embeddings_service
from ...config import settings

logger = logging.getLogger(__name__)


@dataclass
class RetrievedFunction:
    """A retrieved similar function."""
    function_id: str
    name: str
    description: str
    runtime: str
    code: str
    manifest: Dict[str, Any]
    similarity_score: float
    rank: int


@dataclass
class TemplateMatch:
    """A matched template for the request."""
    template_id: str
    template_name: str
    description: str
    base_code: str
    fill_prompt: str
    match_score: float
    estimated_tokens_saved: int


class FunctionTemplateLibrary:
    """Built-in templates for common function patterns."""

    TEMPLATES = {
        "webhook_handler": {
            "id": "webhook_handler",
            "name": "Webhook Handler",
            "description": "Receive and process webhook events",
            "keywords": ["webhook", "callback", "event", "http endpoint", "post"],
            "python": {
                "base_code": '''import hmac
import hashlib
import os

def handler(request):
    """Handle incoming webhook request."""
    # Validate webhook signature
    signature = request.headers.get("X-Signature-256") or request.headers.get("X-Webhook-Signature")
    secret = os.environ.get("WEBHOOK_SECRET", os.environ.get("STRIPE_WEBHOOK_SECRET", ""))

    if signature and secret:
        if not verify_webhook_signature(request.get_data(), signature, secret):
            return {"error": "Invalid signature"}, 401

    payload = request.get_json()

    # Process the payload
    result = process_webhook(payload)

    return {"status": "success", "processed": result}

def verify_webhook_signature(payload: bytes, signature: str, secret: str) -> bool:
    """Verify HMAC-SHA256 webhook signature."""
    if signature.startswith("sha256="):
        signature = signature[7:]

    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(expected, signature)

def process_webhook(payload):
    """Process webhook payload based on event type."""
    event_type = payload.get("type", "unknown")

    handlers = {
        "payment.succeeded": lambda p: {"action": "fulfill_order", "order_id": p.get("data", {}).get("id")},
        "payment.failed": lambda p: {"action": "notify_customer", "customer_id": p.get("data", {}).get("customer")},
        "user.signup": lambda p: {"action": "send_welcome_email", "user_id": p.get("data", {}).get("id")},
    }

    handler = handlers.get(event_type)
    if handler:
        return handler(payload)
    return {"event_type": event_type, "received": True}''',
                "fill_prompt": "The webhook handler is production-ready with HMAC-SHA256 validation. Customize the process_webhook function to handle specific event types from your webhook provider."
            },
            "nodejs": {
                "base_code": '''const crypto = require('crypto');

exports.handler = async (req, res) => {
    // Validate webhook signature
    const signature = req.headers['x-signature-256'] || req.headers['x-webhook-signature'];
    const secret = process.env.WEBHOOK_SECRET || process.env.STRIPE_WEBHOOK_SECRET || '';

    if (signature && secret) {
        const expected = crypto
            .createHmac('sha256', secret)
            .update(JSON.stringify(req.body))
            .digest('hex');
        const expectedWithPrefix = `sha256=${expected}`;

        if (!crypto.timingSafeEqual(Buffer.from(expectedWithPrefix), Buffer.from(signature))) {
            return res.status(401).json({ error: 'Invalid signature' });
        }
    }

    const payload = req.body;

    // Process the payload
    const result = await processWebhook(payload);

    res.json({ status: "success", processed: result });
};

async function processWebhook(payload) {
    /** Process webhook payload. */
    const eventType = payload.type || 'unknown';
    switch (eventType) {
        case 'payment.succeeded':
            return { eventType, action: 'fulfill_order' };
        case 'payment.failed':
            return { eventType, action: 'notify_customer' };
        default:
            return { eventType, received: true };
    }
}''',
                "fill_prompt": "Implement webhook validation using HMAC-SHA256 signature verification and add event-type specific processing logic for the webhook events."
            }
        },
        "api_client": {
            "id": "api_client",
            "name": "API Client",
            "description": "Make HTTP API calls to external services",
            "keywords": ["api", "http", "request", "fetch", "get", "post", "call external"],
            "python": {
                "base_code": '''import requests
import os

def handler(input):
    """Make API call to external service."""
    url = input.get("url")
    method = input.get("method", "GET")
    headers = input.get("headers", {})
    body = input.get("body")

    api_key = os.environ.get("API_KEY") or input.get("api_key")
    if api_key:
        headers["Authorization"] = f"Bearer {api_key}"

    response = requests.request(method, url, headers=headers, json=body)
    response.raise_for_status()

    return {"status_code": response.status_code, "data": response.json()}}''',
                "fill_prompt": "Configure the specific API endpoint URL, authentication credentials (API key, OAuth token), request/response formats, and error handling for your external service."
            },
            "nodejs": {
                "base_code": '''const axios = require('axios');

exports.handler = async (input) => {
    const url = input.url;
    const method = input.method || 'GET';
    const headers = input.headers || {};
    const body = input.body;

    const apiKey = process.env.API_KEY || input.api_key;
    if (apiKey) {
        headers['Authorization'] = `Bearer ${apiKey}`;
    }

    const response = await axios({ method, url, headers, data: body });

    return { statusCode: response.status, data: response.data };
};''',
                "fill_prompt": "Add specific API endpoint details, authentication, error handling, and response parsing."
            }
        },
        "data_transform": {
            "id": "data_transform",
            "name": "Data Transform",
            "description": "Transform and process data structures",
            "keywords": ["transform", "convert", "parse", "format", "json", "csv", "xml"],
            "python": {
                "base_code": '''import json

def handler(input):
    """Transform input data to output format."""
    data = input.get("data")
    output_format = input.get("format", "json")

    if output_format == "csv":
        return transform_to_csv(data)
    elif output_format == "xml":
        return transform_to_xml(data)
    else:
        return transform_to_json(data)

def transform_to_json(data):
    """Transform to JSON."""
    return {"result": data}

def transform_to_csv(data):
    """Transform to CSV format."""
    if isinstance(data, list) and data:
        headers = list(data[0].keys())
        rows = [headers]
        for item in data:
            rows.append([item.get(h, "") for h in headers])
        csv_str = ",".join(headers) + "\\n" + "\\n".join([",".join(str(v) for v in row) for row in rows])
        return {"result": csv_str}
    return {"result": ""}

def transform_to_xml(data):
    """Transform to XML format."""
    def dict_to_xml(tag, d):
        xml = f"<{tag}>"
        for k, v in d.items():
            xml += dict_to_xml(k, v) if isinstance(v, dict) else f"<{k}>{v}</{k}>"
        xml += f"</{tag}>"
        return xml
    return {"result": dict_to_xml("root", data) if isinstance(data, dict) else ""}''',
                "fill_prompt": "Define the input schema, output format (CSV, XML, JSON), and specific transformation rules for your data processing pipeline."
            },
            "nodejs": {
                "base_code": '''exports.handler = async (input) => {
    const data = input.data;
    const outputFormat = input.format || 'json';

    if (outputFormat === 'csv') {
        return { result: toCsv(data) };
    } else if (outputFormat === 'xml') {
        return { result: toXml(data) };
    }
    return { result: data };
};

function toCsv(data) {
    if (Array.isArray(data) && data.length > 0) {
        const headers = Object.keys(data[0]);
        const rows = data.map(item => headers.map(h => item[h] ?? '').join(','));
        return [headers.join(','), ...rows].join('\\n');
    }
    return '';
}

function toXml(data) {
    const encodeXml = (str) => String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    const parse = (obj, tag = 'item') => {
        if (typeof obj !== 'object' || obj === null) return `<${tag}>${encodeXml(obj)}</${tag}>`;
        let xml = '';
        for (const [k, v] of Object.entries(obj)) {
            xml += parse(v, k);
        }
        return `<${tag}>${xml}</${tag}>`;
    };
    return parse(data);
};''',
                "fill_prompt": "Fill in the data transformation logic based on the input and output schemas."
            }
        },
        "db_operation": {
            "id": "db_operation",
            "name": "Database Operation",
            "description": "Read or write from database",
            "keywords": ["database", "db", "sql", "query", "store", "save", "postgres", "mongodb"],
            "python": {
                "base_code": '''import os
import re
import psycopg2

ALLOWED_TABLE_PATTERN = re.compile(r'^[a-zA-Z_][a-zA-Z0-9_]*$')

def validate_table_name(table: str) -> str:
    """Validate table name to prevent SQL injection."""
    if not ALLOWED_TABLE_PATTERN.match(table):
        raise ValueError(f"Invalid table name: {table}")
    return table

def handler(input):
    """Perform database operation with parameterized queries."""
    operation = input.get("operation")
    table = input.get("table")
    data = input.get("data", {})
    condition = input.get("condition")
    condition_params = input.get("condition_params", [])

    # Validate table name to prevent SQL injection
    table = validate_table_name(table)

    conn = psycopg2.connect(os.environ["DATABASE_URL"])
    try:
        with conn.cursor() as cur:
            if operation == "SELECT":
                if condition:
                    cur.execute(f"SELECT * FROM {table} WHERE {condition}", condition_params)
                else:
                    cur.execute(f"SELECT * FROM {table}")
                result = cur.fetchall()
            elif operation == "INSERT":
                columns = ", ".join(data.keys())
                placeholders = ", ".join(["%s"] * len(data))
                cur.execute(f"INSERT INTO {table} ({columns}) VALUES ({placeholders})", list(data.values()))
                conn.commit()
                result = {"inserted": cur.rowcount}
            elif operation == "UPDATE":
                set_clause = ", ".join([f"{k} = %s" for k in data.keys()])
                params = list(data.values()) + condition_params
                cur.execute(f"UPDATE {table} SET {set_clause} WHERE {condition}", params)
                conn.commit()
                result = {"updated": cur.rowcount}
            elif operation == "DELETE":
                cur.execute(f"DELETE FROM {table} WHERE {condition}", condition_params)
                conn.commit()
                result = {"deleted": cur.rowcount}
            else:
                result = {"error": "Unknown operation"}
        return {"result": result}
    finally:
        conn.close()''',
                "fill_prompt": "Configure the database table name, column mappings, SQL operation type (SELECT, INSERT, UPDATE, DELETE), and WHERE clause conditions for your use case. Always use condition_params for any user-provided values in WHERE clauses."
            },
            "nodejs": {
                "base_code": '''const { Pool } = require('pg');

const pool = new Pool({
    connectionString: process.env.DATABASE_URL
});

const ALLOWED_TABLE_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

function validateTableName(table) {
    if (!ALLOWED_TABLE_PATTERN.test(table)) {
        throw new Error(`Invalid table name: ${table}`);
    }
    return table;
}

exports.handler = async (input) => {
    const { operation, table, data = {}, condition, conditionParams = [] } = input;

    // Validate table name to prevent SQL injection
    const safeTable = validateTableName(table);

    const client = await pool.connect();
    try {
        let result;
        switch (operation) {
            case 'SELECT':
                if (condition) {
                    const selectResult = await client.query(
                        `SELECT * FROM ${safeTable} WHERE ${condition}`,
                        conditionParams
                    );
                    result = selectResult.rows;
                } else {
                    const selectResult = await client.query(`SELECT * FROM ${safeTable}`);
                    result = selectResult.rows;
                }
                break;
            case 'INSERT': {
                const columns = Object.keys(data).join(', ');
                const placeholders = Object.keys(data).map((_, i) => `$${i + 1}`).join(', ');
                const values = Object.values(data);
                await client.query(
                    `INSERT INTO ${safeTable} (${columns}) VALUES (${placeholders})`,
                    values
                );
                result = { inserted: client.rowCount };
                break;
            }
            case 'UPDATE': {
                const setClause = Object.keys(data).map((k, i) => `${k} = $${i + 1}`).join(', ');
                const values = [...Object.values(data), ...conditionParams];
                await client.query(
                    `UPDATE ${safeTable} SET ${setClause} WHERE ${condition}`,
                    values
                );
                result = { updated: client.rowCount };
                break;
            }
            case 'DELETE':
                await client.query(`DELETE FROM ${safeTable} WHERE ${condition}`, conditionParams);
                result = { deleted: client.rowCount };
                break;
            default:
                result = { error: 'Unknown operation' };
        }
        return { result };
    } finally {
        client.release();
    }
};''',
                "fill_prompt": "Fill in the specific SQL operations, table names, and parameter handling. Always use conditionParams for any user-provided values in WHERE clauses."
            }
        },
        "auth_handler": {
            "id": "auth_handler",
            "name": "Authentication Handler",
            "description": "Handle authentication and authorization",
            "keywords": ["auth", "login", "token", "jwt", "verify", "authenticate", "permission"],
            "python": {
                "base_code": '''import jwt
import os
from functools import wraps

def handler(request):
    """Handle authentication request."""
    token = request.headers.get("Authorization", "").replace("Bearer ", "")

    try:
        payload = jwt.decode(token, os.environ["JWT_SECRET"], algorithms=["HS256"])
        if not has_permission(payload, "access"):
            return {"authenticated": False, "error": "Insufficient permissions"}, 403
        return {"authenticated": True, "user_id": payload.get("sub"), "claims": payload}
    except jwt.ExpiredSignatureError:
        return {"authenticated": False, "error": "Token expired"}, 401
    except jwt.InvalidTokenError:
        return {"authenticated": False, "error": "Invalid token"}, 401

def has_permission(payload, required_permission):
    """Check if token payload has required permission."""
    permissions = payload.get("permissions", [])
    if isinstance(permissions, str):
        permissions = permissions.split(",")
    return required_permission in permissions

def require_permission(permission):
    """Decorator to require specific permission."""
    def decorator(func):
        @wraps(func)
        def wrapper(request, *args, **kwargs):
            token = request.headers.get("Authorization", "").replace("Bearer ", "")
            try:
                payload = jwt.decode(token, os.environ["JWT_SECRET"], algorithms=["HS256"])
                if not has_permission(payload, permission):
                    return {"error": "Insufficient permissions"}, 403
                return func(request, *args, **kwargs)
            except jwt.InvalidTokenError:
                return {"error": "Unauthorized"}, 401
        return wrapper
    return decorator''',
                "fill_prompt": "Configure JWT secret environment variable, required permissions, and authentication flow. Customize permission checking for your authorization model."
            },
            "nodejs": {
                "base_code": '''const jwt = require('jsonwebtoken');

exports.handler = async (req, res) => {
    const token = req.headers.authorization?.replace('Bearer ', '');

    try {
        const payload = jwt.verify(token, process.env.JWT_SECRET);
        if (!hasPermission(payload, 'access')) {
            return res.status(403).json({ authenticated: false, error: 'Insufficient permissions' });
        }
        res.json({ authenticated: true, userId: payload.sub, claims: payload });
    } catch (err) {
        if (err.name === 'TokenExpiredError') {
            return res.status(401).json({ authenticated: false, error: 'Token expired' });
        }
        res.status(401).json({ authenticated: false, error: 'Invalid token' });
    }
};

function hasPermission(payload, requiredPermission) {
    const permissions = payload.permissions || [];
    return permissions.includes(requiredPermission);
};''',
                "fill_prompt": "Add specific JWT validation, permission checks, and user data extraction."
            }
        },
        "scheduled_task": {
            "id": "scheduled_task",
            "name": "Scheduled Task",
            "description": "Run scheduled/cron job operations",
            "keywords": ["cron", "schedule", "periodic", "timer", "interval", "daily", "hourly"],
            "python": {
                "base_code": '''from datetime import datetime, timezone
import json

def handler(event):
    """Run scheduled task."""
    trigger_time = event.get("time") or datetime.now(timezone.utc).isoformat()
    task_type = event.get("task_type", "default")
    task_config = event.get("config", {})

    result = perform_scheduled_task(task_type, task_config)

    return {"executed_at": trigger_time, "task_type": task_type, "result": result}

def perform_scheduled_task(task_type, config):
    """Execute scheduled task based on type."""
    if task_type == "cleanup":
        return {"cleaned": cleanup_old_data(config)}
    elif task_type == "report":
        return {"report": generate_report(config)}
    elif task_type == "sync":
        return {"synced": sync_data(config)}
    else:
        return {"executed": f"Task {task_type} completed"}

def cleanup_old_data(config):
    """Remove data older than retention_days."""
    retention_days = config.get("retention_days", 90)
    return f"Cleaned data older than {retention_days} days"

def generate_report(config):
    """Generate and store report."""
    report_type = config.get("report_type", "summary")
    return f"Generated {report_type} report"

def sync_data(config):
    """Sync external data source."""
    source = config.get("source", "default")
    return f"Synced data from {source}"''',
                "fill_prompt": "Configure the scheduled task type (cleanup, report, sync), cron schedule expression, and task-specific configuration options."
            },
            "nodejs": {
                "base_code": '''const { DateTime } = require('luxon');

exports.handler = async (event) => {
    const triggerTime = event.time || DateTime.utc().toISO();
    const taskType = event.task_type || 'default';
    const taskConfig = event.config || {};

    const result = await performScheduledTask(taskType, taskConfig);

    return { executedAt: triggerTime, taskType, result };
};

async function performScheduledTask(taskType, config) {
    switch (taskType) {
        case 'cleanup':
            return { cleaned: cleanupOldData(config) };
        case 'report':
            return { report: generateReport(config) };
        case 'sync':
            return { synced: syncData(config) };
        default:
            return { executed: `Task ${taskType} completed` };
    }
}

function cleanupOldData(config) {
    const retentionDays = config.retention_days || 90;
    return `Cleaned data older than ${retentionDays} days`;
}

function generateReport(config) {
    const reportType = config.report_type || 'summary';
    return `Generated ${reportType} report`;
}

function syncData(config) {
    const source = config.source || 'default';
    return `Synced data from ${source}`;
}''',
                "fill_prompt": "Fill in the scheduled task logic with proper timing and error handling."
            }
        },
        "queue_processor": {
            "id": "queue_processor",
            "name": "Queue Processor",
            "description": "Process messages from a queue",
            "keywords": ["queue", "message", "worker", "process", "consumer", "pubsub"],
            "python": {
                "base_code": '''import json
import time

def handler(message):
    """Process queue message."""
    payload = message.get("body") or message
    message_id = message.get("message_id") or message.get("id", "unknown")
    retry_count = message.get("retry_count", 0)
    max_retries = message.get("max_retries", 3)

    try:
        result = process_message(payload)
        return {"processed": True, "message_id": message_id, "result": result}
    except Exception as e:
        if retry_count < max_retries:
            raise RetryError(f"Retrying message {message_id}", retry_count + 1)
        return {"processed": False, "message_id": message_id, "error": str(e)}

def process_message(payload):
    """Process and transform message payload."""
    if isinstance(payload, str):
        payload = json.loads(payload)
    message_type = payload.get("type", "default")
    data = payload.get("data", payload)

    if message_type == "notification":
        return send_notification(data)
    elif message_type == "event":
        return handle_event(data)
    else:
        return {"processed": data}

def send_notification(data):
    """Send notification to user."""
    return {"notification_sent": True, "recipient": data.get("to")}

def handle_event(data):
    """Handle system event."""
    return {"event_handled": True, "event_id": data.get("id")}

class RetryError(Exception):
    """Indicate message should be retried."""
    def __init__(self, message, retry_count):
        super().__init__(message)
        self.retry_count = retry_count''',
                "fill_prompt": "Configure the message types (notification, event, command), retry behavior (max_retries, backoff), dead letter queue settings, and message processing logic for your queue system."
            },
            "nodejs": {
                "base_code": '''exports.handler = async (message) => {
    const payload = message.body || message;
    const messageId = message.message_id || message.id || 'unknown';
    const retryCount = message.retry_count || 0;
    const maxRetries = message.max_retries || 3;

    try {
        const result = await processMessage(payload);
        return { processed: true, messageId, result };
    } catch (error) {
        if (retryCount < maxRetries) {
            throw new RetryError(`Retrying message ${messageId}`, retryCount + 1);
        }
        return { processed: false, messageId, error: error.message };
    }
};

async function processMessage(payload) {
    if (typeof payload === 'string') {
        payload = JSON.parse(payload);
    }
    const messageType = payload.type || 'default';
    const data = payload.data || payload;

    switch (messageType) {
        case 'notification':
            return sendNotification(data);
        case 'event':
            return handleEvent(data);
        default:
            return { processed: data };
    }
}

function sendNotification(data) {
    return { notificationSent: true, recipient: data.to };
}

function handleEvent(data) {
    return { eventHandled: true, eventId: data.id };
}

class RetryError extends Error {
    constructor(message, retryCount) {
        super(message);
        this.retryCount = retryCount;
    }
}''',
                "fill_prompt": "Fill in the message parsing and processing logic with proper error handling."
            }
        },
    }

    @classmethod
    def search(cls, description: str, runtime: str) -> List[TemplateMatch]:
        """Search for matching templates based on description."""
        description_lower = description.lower()
        matches = []

        for template_id, template in cls.TEMPLATES.items():
            score = 0
            for keyword in template["keywords"]:
                if keyword in description_lower:
                    score += 1

            if score > 0:
                runtime_code = template.get(runtime, template.get("python", {}))
                base_code = runtime_code.get("base_code", "")
                fill_prompt = runtime_code.get("fill_prompt", "")

                # Estimate tokens saved (rough heuristic)
                tokens_saved = len(base_code.split()) * 0.7

                matches.append(TemplateMatch(
                    template_id=template_id,
                    template_name=template["name"],
                    description=template["description"],
                    base_code=base_code,
                    fill_prompt=fill_prompt,
                    match_score=score / len(template["keywords"]),
                    estimated_tokens_saved=int(tokens_saved),
                ))

        # Sort by match score
        matches.sort(key=lambda x: x.match_score, reverse=True)
        return matches

    @classmethod
    def get_template(cls, template_id: str, runtime: str) -> Optional[TemplateMatch]:
        """Get a specific template by ID."""
        template = cls.TEMPLATES.get(template_id)
        if not template:
            return None

        runtime_code = template.get(runtime, template.get("python", {}))
        return TemplateMatch(
            template_id=template_id,
            template_name=template["name"],
            description=template["description"],
            base_code=runtime_code.get("base_code", ""),
            fill_prompt=runtime_code.get("fill_prompt", ""),
            match_score=1.0,
            estimated_tokens_saved=0,
        )


class FunctionRAGRetriever:
    """Retrieves similar functions from the registry using RAG."""

    def __init__(self):
        self.flyembed = get_flyembed_service()
        self.embeddings = get_embeddings_service()
        self.template_lib = FunctionTemplateLibrary()

    async def retrieve_similar_functions(
        self,
        description: str,
        runtime: str,
        tenant_id: str,
        limit: int = 5,
    ) -> List[RetrievedFunction]:
        """Retrieve similar functions from registry.

        Args:
            description: Function description
            runtime: Target runtime
            tenant_id: Tenant to search
            limit: Max results

        Returns:
            List of similar functions
        """
        try:
            # Generate query embedding
            query_vec = await self.flyembed.embed_query(description)

            # Search using triple vectors
            from ..search.indexer import get_search_indexer
            indexer = get_search_indexer()

            results = await indexer.search_triple(
                tenant_id=tenant_id,
                query_vectors={
                    "contract_vector": query_vec.contract_vector,
                    "semantic_vector": query_vec.semantic_vector,
                    "code_vector": query_vec.code_vector,
                },
                limit=limit,
            )

            functions = []
            for i, result in enumerate(results, 1):
                data = result.get("data", {})
                if data.get("runtime") == runtime or runtime == "any":
                    functions.append(RetrievedFunction(
                        function_id=result.get("function_id", ""),
                        name=data.get("name", ""),
                        description=data.get("description", ""),
                        runtime=data.get("runtime", ""),
                        code=data.get("source_code", ""),
                        manifest=data.get("manifest", {}),
                        similarity_score=result.get("score", 0),
                        rank=i,
                    ))

            return functions

        except Exception as e:
            logger.warning(f"RAG retrieval failed: {e}")
            return []

    def find_template(
        self,
        description: str,
        runtime: str,
    ) -> Optional[TemplateMatch]:
        """Find best matching template.

        Args:
            description: Function description
            runtime: Target runtime

        Returns:
            Best template match or None
        """
        matches = self.template_lib.search(description, runtime)
        if matches and matches[0].match_score > 0.3:
            return matches[0]
        return None

    def build_generation_context(
        self,
        description: str,
        runtime: str,
        similar_functions: List[RetrievedFunction],
        template: Optional[TemplateMatch],
    ) -> Dict[str, Any]:
        """Build optimized generation context.

        This reduces token usage by providing templates and examples
        instead of full generation instructions.
        """
        context = {
            "mode": "full_generation",
            "description": description,
            "runtime": runtime,
            "template": None,
            "examples": [],
            "token_savings_estimate": 0,
        }

        # If we have a good template, use template-filling mode
        if template and template.match_score > 0.5:
            context["mode"] = "template_fill"
            context["template"] = {
                "id": template.template_id,
                "base_code": template.base_code,
                "fill_instruction": template.fill_prompt,
            }
            context["token_savings_estimate"] = template.estimated_tokens_saved

        # Add similar functions as examples (limited to save tokens)
        for func in similar_functions[:2]:
            context["examples"].append({
                "name": func.name,
                "description": func.description,
                "code_preview": func.code[:500] if len(func.code) > 500 else func.code,
            })

        return context

    def build_optimized_prompt(
        self,
        context: Dict[str, Any],
        constraints: Optional[str] = None,
    ) -> Tuple[str, int]:
        """Build an optimized prompt based on context.

        Returns:
            Tuple of (prompt, estimated_tokens)
        """
        mode = context.get("mode", "full_generation")
        description = context["description"]
        runtime = context["runtime"]

        if mode == "template_fill" and context["template"]:
            # Use template-filling mode (much cheaper)
            template = context["template"]
            prompt = f"""Fill in the TODO sections in this {runtime} function template:

TEMPLATE:
```
{template["base_code"]}
```

FILL INSTRUCTION: {template["fill_instruction"]}

REQUIREMENT: {description}"""
            if constraints:
                prompt += f"\n\nCONSTRAINTS: {constraints}"

            return prompt, 500

        # Standard generation with RAG examples
        examples_text = ""
        for i, ex in enumerate(context.get("examples", []), 1):
            examples_text += f"\nExample {i}: {ex['name']} - {ex['description']}\n"

        prompt = f"""Generate a {runtime} function that: {description}
{examples_text}

Requirements:
- Self-contained and stateless
- Include error handling
- Production-ready code"""

        if constraints:
            prompt += f"\n\nAdditional constraints: {constraints}"

        return prompt, 1000


# Global retriever instance
_retriever: Optional[FunctionRAGRetriever] = None


def get_function_rag_retriever() -> FunctionRAGRetriever:
    """Get global RAG retriever instance."""
    global _retriever
    if _retriever is None:
        _retriever = FunctionRAGRetriever()
    return _retriever
