from collections import deque


class Queue:
    def __init__(self, items=None):
        self.items = deque(items or [])

    def enqueue(self, item):
        self.items.append(item)

    def dequeue(self):
        if self.is_empty():
            return None
        return self.items.popleft()

    def peek(self):
        if self.is_empty():
            return None
        return self.items[0]

    def is_empty(self):
        return len(self.items) == 0

    def size(self):
        return len(self.items)

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
        queue = Queue(items)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["enqueue", "dequeue", "peek", "size", "is_empty"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "enqueue":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "enqueue action requires 'value'"}
                queue.enqueue(value)
                results.append({"action": "enqueue", "value": value, "success": True})

            elif action == "dequeue":
                result = queue.dequeue()
                results.append({"action": "dequeue", "result": result, "success": result is not None})

            elif action == "peek":
                result = queue.peek()
                results.append({"action": "peek", "result": result})

            elif action == "size":
                results.append({"action": "size", "result": queue.size()})

            elif action == "is_empty":
                results.append({"action": "is_empty", "result": queue.is_empty()})

        return {
            "ok": True,
            "result": {
                "final_queue": queue.to_list(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"queue operation failed: {str(e)}"}