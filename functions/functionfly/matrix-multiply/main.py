def handler(event):
    matrix_a = event.get("matrix_a")
    matrix_b = event.get("matrix_b")

    if matrix_a is None or matrix_b is None:
        return {"ok": False, "error": "matrix_a and matrix_b are required"}

    if not isinstance(matrix_a, list) or not isinstance(matrix_b, list):
        return {"ok": False, "error": "matrix_a and matrix_b must be arrays"}

    if not matrix_a or not matrix_b:
        return {"ok": False, "error": "matrices cannot be empty"}

    # Check if matrix_a is valid (list of lists)
    if not all(isinstance(row, list) for row in matrix_a):
        return {"ok": False, "error": "matrix_a must be a 2D array"}

    # Check if matrix_b is valid (list of lists)
    if not all(isinstance(row, list) for row in matrix_b):
        return {"ok": False, "error": "matrix_b must be a 2D array"}

    # Get dimensions
    rows_a = len(matrix_a)
    cols_a = len(matrix_a[0]) if matrix_a else 0
    rows_b = len(matrix_b)
    cols_b = len(matrix_b[0]) if matrix_b else 0

    # Check if all rows have same length in matrix_a
    if not all(len(row) == cols_a for row in matrix_a):
        return {"ok": False, "error": "all rows in matrix_a must have the same length"}

    # Check if all rows have same length in matrix_b
    if not all(len(row) == cols_b for row in matrix_b):
        return {"ok": False, "error": "all rows in matrix_b must have the same length"}

    # Check if matrices can be multiplied
    if cols_a != rows_b:
        return {"ok": False, "error": f"cannot multiply matrices: matrix_a columns ({cols_a}) must equal matrix_b rows ({rows_b})"}

    try:
        # Initialize result matrix
        result = [[0 for _ in range(cols_b)] for _ in range(rows_a)]

        # Perform matrix multiplication
        for i in range(rows_a):
            for j in range(cols_b):
                for k in range(cols_a):
                    result[i][j] += matrix_a[i][k] * matrix_b[k][j]

        return {"ok": True, "result": result}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": f"matrices must contain numeric values: {str(e)}"}