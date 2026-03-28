import yaml


def handler(event):
    """Generate a Kubernetes ConfigMap manifest."""
    try:
        name = event.get("name")
        data = event.get("data")
        if not name or data is None:
            return {"ok": False, "error": "name and data are required"}

        namespace = event.get("namespace", "default")
        labels = event.get("labels", {"app": name})

        manifest = {
            "apiVersion": "v1",
            "kind": "ConfigMap",
            "metadata": {"name": name, "namespace": namespace, "labels": labels},
            "data": {str(k): str(v) for k, v in data.items()},
        }

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "ConfigMap manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
