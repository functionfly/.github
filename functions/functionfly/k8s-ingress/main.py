import yaml


def handler(event):
    """Generate a Kubernetes Ingress manifest."""
    try:
        name = event.get("name")
        host = event.get("host")
        if not name or not host:
            return {"ok": False, "error": "name and host are required"}

        paths = event.get("paths", [{"path": "/", "service_name": name, "service_port": 80}])
        tls = event.get("tls", [])
        namespace = event.get("namespace", "default")
        ingress_class = event.get("ingress_class", "nginx")

        path_specs = []
        for p in paths:
            path_specs.append({
                "path": p.get("path", "/"),
                "pathType": p.get("path_type", "Prefix"),
                "backend": {
                    "service": {
                        "name": p.get("service_name", name),
                        "port": {"number": p.get("service_port", 80)},
                    }
                },
            })

        manifest = {
            "apiVersion": "networking.k8s.io/v1",
            "kind": "Ingress",
            "metadata": {
                "name": name,
                "namespace": namespace,
                "annotations": {"kubernetes.io/ingress.class": ingress_class},
            },
            "spec": {
                "rules": [{"host": host, "http": {"paths": path_specs}}]
            },
        }

        if tls:
            manifest["spec"]["tls"] = tls

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "Ingress manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
