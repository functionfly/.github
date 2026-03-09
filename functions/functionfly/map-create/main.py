class CustomMap:
    def __init__(self, items=None):
        self.items = dict(items or {})

    def set(self, key, value):
        self.items[key] = value

    def get(self, key, default=None):
        return self.items.get(key, default)

    def delete(self, key):
        return self.items.pop(key, None) is not None

    def has(self, key):
        return key in self.items

    def keys(self):
        return list(self.items.keys())

    def values(self):
        return list(self.items.values())

    def entries(self):
        return list(self.items.items())

    def clear(self):
        self.items.clear()

    def size(self):
        return len(self.items)

    def is_empty(self):
        return len(self.items) == 0

    def merge(self, other_map):
        if isinstance(other_map, dict):
            self.items.update(other_map)
            return True
        elif isinstance(other_map, list) and all(isinstance(item, list) and len(item) == 2 for item in other_map):
            for key, value in other_map:
                self.items[key] = value
            return True
        return False

    def to_dict(self):
        return dict(self.items)


def handler(event):
    items = event.get("items", {})
    operations = event.get("operations", [])

    if not isinstance(items, dict):
        return {"ok": False, "error": "items must be an object"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        custom_map = CustomMap(items)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["set", "get", "delete", "has", "keys", "values", "entries", "clear", "size", "is_empty", "merge"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "set":
                key = op.get("key")
                value = op.get("value")
                if key is None:
                    return {"ok": False, "error": "set action requires 'key'"}
                custom_map.set(key, value)
                results.append({"action": "set", "key": key, "value": value, "success": True})

            elif action == "get":
                key = op.get("key")
                default = op.get("default")
                if key is None:
                    return {"ok": False, "error": "get action requires 'key'"}
                result = custom_map.get(key, default)
                results.append({"action": "get", "key": key, "result": result})

            elif action == "delete":
                key = op.get("key")
                if key is None:
                    return {"ok": False, "error": "delete action requires 'key'"}
                deleted = custom_map.delete(key)
                results.append({"action": "delete", "key": key, "deleted": deleted})

            elif action == "has":
                key = op.get("key")
                if key is None:
                    return {"ok": False, "error": "has action requires 'key'"}
                result = custom_map.has(key)
                results.append({"action": "has", "key": key, "result": result})

            elif action == "keys":
                result = custom_map.keys()
                results.append({"action": "keys", "result": result})

            elif action == "values":
                result = custom_map.values()
                results.append({"action": "values", "result": result})

            elif action == "entries":
                result = custom_map.entries()
                results.append({"action": "entries", "result": result})

            elif action == "clear":
                custom_map.clear()
                results.append({"action": "clear", "success": True})

            elif action == "size":
                result = custom_map.size()
                results.append({"action": "size", "result": result})

            elif action == "is_empty":
                result = custom_map.is_empty()
                results.append({"action": "is_empty", "result": result})

            elif action == "merge":
                other_map = op.get("map")
                if other_map is None:
                    return {"ok": False, "error": "merge action requires 'map'"}
                success = custom_map.merge(other_map)
                if not success:
                    return {"ok": False, "error": "invalid map format for merge"}
                results.append({"action": "merge", "map": other_map, "success": True})

        return {
            "ok": True,
            "result": {
                "final_map": custom_map.to_dict(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"map operation failed: {str(e)}"}