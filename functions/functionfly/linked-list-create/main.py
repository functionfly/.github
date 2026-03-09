class Node:
    def __init__(self, value):
        self.value = value
        self.next = None


class LinkedList:
    def __init__(self, items=None):
        self.head = None
        self.size = 0
        if items:
            for item in reversed(items):
                self.insert_front(item)

    def insert_front(self, value):
        new_node = Node(value)
        new_node.next = self.head
        self.head = new_node
        self.size += 1

    def insert_back(self, value):
        new_node = Node(value)
        if not self.head:
            self.head = new_node
        else:
            current = self.head
            while current.next:
                current = current.next
            current.next = new_node
        self.size += 1

    def remove_front(self):
        if not self.head:
            return None
        value = self.head.value
        self.head = self.head.next
        self.size -= 1
        return value

    def remove_back(self):
        if not self.head:
            return None
        if not self.head.next:
            value = self.head.value
            self.head = None
            self.size -= 1
            return value

        current = self.head
        while current.next.next:
            current = current.next
        value = current.next.value
        current.next = None
        self.size -= 1
        return value

    def get_front(self):
        return self.head.value if self.head else None

    def get_back(self):
        if not self.head:
            return None
        current = self.head
        while current.next:
            current = current.next
        return current.value

    def is_empty(self):
        return self.size == 0

    def get_size(self):
        return self.size

    def to_list(self):
        result = []
        current = self.head
        while current:
            result.append(current.value)
            current = current.next
        return result


def handler(event):
    items = event.get("items", [])
    operations = event.get("operations", [])

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        linked_list = LinkedList(items)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["insert_front", "insert_back", "remove_front", "remove_back", "get_front", "get_back", "size", "is_empty"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "insert_front":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "insert_front action requires 'value'"}
                linked_list.insert_front(value)
                results.append({"action": "insert_front", "value": value, "success": True})

            elif action == "insert_back":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "insert_back action requires 'value'"}
                linked_list.insert_back(value)
                results.append({"action": "insert_back", "value": value, "success": True})

            elif action == "remove_front":
                result = linked_list.remove_front()
                results.append({"action": "remove_front", "result": result, "success": result is not None})

            elif action == "remove_back":
                result = linked_list.remove_back()
                results.append({"action": "remove_back", "result": result, "success": result is not None})

            elif action == "get_front":
                result = linked_list.get_front()
                results.append({"action": "get_front", "result": result})

            elif action == "get_back":
                result = linked_list.get_back()
                results.append({"action": "get_back", "result": result})

            elif action == "size":
                results.append({"action": "size", "result": linked_list.get_size()})

            elif action == "is_empty":
                results.append({"action": "is_empty", "result": linked_list.is_empty()})

        return {
            "ok": True,
            "result": {
                "final_list": linked_list.to_list(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"linked list operation failed: {str(e)}"}