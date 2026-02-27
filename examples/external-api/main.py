import os
import json

def handler(input_data):
    """
    Make external API calls using the external_api capability.
    """
    try:
        base_url = os.getenv('API_BASE_URL', 'https://httpbin.org')
        method = input_data.get('method', 'GET')
        endpoint = input_data.get('endpoint', '/get')
        headers = input_data.get('headers', {})
        body = input_data.get('body')

        # Construct full URL
        url = f"{base_url}{endpoint}"

        # Prepare request data
        request_data = {
            "method": method,
            "url": url,
            "headers": headers
        }

        if body:
            request_data["body"] = body

        # In a real implementation, this would call the external_api host function
        # For now, we'll simulate the API call

        # Simulate different API responses based on endpoint
        if '/post' in endpoint:
            simulated_response = {
                "url": url,
                "method": method,
                "data": body or {},
                "headers": headers,
                "simulated": True
            }
        elif '/get' in endpoint:
            simulated_response = {
                "url": url,
                "method": method,
                "args": {},
                "headers": headers,
                "simulated": True
            }
        else:
            simulated_response = {
                "url": url,
                "method": method,
                "status": "simulated",
                "simulated": True
            }

        result = {
            "status": "success",
            "request": request_data,
            "response": simulated_response,
            "response_length": len(json.dumps(simulated_response))
        }

        return result

    except Exception as e:
        return {
            "status": "error",
            "message": f"External API call failed: {str(e)}"
        }