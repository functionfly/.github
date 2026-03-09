class CustomSet:
    def __init__(self, items=None):
        self.items = set(items or [])

    def add(self, item):
        self.items.add(item)

    def remove(self, item):
        self.items.discard(item)

    def contains(self, item):
        return item in self.items

    def union(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        self.items = self.items.union(other)
        return True

    def intersection(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        self.items = self.items.intersection(other)
        return True

    def difference(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        self.items = self.items.difference(other)
        return True

    def symmetric_difference(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        self.items = self.items.symmetric_difference(other)
        return True

    def is_subset(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        return self.items.issubset(other)

    def is_superset(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        return self.items.issuperset(other)

    def is_disjoint(self, other_set):
        if not isinstance(other_set, (set, list, CustomSet)):
            return False
        other = set(other_set) if isinstance(other_set, list) else (other_set.items if isinstance(other_set, CustomSet) else other_set)
        return self.items.isdisjoint(other)

    def clear(self):
        self.items.clear()

    def size(self):
        return len(self.items)

    def is_empty(self):
        return len(self.items) == 0

    def to_list(self):
        return list(self.items)


def handler(event):
    items = event.get("items", [])
    operations = event.get("operations", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        custom_set = CustomSet(items)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["add", "remove", "contains", "union", "intersection", "difference", "symmetric_difference", "is_subset", "is_superset", "is_disjoint", "clear", "size", "is_empty"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "add":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "add action requires 'value'"}
                custom_set.add(value)
                results.append({"action": "add", "value": value, "success": True})

            elif action == "remove":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "remove action requires 'value'"}
                old_size = custom_set.size()
                custom_set.remove(value)
                removed = custom_set.size() < old_size
                results.append({"action": "remove", "value": value, "removed": removed})

            elif action == "contains":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "contains action requires 'value'"}
                result = custom_set.contains(value)
                results.append({"action": "contains", "value": value, "result": result})

            elif action in ["union", "intersection", "difference", "symmetric_difference"]:
                other_set = op.get("set")
                if other_set is None:
                    return {"ok": False, "error": f"{action} action requires 'set'"}
                method = getattr(custom_set, action)
                success = method(other_set)
                if not success:
                    return {"ok": False, "error": f"invalid set provided for {action}"}
                results.append({"action": action, "set": other_set, "success": True})

            elif action in ["is_subset", "is_superset", "is_disjoint"]:
                other_set = op.get("set")
                if other_set is None:
                    return {"ok": False, "error": f"{action} action requires 'set'"}
                method = getattr(custom_set, action)
                result = method(other_set)
                if result is False and not isinstance(result, bool):
                    return {"ok": False, "error": f"invalid set provided for {action}"}
                results.append({"action": action, "set": other_set, "result": result})

            elif action == "clear":
                custom_set.clear()
                results.append({"action": "clear", "success": True})

            elif action == "size":
                result = custom_set.size()
                results.append({"action": "size", "result": result})

            elif action == "is_empty":
                result = custom_set.is_empty()
                results.append({"action": "is_empty", "result": result})

        return {
            "ok": True,
            "result": {
                "final_set": sorted(custom_set.to_list()),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"set operation failed: {str(e)}"}