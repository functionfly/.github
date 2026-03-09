import heapq


class MaxHeap:
    def __init__(self, items=None):
        # Use negative values to simulate max heap with min heap
        self.items = [-x for x in (items or [])]
        heapq.heapify(self.items)

    def push(self, item):
        heapq.heappush(self.items, -item)

    def pop(self):
        if self.is_empty():
            return None
        return -heapq.heappop(self.items)

    def peek(self):
        if self.is_empty():
            return None
        return -self.items[0]

    def is_empty(self):
        return len(self.items) == 0

    def size(self):
        return len(self.items)

    def to_list(self):
        # Return sorted list (largest to smallest)
        return [-x for x in sorted(self.items)]


def handler(event):
    items = event.get("items", [])
    operations = event.get("operations", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        heap = MaxHeap(items)

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
                heap.push(value)
                results.append({"action": "push", "value": value, "success": True})

            elif action == "pop":
                result = heap.pop()
                results.append({"action": "pop", "result": result, "success": result is not None})

            elif action == "peek":
                result = heap.peek()
                results.append({"action": "peek", "result": result})

            elif action == "size":
                results.append({"action": "size", "result": heap.size()})

            elif action == "is_empty":
                results.append({"action": "is_empty", "result": heap.is_empty()})

        return {
            "ok": True,
            "result": {
                "final_heap": heap.to_list(),
                "operations": results
            }
        }

    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"heap operation failed: {str(e)}"}