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
                "base_code": '''def handler(request):
    """Handle incoming webhook request."""
    # TODO: Add webhook validation
    payload = request.get_json()
    
    # TODO: Process the payload
    result = process_webhook(payload)
    
    return {{"status": "success", "processed": result}}

def process_webhook(payload):
    # TODO: Implement webhook processing logic
    return payload''',
                "fill_prompt": "Fill in the webhook validation and processing logic based on the specific webhook type."
            },
            "nodejs": {
                "base_code": '''exports.handler = async (req, res) => {
    // TODO: Add webhook validation
    const payload = req.body;
    
    // TODO: Process the payload
    const result = await processWebhook(payload);
    
    res.json({ status: "success", processed: result });
};

async function processWebhook(payload) {
    // TODO: Implement webhook processing logic
    return payload;
}''',
                "fill_prompt": "Fill in the webhook validation and processing logic based on the specific webhook type."
            }
        },
        "api_client": {
            "id": "api_client",
            "name": "API Client",
            "description": "Make HTTP API calls to external services",
            "keywords": ["api", "http", "request", "fetch", "get", "post", "call external"],
            "python": {
                "base_code": '''import requests

def handler(input):
    """Make API call to external service."""
    url = input.get("url")
    method = input.get("method", "GET")
    headers = input.get("headers", {})
    body = input.get("body")
    
    # TODO: Add authentication if needed
    
    response = requests.request(method, url, headers=headers, json=body)
    response.raise_for_status()
    
    return {{"status_code": response.status_code, "data": response.json()}}''',
                "fill_prompt": "Add specific API endpoint details, authentication, error handling, and response parsing."
            },
            "nodejs": {
                "base_code": '''const axios = require('axios');

exports.handler = async (input) => {
    const url = input.url;
    const method = input.method || 'GET';
    const headers = input.headers || {};
    const body = input.body;
    
    // TODO: Add authentication if needed
    
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
                "base_code": '''def handler(input):
    """Transform input data to output format."""
    data = input.get("data")
    
    # TODO: Implement transformation logic
    result = transform_data(data)
    
    return {{"result": result}}

def transform_data(data):
    # TODO: Add transformation logic
    return data''',
                "fill_prompt": "Fill in the data transformation logic based on the input and output schemas."
            },
            "nodejs": {
                "base_code": '''exports.handler = async (input) => {
    const data = input.data;
    
    // TODO: Implement transformation logic
    const result = transformData(data);
    
    return { result };
};

function transformData(data) {
    // TODO: Add transformation logic
    return data;
}''',
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
import psycopg2

def handler(input):
    """Perform database operation."""
    # TODO: Extract operation details from input
    operation = input.get("operation")
    
    conn = psycopg2.connect(os.environ["DATABASE_URL"])
    try:
        with conn.cursor() as cur:
            # TODO: Implement database operation
            cur.execute("SELECT 1")
            result = cur.fetchall()
        return {{"result": result}}
    finally:
        conn.close()''',
                "fill_prompt": "Fill in the specific SQL operations, table names, and parameter handling."
            },
            "nodejs": {
                "base_code": '''const {{ Pool }} = require('pg');

const pool = new Pool({
    connectionString: process.env.DATABASE_URL
});

exports.handler = async (input) => {
    // TODO: Extract operation details from input
    const operation = input.operation;
    
    const client = await pool.connect();
    try {
        // TODO: Implement database operation
        const result = await client.query('SELECT 1');
        return { result: result.rows };
    } finally {
        client.release();
    }
};''',
                "fill_prompt": "Fill in the specific SQL operations, table names, and parameter handling."
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

def handler(request):
    """Handle authentication request."""
    # TODO: Extract token from request
    token = request.headers.get("Authorization", "").replace("Bearer ", "")
    
    try:
        payload = jwt.decode(token, os.environ["JWT_SECRET"], algorithms=["HS256"])
        # TODO: Validate permissions
        return {{"authenticated": True, "user_id": payload.get("sub")}}
    except jwt.InvalidTokenError:
        return {{"authenticated": False, "error": "Invalid token"}}''',
                "fill_prompt": "Add specific JWT validation, permission checks, and user data extraction."
            },
            "nodejs": {
                "base_code": '''const jwt = require('jsonwebtoken');

exports.handler = async (req, res) => {
    // TODO: Extract token from request
    const token = req.headers.authorization?.replace('Bearer ', '');
    
    try {
        const payload = jwt.verify(token, process.env.JWT_SECRET);
        // TODO: Validate permissions
        res.json({ authenticated: true, userId: payload.sub });
    } catch (err) {
        res.status(401).json({ authenticated: false, error: 'Invalid token' });
    }
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
                "base_code": '''def handler(event):
    """Run scheduled task."""
    # TODO: Get current time or trigger info
    trigger_time = event.get("time")
    
    # TODO: Implement scheduled task logic
    result = perform_scheduled_task()
    
    return {{"executed_at": trigger_time, "result": result}}

def perform_scheduled_task():
    # TODO: Add scheduled task logic
    return "Task completed"''',
                "fill_prompt": "Fill in the scheduled task logic with proper timing and error handling."
            },
            "nodejs": {
                "base_code": '''exports.handler = async (event) => {
    // TODO: Get current time or trigger info
    const triggerTime = event.time;
    
    // TODO: Implement scheduled task logic
    const result = await performScheduledTask();
    
    return { executedAt: triggerTime, result };
};

async function performScheduledTask() {
    // TODO: Add scheduled task logic
    return 'Task completed';
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
                "base_code": '''def handler(message):
    """Process queue message."""
    # TODO: Parse message payload
    payload = message.get("body")
    
    # TODO: Process the message
    result = process_message(payload)
    
    return {{"processed": True, "result": result}}

def process_message(payload):
    # TODO: Implement message processing
    return payload''',
                "fill_prompt": "Fill in the message parsing and processing logic with proper error handling."
            },
            "nodejs": {
                "base_code": '''exports.handler = async (message) => {
    // TODO: Parse message payload
    const payload = message.body;
    
    // TODO: Process the message
    const result = await processMessage(payload);
    
    return { processed: true, result };
};

async function processMessage(payload) {
    // TODO: Implement message processing
    return payload;
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
