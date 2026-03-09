import json

try:
    import yaml
except ImportError:
    yaml = None


def handler(event):
    """
    Convert YAML to JSON. Requires PyYAML; uses stdlib-only fallback for simple YAML.

    Input:
        - yaml: YAML string to convert (required)

    Returns:
        - ok: True on success
        - json: JSON string
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        data = event.get("yaml", event.get("data", event.get("input", "")))
    else:
        data = event

    if data is None or (isinstance(data, str) and not data.strip()):
        return {"ok": False, "error": "Input 'yaml' is required"}

    yaml_str = str(data).strip()

    if yaml is not None:
        try:
            obj = yaml.safe_load(yaml_str)
            out = json.dumps(obj, indent=None, ensure_ascii=False)
            return {"ok": True, "json": out}
        except yaml.YAMLError as e:
            return {"ok": False, "error": f"Invalid YAML: {e}"}
        except Exception as e:
            return {"ok": False, "error": str(e)}

    # Minimal fallback: only supports key: value lines (no nesting, no lists)
    try:
        obj = _simple_yaml_parse(yaml_str)
        out = json.dumps(obj, ensure_ascii=False)
        return {"ok": True, "json": out}
    except Exception as e:
        return {"ok": False, "error": f"Simple YAML parse failed (install PyYAML for full support): {e}"}


def _simple_yaml_parse(s):
    """Parse very simple YAML: key: value per line. No lists or nested objects."""
    d = {}
    for line in s.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if ":" not in line:
            continue
        k, _, v = line.partition(":")
        k = k.strip()
        v = v.strip()
        if v.startswith('"') and v.endswith('"'):
            v = v[1:-1].replace('\\"', '"')
        elif v.startswith("'") and v.endswith("'"):
            v = v[1:-1].replace("\\'", "'")
        elif v.lower() == "true":
            v = True
        elif v.lower() == "false":
            v = False
        elif v.lower() == "null":
            v = None
        elif v.isdigit():
            v = int(v)
        elif v.replace(".", "").replace("-", "").isdigit():
            try:
                v = float(v)
            except ValueError:
                pass
        d[k] = v
    return d
