class TrieNode:
    def __init__(self):
        self.children = {}
        self.is_end_of_word = False
        self.count = 0


class Trie:
    def __init__(self):
        self.root = TrieNode()

    def insert(self, word):
        if not isinstance(word, str):
            return False

        node = self.root
        for char in word:
            if char not in node.children:
                node.children[char] = TrieNode()
            node = node.children[char]
            node.count += 1
        node.is_end_of_word = True
        return True

    def search(self, word):
        if not isinstance(word, str):
            return False

        node = self.root
        for char in word:
            if char not in node.children:
                return False
            node = node.children[char]
        return node.is_end_of_word

    def starts_with(self, prefix):
        if not isinstance(prefix, str):
            return False

        node = self.root
        for char in prefix:
            if char not in node.children:
                return False
            node = node.children[char]
        return True

    def get_words_with_prefix(self, prefix):
        if not isinstance(prefix, str):
            return []

        node = self._get_node(prefix)
        if not node:
            return []

        words = []
        self._collect_words(node, prefix, words)
        return words

    def _get_node(self, prefix):
        node = self.root
        for char in prefix:
            if char not in node.children:
                return None
            node = node.children[char]
        return node

    def _collect_words(self, node, current_word, words):
        if node.is_end_of_word:
            words.append(current_word)

        for char, child_node in sorted(node.children.items()):
            self._collect_words(child_node, current_word + char, words)

    def delete(self, word):
        if not isinstance(word, str):
            return False

        return self._delete_helper(self.root, word, 0)

    def _delete_helper(self, node, word, index):
        if index == len(word):
            if not node.is_end_of_word:
                return False
            node.is_end_of_word = False
            return len(node.children) == 0

        char = word[index]
        if char not in node.children:
            return False

        should_delete_child = self._delete_helper(node.children[char], word, index + 1)

        if should_delete_child:
            del node.children[char]
            return len(node.children) == 0 and not node.is_end_of_word
        return False

    def get_prefix_count(self, prefix):
        if not isinstance(prefix, str):
            return 0

        node = self._get_node(prefix)
        return node.count if node else 0

    def is_empty(self):
        return len(self.root.children) == 0

    def clear(self):
        self.root = TrieNode()


def handler(event):
    words = event.get("words", [])
    operations = event.get("operations", [])

    if not isinstance(words, list):
        return {"ok": False, "error": "words must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        trie = Trie()

        # Insert initial words
        for word in words:
            if not isinstance(word, str):
                return {"ok": False, "error": "all words must be strings"}
            trie.insert(word)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["insert", "search", "starts_with", "get_words_with_prefix", "delete", "prefix_count", "is_empty", "clear"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "insert":
                word = op.get("word")
                if not isinstance(word, str):
                    return {"ok": False, "error": "insert action requires string 'word'"}
                success = trie.insert(word)
                results.append({"action": "insert", "word": word, "success": success})

            elif action == "search":
                word = op.get("word")
                if not isinstance(word, str):
                    return {"ok": False, "error": "search action requires string 'word'"}
                result = trie.search(word)
                results.append({"action": "search", "word": word, "result": result})

            elif action == "starts_with":
                prefix = op.get("prefix")
                if not isinstance(prefix, str):
                    return {"ok": False, "error": "starts_with action requires string 'prefix'"}
                result = trie.starts_with(prefix)
                results.append({"action": "starts_with", "prefix": prefix, "result": result})

            elif action == "get_words_with_prefix":
                prefix = op.get("prefix")
                if not isinstance(prefix, str):
                    return {"ok": False, "error": "get_words_with_prefix action requires string 'prefix'"}
                result = trie.get_words_with_prefix(prefix)
                results.append({"action": "get_words_with_prefix", "prefix": prefix, "result": result})

            elif action == "delete":
                word = op.get("word")
                if not isinstance(word, str):
                    return {"ok": False, "error": "delete action requires string 'word'"}
                deleted = trie.delete(word)
                results.append({"action": "delete", "word": word, "deleted": deleted})

            elif action == "prefix_count":
                prefix = op.get("prefix")
                if not isinstance(prefix, str):
                    return {"ok": False, "error": "prefix_count action requires string 'prefix'"}
                result = trie.get_prefix_count(prefix)
                results.append({"action": "prefix_count", "prefix": prefix, "result": result})

            elif action == "is_empty":
                result = trie.is_empty()
                results.append({"action": "is_empty", "result": result})

            elif action == "clear":
                trie.clear()
                results.append({"action": "clear", "success": True})

        return {
            "ok": True,
            "result": {
                "word_count": trie.get_prefix_count(""),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"trie operation failed: {str(e)}"}