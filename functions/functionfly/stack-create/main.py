class Stack:
    def __init__(self, items=None):
        self.items = list(items or [])

    def push(self, item):
        self.items.append(item)

    def pop(self):
        if self.is_empty():
            return None
        return self.items.pop()

    def peek(self):
        if self.is_empty():
            return None
        return self.items[-1]

    def is_empty(self):
        return len(self.items) == 0

    def size(self):
        return len(self.items)

    def to_list(self):
        return self.items.copy()


def handler(event):
    items = event.get("items", [])
    operations = event.get("operations", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        stack = Stack(items)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["push", "pop", "peek", "size", "is_empty"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "push":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "push action requires 'value'"}
                stack.push(value)
                results.append({"action": "push", "value": value, "success": True})

            elif action == "pop":
                result = stack.pop()
                results.append({"action": "pop", "result": result, "success": result is not None})

            elif action == "peek":
                result = stack.peek()
                results.append({"action": "peek", "result": result})

            elif action == "size":
                results.append({"action": "size", "result": stack.size()})

            elif action == "is_empty":
                results.append({"action": "is_empty", "result": stack.is_empty()})

        return {
            "ok": True,
            "result": {
                "final_stack": stack.to_list(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"stack operation failed: {str(e)}"}