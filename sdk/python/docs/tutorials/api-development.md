# Tutorial: Building REST APIs with FlyPy

This tutorial shows how to build scalable REST APIs using FlyPy, with proper request routing, validation, error handling, and performance optimizations.

## Overview

We'll build a complete REST API for a task management system that includes:

1. **API Gateway** - Route requests to appropriate handlers
2. **CRUD Operations** - Create, read, update, delete tasks
3. **Authentication** - API key validation
4. **Rate Limiting** - Prevent abuse
5. **Error Handling** - Consistent error responses
6. **Data Validation** - Input/output schema validation

## Prerequisites

```bash
pip install flypy
```

## Step 1: Define API Models

Let's define the data models for our API:

```python
# models.py
from typing import List, Dict, Any, Optional
from pydantic import BaseModel
from enum import Enum
from datetime import datetime

class TaskStatus(str, Enum):
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    DONE = "done"
    CANCELLED = "cancelled"

class TaskPriority(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"

class Task(BaseModel):
    id: str
    title: str
    description: Optional[str] = None
    status: TaskStatus = TaskStatus.TODO
    priority: TaskPriority = TaskPriority.MEDIUM
    assignee: Optional[str] = None
    tags: List[str] = []
    created_at: datetime
    updated_at: datetime
    due_date: Optional[datetime] = None

class CreateTaskRequest(BaseModel):
    title: str
    description: Optional[str] = None
    priority: TaskPriority = TaskPriority.MEDIUM
    assignee: Optional[str] = None
    tags: List[str] = []
    due_date: Optional[datetime] = None

class UpdateTaskRequest(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    status: Optional[TaskStatus] = None
    priority: Optional[TaskPriority] = None
    assignee: Optional[str] = None
    tags: Optional[List[str]] = None
    due_date: Optional[datetime] = None

class TaskListResponse(BaseModel):
    tasks: List[Task]
    total: int
    page: int
    page_size: int
    has_more: bool

class APIResponse(BaseModel):
    success: bool
    data: Optional[Any] = None
    error: Optional[Dict[str, Any]] = None
    request_id: str

class APIError(BaseModel):
    code: str
    message: str
    details: Optional[Dict[str, Any]] = None
```

## Step 2: Authentication Middleware

Create an authentication function:

```python
# auth.py
import flypy
from typing import Dict, Any, Optional
import hashlib
import time

# Mock API keys (in real app, store in database)
VALID_API_KEYS = {
    "dev-key-123": {"user_id": "user-123", "permissions": ["read", "write"]},
    "admin-key-456": {"user_id": "admin-456", "permissions": ["read", "write", "admin"]},
}

@flypy.function(
    name="authenticate-request",
    description="Validate API key and return user context",
    deterministic=False,  # Uses external key validation
    capabilities=["auth"]
)
def authenticate_request(headers: Dict[str, Any]) -> Dict[str, Any]:
    """
    Authenticate an API request using API key.

    Args:
        headers: Request headers containing authorization

    Returns:
        Authentication result with user context
    """
    auth_header = headers.get("authorization", "").strip()

    if not auth_header:
        return {
            "authenticated": False,
            "error": "Missing authorization header"
        }

    # Extract API key from Bearer token
    if not auth_header.startswith("Bearer "):
        return {
            "authenticated": False,
            "error": "Invalid authorization format. Use 'Bearer <api-key>'"
        }

    api_key = auth_header[7:].strip()  # Remove "Bearer " prefix

    if not api_key:
        return {
            "authenticated": False,
            "error": "Empty API key"
        }

    # Validate API key
    user_context = VALID_API_KEYS.get(api_key)
    if not user_context:
        return {
            "authenticated": False,
            "error": "Invalid API key"
        }

    return {
        "authenticated": True,
        "user_id": user_context["user_id"],
        "permissions": user_context["permissions"],
        "api_key_hash": hashlib.sha256(api_key.encode()).hexdigest()
    }
```

## Step 3: Rate Limiting Function

Create a rate limiting function:

```python
# rate_limiting.py
import flypy
from typing import Dict, Any
from collections import defaultdict
import time

# In-memory rate limit storage (in real app, use Redis)
rate_limits = defaultdict(list)

@flypy.function(
    name="check-rate-limit",
    description="Check if request is within rate limits",
    deterministic=False,  # Uses external state
    capabilities=["rate_limiting"]
)
def check_rate_limit(user_id: str, endpoint: str, config: Dict[str, Any]) -> Dict[str, Any]:
    """
    Check if a user request is within rate limits.

    Args:
        user_id: User identifier
        endpoint: API endpoint being accessed
        config: Rate limiting configuration

    Returns:
        Rate limit check result
    """
    # Default limits
    limits = config.get("limits", {
        "requests_per_minute": 60,
        "requests_per_hour": 1000
    })

    current_time = time.time()
    user_key = f"{user_id}:{endpoint}"

    # Get user's request history
    user_requests = rate_limits[user_key]

    # Remove old requests outside time windows
    minute_ago = current_time - 60
    hour_ago = current_time - 3600

    user_requests[:] = [req_time for req_time in user_requests
                       if req_time > hour_ago]  # Keep requests from last hour

    # Count requests in current windows
    minute_requests = sum(1 for req_time in user_requests if req_time > minute_ago)
    hour_requests = len(user_requests)

    # Check limits
    minute_limit = limits["requests_per_minute"]
    hour_limit = limits["requests_per_hour"]

    if minute_requests >= minute_limit:
        return {
            "allowed": False,
            "reason": "minute_limit_exceeded",
            "reset_in_seconds": 60 - int(current_time - minute_ago),
            "current_usage": {
                "minute": minute_requests,
                "hour": hour_requests
            }
        }

    if hour_requests >= hour_limit:
        return {
            "allowed": False,
            "reason": "hour_limit_exceeded",
            "reset_in_seconds": 3600 - int(current_time - hour_ago),
            "current_usage": {
                "minute": minute_requests,
                "hour": hour_requests
            }
        }

    # Add current request to history
    user_requests.append(current_time)

    return {
        "allowed": True,
        "current_usage": {
            "minute": minute_requests + 1,
            "hour": hour_requests + 1
        },
        "limits": limits
    }
```

## Step 4: Task CRUD Operations

Create the main CRUD operations for tasks:

```python
# tasks.py
import flypy
from typing import Dict, Any, List, Optional
from models import Task, CreateTaskRequest, UpdateTaskRequest, TaskListResponse
from datetime import datetime
import uuid

# Mock database (in real app, use actual database)
tasks_db = {}

@flypy.function(
    name="create-task",
    description="Create a new task",
    deterministic=False,  # Creates new resources
    capabilities=["database"]
)
def create_task(request: Dict[str, Any], user_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    Create a new task.

    Args:
        request: Task creation data
        user_context: Authenticated user context

    Returns:
        Created task data
    """
    # Validate request data
    create_request = CreateTaskRequest(**request)

    # Generate task ID
    task_id = str(uuid.uuid4())

    # Create task
    now = datetime.utcnow()
    task = Task(
        id=task_id,
        title=create_request.title,
        description=create_request.description,
        priority=create_request.priority,
        assignee=create_request.assignee,
        tags=create_request.tags,
        due_date=create_request.due_date,
        created_at=now,
        updated_at=now
    )

    # Store in database
    tasks_db[task_id] = task

    return {
        "task": task.dict(),
        "created": True
    }

@flypy.function(
    name="get-task",
    description="Get a task by ID",
    deterministic=False,  # Reads from external database
    capabilities=["database"]
)
def get_task(task_id: str, user_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    Get a task by ID.

    Args:
        task_id: Task identifier
        user_context: Authenticated user context

    Returns:
        Task data or error
    """
    task = tasks_db.get(task_id)

    if not task:
        return {
            "found": False,
            "error": "Task not found"
        }

    return {
        "found": True,
        "task": task.dict()
    }

@flypy.function(
    name="list-tasks",
    description="List tasks with pagination and filtering",
    deterministic=False,  # Reads from external database
    capabilities=["database"]
)
def list_tasks(query: Dict[str, Any], user_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    List tasks with pagination and filtering.

    Args:
        query: Query parameters
        user_context: Authenticated user context

    Returns:
        Paginated task list
    """
    # Extract query parameters
    page = max(1, query.get("page", 1))
    page_size = min(100, max(1, query.get("page_size", 20)))
    status_filter = query.get("status")
    assignee_filter = query.get("assignee")
    priority_filter = query.get("priority")

    # Filter tasks
    filtered_tasks = []
    for task in tasks_db.values():
        if status_filter and task.status != status_filter:
            continue
        if assignee_filter and task.assignee != assignee_filter:
            continue
        if priority_filter and task.priority != priority_filter:
            continue

        filtered_tasks.append(task)

    # Sort by created date (newest first)
    filtered_tasks.sort(key=lambda t: t.created_at, reverse=True)

    # Paginate
    total = len(filtered_tasks)
    start_idx = (page - 1) * page_size
    end_idx = start_idx + page_size
    page_tasks = filtered_tasks[start_idx:end_idx]

    return {
        "tasks": [task.dict() for task in page_tasks],
        "total": total,
        "page": page,
        "page_size": page_size,
        "has_more": end_idx < total
    }

@flypy.function(
    name="update-task",
    description="Update an existing task",
    deterministic=False,  # Updates external database
    capabilities=["database"]
)
def update_task(task_id: str, updates: Dict[str, Any], user_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    Update an existing task.

    Args:
        task_id: Task identifier
        updates: Fields to update
        user_context: Authenticated user context

    Returns:
        Updated task data
    """
    task = tasks_db.get(task_id)

    if not task:
        return {
            "found": False,
            "error": "Task not found"
        }

    # Validate update data
    update_request = UpdateTaskRequest(**updates)

    # Apply updates
    update_data = update_request.dict(exclude_unset=True)

    for field, value in update_data.items():
        setattr(task, field, value)

    task.updated_at = datetime.utcnow()

    # Store updated task
    tasks_db[task_id] = task

    return {
        "found": True,
        "task": task.dict(),
        "updated": True
    }

@flypy.function(
    name="delete-task",
    description="Delete a task",
    deterministic=False,  # Deletes from external database
    capabilities=["database"]
)
def delete_task(task_id: str, user_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    Delete a task.

    Args:
        task_id: Task identifier
        user_context: Authenticated user context

    Returns:
        Deletion result
    """
    if task_id not in tasks_db:
        return {
            "found": False,
            "error": "Task not found"
        }

    # Delete task
    del tasks_db[task_id]

    return {
        "found": True,
        "deleted": True
    }
```

## Step 5: API Gateway Function

Create the main API gateway that routes requests:

```python
# api_gateway.py
import flypy
from typing import Dict, Any
import json
import uuid
import time

# Import API functions
from auth import authenticate_request
from rate_limiting import check_rate_limit
from tasks import create_task, get_task, list_tasks, update_task, delete_task

@flypy.function(
    name="api-gateway",
    description="Main API gateway for task management",
    deterministic=False,  # Orchestrates multiple functions
    max_execution_time=30000  # 30 seconds
)
def api_gateway(request: Dict[str, Any]) -> Dict[str, Any]:
    """
    Main API gateway that handles routing, authentication, and orchestration.

    Args:
        request: HTTP request data

    Returns:
        API response
    """
    request_id = str(uuid.uuid4())
    start_time = time.time()

    try:
        # Extract request components
        method = request.get("method", "GET")
        path = request.get("path", "/")
        headers = request.get("headers", {})
        query_params = request.get("query", {})
        body = request.get("body", {})

        # Step 1: Authentication
        auth_result = authenticate_request(headers)
        if not auth_result["authenticated"]:
            return create_error_response(
                "UNAUTHORIZED",
                auth_result["error"],
                request_id,
                status_code=401
            )

        user_context = auth_result

        # Step 2: Rate Limiting
        rate_limit_config = {"limits": {"requests_per_minute": 60, "requests_per_hour": 1000}}
        rate_check = check_rate_limit(user_context["user_id"], path, rate_limit_config)

        if not rate_check["allowed"]:
            return create_error_response(
                "RATE_LIMIT_EXCEEDED",
                "Too many requests",
                request_id,
                status_code=429,
                details={
                    "reason": rate_check["reason"],
                    "reset_in_seconds": rate_check["reset_in_seconds"],
                    "current_usage": rate_check["current_usage"]
                }
            )

        # Step 3: Route request
        route_result = route_request(method, path, query_params, body, user_context)

        # Calculate processing time
        processing_time_ms = int((time.time() - start_time) * 1000)

        return create_success_response(
            route_result["data"],
            request_id,
            status_code=route_result["status_code"],
            processing_time_ms=processing_time_ms
        )

    except Exception as e:
        processing_time_ms = int((time.time() - start_time) * 1000)
        return create_error_response(
            "INTERNAL_ERROR",
            "An unexpected error occurred",
            request_id,
            status_code=500,
            details={"error": str(e)},
            processing_time_ms=processing_time_ms
        )

def route_request(method: str, path: str, query: Dict[str, Any], body: Dict[str, Any], user_context: Dict[str, Any]) -> Dict[str, Any]:
    """Route the request to the appropriate handler."""

    # Parse path
    path_parts = path.strip("/").split("/")
    if not path_parts or path_parts == [""]:
        path_parts = []

    # Task routes
    if len(path_parts) >= 1 and path_parts[0] == "tasks":

        if len(path_parts) == 1:
            # /tasks
            if method == "GET":
                result = list_tasks(query, user_context)
                return {"data": result, "status_code": 200}

            elif method == "POST":
                result = create_task(body, user_context)
                if result["created"]:
                    return {"data": result["task"], "status_code": 201}
                else:
                    return {"data": {"error": "Failed to create task"}, "status_code": 400}

        elif len(path_parts) == 2:
            # /tasks/{id}
            task_id = path_parts[1]

            if method == "GET":
                result = get_task(task_id, user_context)
                if result["found"]:
                    return {"data": result["task"], "status_code": 200}
                else:
                    return {"data": {"error": result["error"]}, "status_code": 404}

            elif method == "PUT":
                result = update_task(task_id, body, user_context)
                if result["found"]:
                    return {"data": result["task"], "status_code": 200}
                else:
                    return {"data": {"error": result["error"]}, "status_code": 404}

            elif method == "DELETE":
                result = delete_task(task_id, user_context)
                if result["found"]:
                    return {"data": {"deleted": True}, "status_code": 200}
                else:
                    return {"data": {"error": result["error"]}, "status_code": 404}

    # Unknown route
    return {
        "data": {"error": f"Route not found: {method} {path}"},
        "status_code": 404
    }

def create_success_response(data: Any, request_id: str, status_code: int = 200, processing_time_ms: int = None) -> Dict[str, Any]:
    """Create a standardized success response."""
    response = {
        "success": True,
        "data": data,
        "request_id": request_id,
        "timestamp": time.time()
    }

    if processing_time_ms is not None:
        response["processing_time_ms"] = processing_time_ms

    return {
        "status_code": status_code,
        "headers": {
            "Content-Type": "application/json",
            "X-Request-ID": request_id
        },
        "body": response
    }

def create_error_response(error_code: str, message: str, request_id: str, status_code: int = 500, details: Dict[str, Any] = None, processing_time_ms: int = None) -> Dict[str, Any]:
    """Create a standardized error response."""
    error = {
        "code": error_code,
        "message": message
    }

    if details:
        error["details"] = details

    response = {
        "success": False,
        "error": error,
        "request_id": request_id,
        "timestamp": time.time()
    }

    if processing_time_ms is not None:
        response["processing_time_ms"] = processing_time_ms

    return {
        "status_code": status_code,
        "headers": {
            "Content-Type": "application/json",
            "X-Request-ID": request_id
        },
        "body": response
    }
```

## Step 6: Testing the API

Create a test script to verify the API works:

```python
# test_api.py
import json
from api_gateway import api_gateway

# Test data
test_requests = [
    # Create a task
    {
        "method": "POST",
        "path": "/tasks",
        "headers": {"authorization": "Bearer dev-key-123"},
        "body": {
            "title": "Implement user authentication",
            "description": "Add OAuth2 support for user login",
            "priority": "high",
            "tags": ["auth", "security"]
        }
    },

    # List tasks
    {
        "method": "GET",
        "path": "/tasks",
        "headers": {"authorization": "Bearer dev-key-123"},
        "query": {"page": 1, "page_size": 10}
    },

    # Get specific task (will need to use the ID from create response)
    # {
    #     "method": "GET",
    #     "path": "/tasks/task-id-here",
    #     "headers": {"authorization": "Bearer dev-key-123"}
    # },

    # Test authentication failure
    {
        "method": "GET",
        "path": "/tasks",
        "headers": {"authorization": "Bearer invalid-key"},
    },

    # Test rate limiting (would need multiple rapid requests)
    # Test invalid route
    {
        "method": "GET",
        "path": "/invalid-route",
        "headers": {"authorization": "Bearer dev-key-123"},
    }
]

if __name__ == "__main__":
    for i, request in enumerate(test_requests):
        print(f"\n=== Test Request {i+1} ===")
        print(f"Method: {request['method']}")
        print(f"Path: {request['path']}")
        print(f"Headers: {request.get('headers', {})}")

        try:
            response = api_gateway(request)

            print(f"Status Code: {response['status_code']}")
            print("Response Body:")
            print(json.dumps(response['body'], indent=2, default=str))

        except Exception as e:
            print(f"Error: {e}")
```

## Step 7: Build and Deploy

Build and deploy the API:

```bash
# Build all API functions
flypy build auth.py rate_limiting.py tasks.py api_gateway.py

# Test locally
flypy local api_gateway.py api-gateway --port 8080

# Deploy to FunctionFly
flypy deploy ./dist/api-gateway --token YOUR_TOKEN --app-id YOUR_APP_ID
```

## Step 8: API Usage Examples

Once deployed, you can interact with the API:

```python
import requests

BASE_URL = "https://your-functionfly-app.com"
API_KEY = "dev-key-123"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# Create a task
task_data = {
    "title": "Review pull request",
    "description": "Review the authentication PR",
    "priority": "medium",
    "tags": ["code-review", "auth"]
}

response = requests.post(f"{BASE_URL}/tasks", json=task_data, headers=headers)
task = response.json()

print(f"Created task: {task['data']['id']}")

# List tasks
response = requests.get(f"{BASE_URL}/tasks?page=1&page_size=10", headers=headers)
tasks = response.json()

print(f"Found {tasks['data']['total']} tasks")

# Get specific task
task_id = task['data']['id']
response = requests.get(f"{BASE_URL}/tasks/{task_id}", headers=headers)
specific_task = response.json()

print(f"Task title: {specific_task['data']['title']}")

# Update task
update_data = {
    "status": "in_progress",
    "assignee": "developer@example.com"
}

response = requests.put(f"{BASE_URL}/tasks/{task_id}", json=update_data, headers=headers)
updated_task = response.json()

print(f"Task status: {updated_task['data']['status']}")

# Delete task
response = requests.delete(f"{BASE_URL}/tasks/{task_id}", headers=headers)
print(f"Delete status: {response.status_code}")
```

## Summary

This tutorial demonstrated how to build a complete REST API using FlyPy with:

- **Authentication** - API key validation
- **Rate limiting** - Prevent abuse
- **CRUD operations** - Full task management
- **Error handling** - Consistent error responses
- **Request routing** - RESTful URL routing
- **Data validation** - Input/output schema validation

The modular function architecture makes it easy to:
- Test individual components
- Scale different parts of the API
- Add new endpoints and features
- Monitor API performance and usage

This approach is ideal for building scalable, secure, and maintainable APIs that can handle production workloads.