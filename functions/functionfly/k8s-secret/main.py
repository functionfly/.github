import yaml
import base64


def handler(event):
    """Generate a Kubernetes Secret manifest."""
    try:
        name = event.get("name")
        if not name:
            return {"ok": False, "error": "name is required"}

        secret_type = event.get("type", "Opaque")
        data = event.get("data", {})
        string_data = event.get("string_data", {})
        namespace = event.get("namespace", "default")

        manifest = {
            "apiVersion": "v1",
            "kind": "Secret",
            "metadata": {"name": name, "namespace": namespace},
            "type": secret_type,
        }

        if data:
            manifest["data"] = data
        if string_data:
            # Encode string_data to base64
            encoded = {}
            for k, v in string_data.items():
                encoded[k] = base64.b64encode(str(v).encode()).decode()
            manifest["data"] = {**manifest.get("data", {}), **encoded}

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "Secret manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
