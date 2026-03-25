import re


def _get_path(obj, path):
    """Simple JSONPath implementation supporting $.key, $[0], $.*, $.key.sub"""
    if path == "$":
        return [obj]
    # Remove leading $
    if path.startswith("$"):
        path = path[1:]
    
    def traverse(current, parts):
        if not parts:
            return [current]
        part = parts[0]
        rest = parts[1:]
        results = []
        
        if part == "*":
            if isinstance(current, dict):
                for v in current.values():
                    results.extend(traverse(v, rest))
            elif isinstance(current, list):
                for v in current:
                    results.extend(traverse(v, rest))
        elif part.startswith("[") and part.endswith("]"):
            inner = part[1:-1]
            if inner == "*":
                if isinstance(current, list):
                    for v in current:
                        results.extend(traverse(v, rest))
            elif re.match(r'^-?\d+$', inner):
                idx = int(inner)
                if isinstance(current, list) and -len(current) <= idx < len(current):
                    results.extend(traverse(current[idx], rest))
            elif inner.startswith("'") or inner.startswith('"'):
                key = inner.strip("'\"")
                if isinstance(current, dict) and key in current:
                    results.extend(traverse(current[key], rest))
        elif isinstance(current, dict) and part in current:
            results.extend(traverse(current[part], rest))
        return results
    
    # Split path into parts
    parts = re.split(r'(?<!\\)\.(?!\d)', path.lstrip("."))
    # Further split on brackets
    all_parts = []
    for p in parts:
        if not p:
            continue
        tokens = re.findall(r'\[[^\]]+\]|[^\[]+', p)
        for t in tokens:
            if t:
                all_parts.append(t)
    
    return traverse(obj, all_parts)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    query = event.get("query")
    if data is None:
        return {"ok": False, "error": "data is required"}
    if not query:
        return {"ok": False, "error": "query is required (JSONPath expression, e.g. '$.users[0].name')"}
    try:
        try:
            from jsonpath_ng import parse as jparse
            expr = jparse(str(query))
            matches = [m.value for m in expr.find(data)]
            return {"ok": True, "result": matches, "count": len(matches)}
        except ImportError:
            matches = _get_path(data, str(query))
            return {"ok": True, "result": matches, "count": len(matches),
                    "note": "jsonpath-ng not installed; using basic JSONPath"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
