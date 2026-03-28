import math
import random

def _distance(a, b):
    return math.sqrt(sum((x - y) ** 2 for x, y in zip(a, b)))

def _mean_point(points):
    if not points:
        return None
    dim = len(points[0])
    return [sum(p[i] for p in points) / len(points) for i in range(dim)]

def _kmeans(data, k, max_iter=100, seed=42):
    if len(data) < k:
        k = len(data)
    # Initialize centroids using deterministic selection
    step = max(1, len(data) // k)
    centroids = [list(data[i * step]) for i in range(k)]
    labels = [0] * len(data)
    for _ in range(max_iter):
        # Assign
        new_labels = []
        for point in data:
            dists = [_distance(point, c) for c in centroids]
            new_labels.append(dists.index(min(dists)))
        if new_labels == labels:
            break
        labels = new_labels
        # Update centroids
        for j in range(k):
            cluster_points = [data[i] for i, l in enumerate(labels) if l == j]
            if cluster_points:
                centroids[j] = _mean_point(cluster_points)
    return labels, centroids

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array of arrays) is required"}
    try:
        k = int(event.get("k", 3))
        max_iter = int(event.get("max_iterations", 100))
        # Normalize data to lists of floats
        processed = []
        for point in data:
            if isinstance(point, (int, float)):
                processed.append([float(point)])
            elif isinstance(point, list):
                processed.append([float(x) for x in point])
            else:
                continue
        if len(processed) < 2:
            return {"ok": False, "error": "data must contain at least 2 points"}
        k = min(k, len(processed))
        labels, centroids = _kmeans(processed, k, max_iter)
        # Build cluster info
        clusters = {}
        for i, label in enumerate(labels):
            if label not in clusters:
                clusters[label] = {"id": label, "centroid": centroids[label], "points": [], "size": 0}
            clusters[label]["points"].append(i)
            clusters[label]["size"] += 1
        # Compute inertia
        inertia = sum(_distance(processed[i], centroids[labels[i]]) ** 2 for i in range(len(processed)))
        return {
            "ok": True,
            "result": {"labels": labels, "centroids": centroids},
            "labels": labels,
            "centroids": [[round(x, 6) for x in c] for c in centroids],
            "clusters": list(clusters.values()),
            "k": k,
            "inertia": round(inertia, 6),
            "iterations": max_iter
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
