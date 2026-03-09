import urllib.request
import urllib.error
import json


def handler(event):
    url = event.get("url")
    method = event.get("method", "GET").upper()
    headers = event.get("headers", {})
    data = event.get("data")
    max_redirects = event.get("max_redirects", 5)
    timeout = event.get("timeout", 30)

    if not url:
        return {"ok": False, "error": "url is required"}

    if not isinstance(url, str):
        return {"ok": False, "error": "url must be a string"}

    if max_redirects < 0 or max_redirects > 20:
        return {"ok": False, "error": "max_redirects must be between 0 and 20"}

    try:
        current_url = url
        redirect_chain = []
        response_data = None
        final_response = None

        # Disable automatic redirects
        class NoRedirect(urllib.request.HTTPRedirectHandler):
            def redirect_request(self, req, fp, code, msg, headers, newurl):
                return None

        opener = urllib.request.build_opener(NoRedirect)
        urllib.request.install_opener(opener)

        for redirect_count in range(max_redirects + 1):
            redirect_chain.append({
                "url": current_url,
                "method": method,
                "redirect_count": redirect_count
            })

            # Prepare request
            if "User-Agent" not in headers:
                headers["User-Agent"] = "FunctionFly/1.0"

            req = urllib.request.Request(current_url, headers=headers, method=method)

            # Add data for POST/PUT/PATCH
            if data and method in ["POST", "PUT", "PATCH"]:
                if isinstance(data, dict):
                    req.data = json.dumps(data).encode('utf-8')
                    headers["Content-Type"] = "application/json"
                elif isinstance(data, str):
                    req.data = data.encode('utf-8')
                else:
                    req.data = json.dumps(data).encode('utf-8')
                    headers["Content-Type"] = "application/json"

            try:
                with urllib.request.urlopen(req, timeout=timeout) as response:
                    final_response = {
                        "status_code": response.status,
                        "headers": dict(response.headers),
                        "url": response.url,
                        "redirect_count": redirect_count
                    }

                    # Read response content
                    content = response.read()
                    content_type = response.headers.get('Content-Type', '')

                    try:
                        if 'application/json' in content_type:
                            response_data = json.loads(content.decode('utf-8'))
                        else:
                            response_data = content.decode('utf-8')
                    except (UnicodeDecodeError, json.JSONDecodeError):
                        import base64
                        response_data = base64.b64encode(content).decode('utf-8')

                    break  # Success, no more redirects needed

            except urllib.error.HTTPError as e:
                if e.code in [301, 302, 303, 307, 308]:  # Redirect status codes
                    redirect_url = e.headers.get('Location')
                    if not redirect_url:
                        return {
                            "ok": False,
                            "error": f"Redirect status {e.code} but no Location header",
                            "redirect_chain": redirect_chain,
                            "final_response": {
                                "status_code": e.code,
                                "headers": dict(e.headers) if e.headers else {},
                                "url": current_url,
                                "redirect_count": redirect_count
                            }
                        }

                    # Handle relative URLs
                    if not redirect_url.startswith(('http://', 'https://')):
                        from urllib.parse import urljoin
                        redirect_url = urljoin(current_url, redirect_url)

                    current_url = redirect_url

                    # For 303, change method to GET
                    if e.code == 303 and method != "GET":
                        method = "GET"
                        data = None  # Remove body for GET

                    continue
                else:
                    # Non-redirect error
                    return {
                        "ok": False,
                        "error": f"HTTP {e.code}: {e.reason}",
                        "status_code": e.code,
                        "headers": dict(e.headers) if e.headers else {},
                        "redirect_chain": redirect_chain,
                        "redirect_count": redirect_count
                    }

            except urllib.error.URLError as e:
                return {
                    "ok": False,
                    "error": f"URL Error: {str(e)}",
                    "redirect_chain": redirect_chain,
                    "redirect_count": redirect_count
                }

        if final_response:
            return {
                "ok": True,
                "result": response_data,
                "final_response": final_response,
                "redirect_chain": redirect_chain,
                "total_redirects": len(redirect_chain) - 1
            }
        else:
            return {
                "ok": False,
                "error": f"Maximum redirects ({max_redirects}) exceeded",
                "redirect_chain": redirect_chain,
                "redirect_count": len(redirect_chain) - 1
            }

    except Exception as e:
        return {"ok": False, "error": f"Request failed: {str(e)}"}