import yaml


def handler(event):
    """Generate a Kubernetes Deployment manifest."""
    try:
        name = event.get("name")
        image = event.get("image")
        if not name or not image:
            return {"ok": False, "error": "name and image are required"}

        replicas = event.get("replicas", 1)
        namespace = event.get("namespace", "default")
        ports = event.get("ports", [])
        env = event.get("env", {})
        resources = event.get("resources", {})
        labels = event.get("labels", {"app": name})

        container = {"name": name, "image": image}
        if ports:
            container["ports"] = [{"containerPort": p} for p in ports]
        if env:
            container["env"] = [{"name": k, "value": str(v)} for k, v in env.items()]
        if resources:
            container["resources"] = resources

        manifest = {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {"name": name, "namespace": namespace, "labels": labels},
            "spec": {
                "replicas": replicas,
                "selector": {"matchLabels": labels},
                "template": {
                    "metadata": {"labels": labels},
                    "spec": {"containers": [container]},
                },
            },
        }

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "Deployment manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
