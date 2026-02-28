"""
Hello World Function
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
