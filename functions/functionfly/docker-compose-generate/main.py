import yaml


def handler(event):
    """Generate a docker-compose.yml file for multi-service applications."""
    try:
        services = event.get("services")
        if not services:
            return {"ok": False, "error": "services is required"}

        version = event.get("version", "3.8")
        networks = event.get("networks", [])
        volumes = event.get("volumes", [])

        compose = {"version": version, "services": {}}

        for svc in services:
            name = svc.get("name")
            if not name:
                continue
            svc_def = {}
            if svc.get("image"):
                svc_def["image"] = svc["image"]
            if svc.get("build"):
                svc_def["build"] = svc["build"]
            if svc.get("ports"):
                svc_def["ports"] = svc["ports"]
            if svc.get("environment"):
                svc_def["environment"] = svc["environment"]
            if svc.get("volumes"):
                svc_def["volumes"] = svc["volumes"]
            if svc.get("depends_on"):
                svc_def["depends_on"] = svc["depends_on"]
            if svc.get("command"):
                svc_def["command"] = svc["command"]
            if svc.get("restart"):
                svc_def["restart"] = svc["restart"]
            compose["services"][name] = svc_def

        if networks:
            compose["networks"] = {n: {} for n in networks}
        if volumes:
            compose["volumes"] = {v: {} for v in volumes}

        compose_str = yaml.dump(compose, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "docker-compose.yml generated", "compose_file": compose_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
