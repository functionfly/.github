import math

def _mean_vec(data):
    n = len(data)
    dim = len(data[0])
    return [sum(data[i][j] for i in range(n)) / n for j in range(dim)]

def _center(data, mean):
    return [[x - m for x, m in zip(row, mean)] for row in data]

def _covariance_matrix(centered):
    n = len(centered)
    dim = len(centered[0])
    cov = [[0.0] * dim for _ in range(dim)]
    for i in range(dim):
        for j in range(dim):
            cov[i][j] = sum(centered[k][i] * centered[k][j] for k in range(n)) / (n - 1)
    return cov

def _power_iteration(matrix, num_iter=50):
    """Find dominant eigenvector using power iteration."""
    n = len(matrix)
    vec = [1.0 / math.sqrt(n)] * n
    for _ in range(num_iter):
        new_vec = [sum(matrix[i][j] * vec[j] for j in range(n)) for i in range(n)]
        norm = math.sqrt(sum(x * x for x in new_vec)) or 1e-9
        vec = [x / norm for x in new_vec]
    eigenvalue = sum(sum(matrix[i][j] * vec[j] for j in range(n)) * vec[i] for i in range(n))
    return vec, eigenvalue

def _deflate(matrix, eigenvec, eigenvalue):
    """Deflate matrix to find next eigenvector."""
    n = len(matrix)
    return [[matrix[i][j] - eigenvalue * eigenvec[i] * eigenvec[j] for j in range(n)] for i in range(n)]

def _project(data, components):
    """Project data onto principal components."""
    return [[sum(row[j] * comp[j] for j in range(len(row))) for comp in components] for row in data]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, list):
        return {"ok": False, "error": "data (array of arrays) is required"}
    try:
        n_components = int(event.get("n_components", 2))
        processed = [[float(x) for x in row] for row in data if isinstance(row, list)]
        if len(processed) < 2:
            return {"ok": False, "error": "data must contain at least 2 samples"}
        dim = len(processed[0])
        n_components = min(n_components, dim, len(processed))
        mean = _mean_vec(processed)
        centered = _center(processed, mean)
        cov = _covariance_matrix(centered)
        # Extract principal components
        components = []
        eigenvalues = []
        current_cov = [row[:] for row in cov]
        for _ in range(n_components):
            eigvec, eigval = _power_iteration(current_cov)
            components.append(eigvec)
            eigenvalues.append(round(eigval, 6))
            current_cov = _deflate(current_cov, eigvec, eigval)
        transformed = _project(centered, components)
        total_var = sum(eigenvalues) or 1e-9
        explained_variance_ratio = [round(ev / total_var, 4) for ev in eigenvalues]
        return {
            "ok": True,
            "result": transformed,
            "transformed": [[round(x, 6) for x in row] for row in transformed],
            "components": [[round(x, 6) for x in comp] for comp in components],
            "eigenvalues": eigenvalues,
            "explained_variance_ratio": explained_variance_ratio,
            "cumulative_variance": round(sum(explained_variance_ratio), 4),
            "n_components": n_components,
            "original_dimensions": dim,
            "n_samples": len(processed)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
