def handler(event):
    """Parse a robots.txt file."""
    try:
        content = event.get("content")
        if not content:
            return {"ok": False, "error": "content is required"}

        target_agent = event.get("user_agent", "*")
        rules = {}
        sitemaps = []
        current_agent = None

        for line in content.splitlines():
            line = line.strip()
            if not line or line.startswith("#"):
                continue

            if ":" not in line:
                continue

            key, _, value = line.partition(":")
            key = key.strip().lower()
            value = value.strip()

            if key == "user-agent":
                current_agent = value
                if current_agent not in rules:
                    rules[current_agent] = {"allow": [], "disallow": [], "crawl_delay": None}
            elif key == "disallow" and current_agent:
                if value:
                    rules[current_agent]["disallow"].append(value)
            elif key == "allow" and current_agent:
                if value:
                    rules[current_agent]["allow"].append(value)
            elif key == "crawl-delay" and current_agent:
                try:
                    rules[current_agent]["crawl_delay"] = float(value)
                except ValueError:
                    pass
            elif key == "sitemap":
                sitemaps.append(value)

        # Get rules for target agent
        agent_rules = rules.get(target_agent) or rules.get("*")
        crawl_delay = agent_rules.get("crawl_delay") if agent_rules else None

        return {
            "ok": True,
            "rules": rules,
            "sitemaps": sitemaps,
            "crawl_delay": crawl_delay,
            "agent_rules": agent_rules,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
