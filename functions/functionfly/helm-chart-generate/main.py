import yaml


def handler(event):
    """Generate a Helm chart structure."""
    try:
        name = event.get("name")
        if not name:
            return {"ok": False, "error": "name is required"}

        version = event.get("version", "0.1.0")
        app_version = event.get("app_version", "1.0.0")
        description = event.get("description", f"A Helm chart for {name}")
        image = event.get("image", f"{name}:latest")
        replicas = event.get("replicas", 1)
        service_port = event.get("service_port", 80)

        chart_yaml = yaml.dump({
            "apiVersion": "v2",
            "name": name,
            "description": description,
            "type": "application",
            "version": version,
            "appVersion": app_version,
        }, default_flow_style=False)

        values_yaml = yaml.dump({
            "replicaCount": replicas,
            "image": {"repository": image.split(":")[0], "tag": image.split(":")[-1] if ":" in image else "latest", "pullPolicy": "IfNotPresent"},
            "service": {"type": "ClusterIP", "port": service_port},
            "resources": {},
            "nodeSelector": {},
            "tolerations": [],
            "affinity": {},
        }, default_flow_style=False)

        deployment_template = """apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "{name}.fullname" . }}
  labels:
    {{- include "{name}.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "{name}.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "{name}.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - containerPort: {port}
""".format(name=name, port=service_port)

        helpers_template = """{{- define "{name}.fullname" -}}
{{- .Release.Name }}-{{- .Chart.Name }}
{{- end }}

{{- define "{name}.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{ include "{name}.selectorLabels" . }}
{{- end }}

{{- define "{name}.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
""".format(name=name)

        chart = {
            "Chart.yaml": chart_yaml,
            "values.yaml": values_yaml,
            "templates/deployment.yaml": deployment_template,
            "templates/_helpers.tpl": helpers_template,
        }

        return {"ok": True, "result": f"Helm chart '{name}' generated", "chart": chart}
    except Exception as e:
        return {"ok": False, "error": str(e)}
