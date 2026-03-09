from collections import deque


class CustomDeque:
    def __init__(self, items=None, maxlen=None):
        self.maxlen = maxlen
        self.items = deque(items or [], maxlen=maxlen)

    def append(self, item):
        self.items.append(item)

    def appendleft(self, item):
        self.items.appendleft(item)

    def pop(self):
        if self.is_empty():
            return None
        return self.items.pop()

    def popleft(self):
        if self.is_empty():
            return None
        return self.items.popleft()

    def peek(self):
        if self.is_empty():
            return None
        return self.items[-1]

    def peekleft(self):
        if self.is_empty():
            return None
        return self.items[0]

    def extend(self, items):
        if isinstance(items, list):
            self.items.extend(items)
        elif hasattr(items, '__iter__'):
            self.items.extend(items)

    def extendleft(self, items):
        if isinstance(items, list):
            self.items.extendleft(items)
        elif hasattr(items, '__iter__'):
            self.items.extendleft(items)

    def rotate(self, n=1):
        self.items.rotate(n)

    def clear(self):
        self.items.clear()

    def size(self):
        return len(self.items)

    def is_empty(self):
        return len(self.items) == 0

    def is_full(self):
        return self.maxlen is not None and len(self.items) == self.maxlen

    def to_list(self):
        return list(self.items)

    def reverse(self):
        self.items.reverse()


def handler(event):
    items = event.get("items", [])
    maxlen = event.get("maxlen")
    operations = event.get("operations", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if maxlen is not None and (not isinstance(maxlen, int) or maxlen <= 0):
        return {"ok": False, "error": "maxlen must be a positive integer or null"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        custom_deque = CustomDeque(items, maxlen)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["append", "appendleft", "pop", "popleft", "peek", "peekleft", "extend", "extendleft", "rotate", "clear", "size", "is_empty", "is_full", "reverse"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "append":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "append action requires 'value'"}
                custom_deque.append(value)
                results.append({"action": "append", "value": value, "success": True})

            elif action == "appendleft":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "appendleft action requires 'value'"}
                custom_deque.appendleft(value)
                results.append({"action": "appendleft", "value": value, "success": True})

            elif action == "pop":
                result = custom_deque.pop()
                results.append({"action": "pop", "result": result, "success": result is not None})

            elif action == "popleft":
                result = custom_deque.popleft()
                results.append({"action": "popleft", "result": result, "success": result is not None})

            elif action == "peek":
                result = custom_deque.peek()
                results.append({"action": "peek", "result": result})

            elif action == "peekleft":
                result = custom_deque.peekleft()
                results.append({"action": "peekleft", "result": result})

            elif action == "extend":
                values = op.get("values", [])
                if not isinstance(values, list):
                    return {"ok": False, "error": "extend action requires 'values' array"}
                custom_deque.extend(values)
                results.append({"action": "extend", "values": values, "success": True})

            elif action == "extendleft":
                values = op.get("values", [])
                if not isinstance(values, list):
                    return {"ok": False, "error": "extendleft action requires 'values' array"}
                custom_deque.extendleft(values)
                results.append({"action": "extendleft", "values": values, "success": True})

            elif action == "rotate":
                n = op.get("n", 1)
                if not isinstance(n, int):
                    n = 1
                custom_deque.rotate(n)
                results.append({"action": "rotate", "n": n, "success": True})

            elif action == "clear":
                custom_deque.clear()
                results.append({"action": "clear", "success": True})

            elif action == "size":
                result = custom_deque.size()
                results.append({"action": "size", "result": result})

            elif action == "is_empty":
                result = custom_deque.is_empty()
                results.append({"action": "is_empty", "result": result})

            elif action == "is_full":
                result = custom_deque.is_full()
                results.append({"action": "is_full", "result": result})

            elif action == "reverse":
                custom_deque.reverse()
                results.append({"action": "reverse", "success": True})

        return {
            "ok": True,
            "result": {
                "final_deque": custom_deque.to_list(),
                "maxlen": custom_deque.maxlen,
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"deque operation failed: {str(e)}"}