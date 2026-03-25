"""
Example function using python-microvm runtime (Enterprise tier).
Requires NumPy - runs in Firecracker MicroVM with full CPython.
"""


def handler(data):
    import json

    # Parse input
    if isinstance(data, str):
        data = json.loads(data) if data else {}

    # NumPy is available in python-microvm
    import numpy as np

    arr = np.array(data.get("values", [1, 2, 3, 4, 5]))
    return {
        "mean": float(np.mean(arr)),
        "std": float(np.std(arr)),
        "sum": int(np.sum(arr)),
    }
