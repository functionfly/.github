from collections import OrderedDict


class LRUCache:
    def __init__(self, capacity=10):
        self.capacity = capacity
        self.cache = OrderedDict()

    def get(self, key):
        if key in self.cache:
            # Move to end (most recently used)
            self.cache.move_to_end(key)
            return self.cache[key]
        return None

    def put(self, key, value):
        if key in self.cache:
            # Update existing value and move to end
            self.cache[key] = value
            self.cache.move_to_end(key)
        else:
            # Add new key-value pair
            self.cache[key] = value
            self.cache.move_to_end(key)

            # Remove least recently used if at capacity
            if len(self.cache) > self.capacity:
                self.cache.popitem(last=False)

        return True

    def delete(self, key):
        if key in self.cache:
            del self.cache[key]
            return True
        return False

    def contains(self, key):
        return key in self.cache

    def clear(self):
        self.cache.clear()

    def size(self):
        return len(self.cache)

    def is_empty(self):
        return len(self.cache) == 0

    def keys(self):
        return list(self.cache.keys())

    def values(self):
        return list(self.cache.values())

    def items(self):
        return list(self.cache.items())

    def set_capacity(self, new_capacity):
        if new_capacity < 1:
            return False

        self.capacity = new_capacity
        # Remove items if current size exceeds new capacity
        while len(self.cache) > self.capacity:
            self.cache.popitem(last=False)
        return True

    def get_stats(self):
        return {
            "capacity": self.capacity,
            "size": len(self.cache),
            "utilization": len(self.cache) / self.capacity if self.capacity > 0 else 0
        }


def handler(event):
    capacity = event.get("capacity", 10)
    items = event.get("items", {})
    operations = event.get("operations", [])

    if not isinstance(capacity, int) or capacity <= 0:
        return {"ok": False, "error": "capacity must be a positive integer"}

    if not isinstance(items, dict):
        return {"ok": False, "error": "items must be an object"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        cache = LRUCache(capacity)

        # Add initial items
        for key, value in items.items():
            cache.put(key, value)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["get", "put", "delete", "contains", "clear", "size", "is_empty", "keys", "values", "items", "set_capacity", "stats"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "get":
                key = op.get("key")
                if key is None:
                    return {"ok": False, "error": "get action requires 'key'"}
                result = cache.get(key)
                results.append({"action": "get", "key": key, "result": result})

            elif action == "put":
                key = op.get("key")
                value = op.get("value")
                if key is None:
                    return {"ok": False, "error": "put action requires 'key'"}
                success = cache.put(key, value)
                results.append({"action": "put", "key": key, "value": value, "success": success})

            elif action == "delete":
                key = op.get("key")
                if key is None:
                    return {"ok": False, "error": "delete action requires 'key'"}
                deleted = cache.delete(key)
                results.append({"action": "delete", "key": key, "deleted": deleted})

            elif action == "contains":
                key = op.get("key")
                if key is None:
                    return {"ok": False, "error": "contains action requires 'key'"}
                result = cache.contains(key)
                results.append({"action": "contains", "key": key, "result": result})

            elif action == "clear":
                cache.clear()
                results.append({"action": "clear", "success": True})

            elif action == "size":
                result = cache.size()
                results.append({"action": "size", "result": result})

            elif action == "is_empty":
                result = cache.is_empty()
                results.append({"action": "is_empty", "result": result})

            elif action == "keys":
                result = cache.keys()
                results.append({"action": "keys", "result": result})

            elif action == "values":
                result = cache.values()
                results.append({"action": "values", "result": result})

            elif action == "items":
                result = cache.items()
                results.append({"action": "items", "result": result})

            elif action == "set_capacity":
                new_capacity = op.get("capacity")
                if not isinstance(new_capacity, int) or new_capacity <= 0:
                    return {"ok": False, "error": "set_capacity action requires positive integer 'capacity'"}
                success = cache.set_capacity(new_capacity)
                results.append({"action": "set_capacity", "capacity": new_capacity, "success": success})

            elif action == "stats":
                stats = cache.get_stats()
                results.append({"action": "stats", "result": stats})

        return {
            "ok": True,
            "result": {
                "final_cache": dict(cache.items()),
                "stats": cache.get_stats(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"LRU cache operation failed: {str(e)}"}