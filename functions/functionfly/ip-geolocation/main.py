import ipaddress
import json
import urllib.request
import urllib.error


def _is_private(ip_str):
    try:
        ip = ipaddress.ip_address(ip_str.strip())
        return ip.is_private or ip.is_loopback or ip.is_reserved or str(ip) == "127.0.0.1"
    except ValueError:
        return False


def handler(event):
    if isinstance(event, dict):
        ip_str = (event.get("ip") or event.get("address") or "").strip()
        api_provider = (event.get("api_provider") or "none").lower()
    else:
        ip_str, api_provider = "", "none"

    if not ip_str:
        return {"ok": False, "error": "Input 'ip' is required"}

    try:
        ipaddress.ip_address(ip_str)
    except ValueError:
        return {"ok": False, "error": "Invalid IP address"}

    if _is_private(ip_str):
        return {"ok": True, "ip": ip_str, "country": "", "city": "", "region": "", "lat": None, "lon": None, "is_private": True}

    if api_provider == "ip-api" or api_provider == "ip-api.com":
        try:
            url = f"http://ip-api.com/json/{ip_str}?fields=status,country,regionName,city,lat,lon"
            req = urllib.request.Request(url, headers={"User-Agent": "FunctionFly/1.0"})
            with urllib.request.urlopen(req, timeout=5) as r:
                data = json.loads(r.read().decode())
            if data.get("status") != "success":
                return {"ok": True, "ip": ip_str, "country": "", "city": "", "region": "", "lat": None, "lon": None, "is_private": False}
            return {"ok": True, "ip": ip_str, "country": data.get("country", ""), "city": data.get("city", ""), "region": data.get("regionName", ""), "lat": data.get("lat"), "lon": data.get("lon"), "is_private": False}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    return {"ok": True, "ip": ip_str, "country": "", "city": "", "region": "", "lat": None, "lon": None, "is_private": False}
