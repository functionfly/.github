def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        lines = str(text).splitlines()
        rules = {}
        current_agents = []
        sitemaps = []
        crawl_delay = {}
        for line in lines:
            stripped = line.split("#")[0].strip()
            if not stripped:
                if current_agents:
                    current_agents = []
                continue
            if ":" not in stripped:
                continue
            key, _, value = stripped.partition(":")
            key, value = key.strip().lower(), value.strip()
            if key == "user-agent":
                current_agents.append(value)
                for agent in current_agents:
                    if agent not in rules:
                        rules[agent] = {"allow": [], "disallow": []}
            elif key == "disallow":
                for agent in current_agents:
                    if agent in rules and value:
                        rules[agent]["disallow"].append(value)
            elif key == "allow":
                for agent in current_agents:
                    if agent in rules and value:
                        rules[agent]["allow"].append(value)
            elif key == "sitemap":
                sitemaps.append(value)
            elif key == "crawl-delay":
                for agent in current_agents:
                    crawl_delay[agent] = float(value)
        return {
            "ok": True,
            "result": rules,
            "rules": rules,
            "sitemaps": sitemaps,
            "crawl_delay": crawl_delay,
            "agents": list(rules.keys())
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
