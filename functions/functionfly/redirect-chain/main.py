import urllib.request
import urllib.error
import ssl


def handler(event):
    """Follow and return the full redirect chain for a URL."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        max_redirects = int(event.get("max_redirects", 10))
        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        chain = []
        current_url = url
        redirect_count = 0

        for _ in range(max_redirects + 1):
            try:
                req = urllib.request.Request(current_url, method="HEAD")
                # Don't follow redirects automatically
                opener = urllib.request.build_opener(urllib.request.HTTPRedirectHandler())
                opener.addheaders = [("User-Agent", "Mozilla/5.0")]

                class NoRedirect(urllib.request.HTTPRedirectHandler):
                    def redirect_request(self, req, fp, code, msg, headers, newurl):
                        return None

                no_redirect_opener = urllib.request.build_opener(NoRedirect())
                try:
                    with no_redirect_opener.open(req, timeout=timeout) as resp:
                        chain.append({"url": current_url, "status": resp.status})
                        break
                except urllib.error.HTTPError as e:
                    if e.code in (301, 302, 303, 307, 308):
                        chain.append({"url": current_url, "status": e.code})
                        location = e.headers.get("Location")
                        if not location:
                            break
                        if not location.startswith("http"):
                            from urllib.parse import urljoin
                            location = urljoin(current_url, location)
                        current_url = location
                        redirect_count += 1
                    else:
                        chain.append({"url": current_url, "status": e.code})
                        break
            except Exception as e:
                chain.append({"url": current_url, "status": None, "error": str(e)})
                break

        return {
            "ok": True,
            "chain": chain,
            "final_url": current_url,
            "redirect_count": redirect_count,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
