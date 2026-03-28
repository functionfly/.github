import yaml


API_VERSIONS = {
    "Deployment": "apps/v1",
    "StatefulSet": "apps/v1",
    "DaemonSet": "apps/v1",
    "ReplicaSet": "apps/v1",
    "Service": "v1",
    "ConfigMap": "v1",
    "Secret": "v1",
    "PersistentVolumeClaim": "v1",
    "PersistentVolume": "v1",
    "Namespace": "v1",
    "ServiceAccount": "v1",
    "Ingress": "networking.k8s.io/v1",
    "NetworkPolicy": "networking.k8s.io/v1",
    "HorizontalPodAutoscaler": "autoscaling/v2",
    "CronJob": "batch/v1",
    "Job": "batch/v1",
    "Role": "rbac.authorization.k8s.io/v1",
    "RoleBinding": "rbac.authorization.k8s.io/v1",
    "ClusterRole": "rbac.authorization.k8s.io/v1",
    "ClusterRoleBinding": "rbac.authorization.k8s.io/v1",
}


def handler(event):
    """Generate Kubernetes manifest YAML files for any resource type."""
    try:
        resource_type = event.get("resource_type")
        name = event.get("name")
        if not resource_type or not name:
            return {"ok": False, "error": "resource_type and name are required"}

        api_version = API_VERSIONS.get(resource_type, "v1")
        namespace = event.get("namespace", "default")
        labels = event.get("labels", {"app": name})
        annotations = event.get("annotations", {})
        spec = event.get("spec", {})

        manifest = {
            "apiVersion": api_version,
            "kind": resource_type,
            "metadata": {
                "name": name,
                "namespace": namespace,
                "labels": labels,
            },
        }
        if annotations:
            manifest["metadata"]["annotations"] = annotations
        if spec:
            manifest["spec"] = spec

        yaml_str = yaml.dump(manifest, default_flow_style=False, sort_keys=False)
        return {"ok": True, "result": f"{resource_type} manifest generated", "yaml": yaml_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
