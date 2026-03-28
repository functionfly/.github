import yaml


def handler(event):
    """Generate a Kubernetes PersistentVolumeClaim manifest."""
    try:
        name = event.get("name")
        storage = event.get("storage")
        if not name or not storage:
            return {"ok": False, "error": "name and storage are required"}

        access_modes = event.get("access_modes", ["ReadWriteOnce"])
        storage_class = event.get("storage_class")
        namespace = event.get("namespace", "default")

        manifest = {
            "apiVersion": "v1",
            "kind": "PersistentVolumeClaim",
            "metadata": {"name": name, "namespace": namespace},
            "spec": {
                "accessModes": access_modes,
                "resources": {"requests": {"storage": storage}},
            },
        }

        if storage_class:
            manifest["spec"]["storageClassName"] = storage_class

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": "PVC manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
