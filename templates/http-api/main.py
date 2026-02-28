"""
HTTP API Function
A RESTful API with built-in routing for CRUD operations.
"""

import json
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
        # GET /users - List all users
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
        
        # POST /users - Create new user
        if method == 'POST':
            try:
                body = await request.json()
            except:
                body = {}
            
            global user_id_counter
            user_id = user_id_counter
            user_id_counter += 1
            
            user = {
                "id": user_id,
                "name": body.get("name", ""),
                "email": body.get("email", ""),
                "created_at": datetime.utcnow().isoformat()
            }
            
            users_db[user_id] = user
            
            return {
                "status": 201,
                "body": user,
                "headers": {"Content-Type": "application/json"}
            }
    
    # Single user endpoints
    if '/users/' in url:
        user_id = url.split('/users/')[-1]
        
        if not user_id.isdigit():
            return {
                "status": 400,
                "body": {"error": "Invalid user ID"},
                "headers": {"Content-Type": "application/json"}
            }
        
        user_id = int(user_id)
        
        # GET /users/{id}
        if method == 'GET':
            user = users_db.get(user_id)
            if not user:
                return {
                    "status": 404,
                    "body": {"error": "User not found"},
                    "headers": {"Content-Type": "application/json"}
                }
            return {
                "status": 200,
                "body": user,
                "headers": {"Content-Type": "application/json"}
            }
        
        # PUT /users/{id}
        if method == 'PUT':
            user = users_db.get(user_id)
            if not user:
                return {
                    "status": 404,
                    "body": {"error": "User not found"},
                    "headers": {"Content-Type": "application/json"}
                }
            
            try:
                body = await request.json()
            except:
                body = {}
            
            user["name"] = body.get("name", user["name"])
            user["email"] = body.get("email", user["email"])
            user["updated_at"] = datetime.utcnow().isoformat()
            
            return {
                "status": 200,
                "body": user,
                "headers": {"Content-Type": "application/json"}
            }
        
        # DELETE /users/{id}
        if method == 'DELETE':
            if user_id not in users_db:
                return {
                    "status": 404,
                    "body": {"error": "User not found"},
                    "headers": {"Content-Type": "application/json"}
                }
            
            del users_db[user_id]
            
            return {
                "status": 204,
                "body": "",
                "headers": {}
            }
    
    # Default: 404 Not Found
    return {
        "status": 404,
        "body": {"error": "Not found", "path": url},
        "headers": {"Content-Type": "application/json"}
    }
