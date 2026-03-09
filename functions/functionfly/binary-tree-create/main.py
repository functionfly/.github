class TreeNode:
    def __init__(self, value):
        self.value = value
        self.left = None
        self.right = None


class BinaryTree:
    def __init__(self, root=None):
        self.root = TreeNode(root) if root is not None else None

    def insert(self, value):
        if not self.root:
            self.root = TreeNode(value)
            return

        # Simple insertion: left if < current, right if >=
        current = self.root
        while True:
            if value < current.value:
                if current.left is None:
                    current.left = TreeNode(value)
                    break
                current = current.left
            else:
                if current.right is None:
                    current.right = TreeNode(value)
                    break
                current = current.right

    def inorder_traversal(self):
        result = []
        self._inorder_helper(self.root, result)
        return result

    def _inorder_helper(self, node, result):
        if node:
            self._inorder_helper(node.left, result)
            result.append(node.value)
            self._inorder_helper(node.right, result)

    def preorder_traversal(self):
        result = []
        self._preorder_helper(self.root, result)
        return result

    def _preorder_helper(self, node, result):
        if node:
            result.append(node.value)
            self._preorder_helper(node.left, result)
            self._preorder_helper(node.right, result)

    def postorder_traversal(self):
        result = []
        self._postorder_helper(self.root, result)
        return result

    def _postorder_helper(self, node, result):
        if node:
            self._postorder_helper(node.left, result)
            self._postorder_helper(node.right, result)
            result.append(node.value)

    def find(self, value):
        current = self.root
        while current:
            if value == current.value:
                return True
            elif value < current.value:
                current = current.left
            else:
                current = current.right
        return False

    def get_min(self):
        if not self.root:
            return None
        current = self.root
        while current.left:
            current = current.left
        return current.value

    def get_max(self):
        if not self.root:
            return None
        current = self.root
        while current.right:
            current = current.right
        return current.value

    def is_empty(self):
        return self.root is None

    def get_height(self):
        return self._height_helper(self.root)

    def _height_helper(self, node):
        if not node:
            return 0
        return 1 + max(self._height_helper(node.left), self._height_helper(node.right))


def handler(event):
    root = event.get("root")
    operations = event.get("operations", [])

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        tree = BinaryTree(root)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["insert", "find", "inorder", "preorder", "postorder", "min", "max", "height", "is_empty"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "insert":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "insert action requires 'value'"}
                tree.insert(value)
                results.append({"action": "insert", "value": value, "success": True})

            elif action == "find":
                value = op.get("value")
                if value is None:
                    return {"ok": False, "error": "find action requires 'value'"}
                result = tree.find(value)
                results.append({"action": "find", "value": value, "result": result})

            elif action == "inorder":
                result = tree.inorder_traversal()
                results.append({"action": "inorder", "result": result})

            elif action == "preorder":
                result = tree.preorder_traversal()
                results.append({"action": "preorder", "result": result})

            elif action == "postorder":
                result = tree.postorder_traversal()
                results.append({"action": "postorder", "result": result})

            elif action == "min":
                result = tree.get_min()
                results.append({"action": "min", "result": result})

            elif action == "max":
                result = tree.get_max()
                results.append({"action": "max", "result": result})

            elif action == "height":
                result = tree.get_height()
                results.append({"action": "height", "result": result})

            elif action == "is_empty":
                result = tree.is_empty()
                results.append({"action": "is_empty", "result": result})

        return {
            "ok": True,
            "result": {
                "inorder_traversal": tree.inorder_traversal(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"binary tree operation failed: {str(e)}"}