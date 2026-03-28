import yaml


def handler(event):
    """Generate a Kubernetes Service manifest."""
    try:
        name = event.get("name")
        ports = event.get("ports")
        if not name or not ports:
            return {"ok": False, "error": "name and ports are required"}

        svc_type = event.get("type", "ClusterIP")
        selector = event.get("selector", {"app": name})
        namespace = event.get("namespace", "default")

        port_specs = []
        for p in ports:
            spec = {"port": p.get("port", 80)}
            if p.get("targetPort"):
                spec["targetPort"] = p["targetPort"]
            if p.get("protocol"):
                spec["protocol"] = p["protocol"]
            if p.get("nodePort") and svc_type == "NodePort":
                spec["nodePort"] = p["nodePort"]
            port_specs.append(spec)

        manifest = {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {"name": name, "namespace": namespace},
            "spec": {
                "type": svc_type,
                "selector": selector,
                "ports": port_specs,
            },
        }

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "Service manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
