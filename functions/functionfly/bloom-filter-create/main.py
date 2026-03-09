import hashlib


class BloomFilter:
    def __init__(self, size=1000, hash_functions=3):
        self.size = size
        self.hash_functions = hash_functions
        self.bit_array = [0] * size

    def _hashes(self, item):
        """Generate multiple hash values for an item"""
        if not isinstance(item, str):
            item = str(item)

        hashes = []
        # Use different hash functions
        for i in range(self.hash_functions):
            # Use different hash algorithms and combine them
            hash_obj = hashlib.md5(f"{item}{i}".encode())
            hash_val = int(hash_obj.hexdigest(), 16) % self.size
            hashes.append(hash_val)
        return hashes

    def add(self, item):
        hashes = self._hashes(item)
        for hash_val in hashes:
            self.bit_array[hash_val] = 1
        return True

    def contains(self, item):
        hashes = self._hashes(item)
        for hash_val in hashes:
            if self.bit_array[hash_val] == 0:
                return False
        return True

    def clear(self):
        self.bit_array = [0] * self.size

    def get_stats(self):
        set_bits = sum(self.bit_array)
        return {
            "size": self.size,
            "hash_functions": self.hash_functions,
            "set_bits": set_bits,
            "fill_ratio": set_bits / self.size
        }


def handler(event):
    size = event.get("size", 1000)
    hash_functions = event.get("hash_functions", 3)
    items = event.get("items", [])
    operations = event.get("operations", [])

    if not isinstance(size, int) or size <= 0:
        return {"ok": False, "error": "size must be a positive integer"}

    if not isinstance(hash_functions, int) or hash_functions <= 0:
        return {"ok": False, "error": "hash_functions must be a positive integer"}

    if not isinstance(items, list):
        return {"ok": False, "error": "items must be an array"}

    if not isinstance(operations, list):
        return {"ok": False, "error": "operations must be an array"}

    try:
        bloom_filter = BloomFilter(size, hash_functions)

        # Add initial items
        for item in items:
            bloom_filter.add(item)

        results = []
        for op in operations:
            if not isinstance(op, dict):
                return {"ok": False, "error": "each operation must be an object"}

            action = op.get("action")
            if action not in ["add", "contains", "clear", "stats"]:
                return {"ok": False, "error": f"unknown action: {action}"}

            if action == "add":
                item = op.get("item")
                if item is None:
                    return {"ok": False, "error": "add action requires 'item'"}
                success = bloom_filter.add(item)
                results.append({"action": "add", "item": item, "success": success})

            elif action == "contains":
                item = op.get("item")
                if item is None:
                    return {"ok": False, "error": "contains action requires 'item'"}
                result = bloom_filter.contains(item)
                results.append({"action": "contains", "item": item, "result": result})

            elif action == "clear":
                bloom_filter.clear()
                results.append({"action": "clear", "success": True})

            elif action == "stats":
                stats = bloom_filter.get_stats()
                results.append({"action": "stats", "result": stats})

        return {
            "ok": True,
            "result": {
                "stats": bloom_filter.get_stats(),
                "operations": results
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"bloom filter operation failed: {str(e)}"}