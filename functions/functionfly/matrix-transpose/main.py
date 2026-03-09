def handler(event):
    matrix = event.get("matrix")

    if matrix is None:
        return {"ok": False, "error": "matrix is required"}

    if not isinstance(matrix, list):
        return {"ok": False, "error": "matrix must be an array"}

    if not matrix:
        return {"ok": True, "result": []}

    # Check if matrix is valid (list of lists)
    if not all(isinstance(row, list) for row in matrix):
        return {"ok": False, "error": "matrix must be a 2D array"}

    # Get dimensions
    rows = len(matrix)
    cols = len(matrix[0]) if matrix else 0

    # Check if all rows have same length
    if not all(len(row) == cols for row in matrix):
        return {"ok": False, "error": "all rows in matrix must have the same length"}

    try:
        # Create transposed matrix
        result = [[matrix[j][i] for j in range(rows)] for i in range(cols)]
        return {"ok": True, "result": result}
    except (IndexError, TypeError):
        return {"ok": False, "error": "invalid matrix structure"}