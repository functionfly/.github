def _determinant(matrix):
    """Calculate determinant using Gaussian elimination"""
    n = len(matrix)
    if n == 1:
        return matrix[0][0]
    if n == 2:
        return matrix[0][0] * matrix[1][1] - matrix[0][1] * matrix[1][0]

    # Create a copy of the matrix
    A = [row[:] for row in matrix]

    det = 1
    # Forward elimination
    for i in range(n):
        # Find pivot
        pivot_row = i
        for j in range(i + 1, n):
            if abs(A[j][i]) > abs(A[pivot_row][i]):
                pivot_row = j

        # Swap rows if needed
        if pivot_row != i:
            A[i], A[pivot_row] = A[pivot_row], A[i]
            det *= -1

        # Check for singular matrix
        if abs(A[i][i]) < 1e-10:
            return 0

        # Eliminate
        for j in range(i + 1, n):
            factor = A[j][i] / A[i][i]
            for k in range(i, n):
                A[j][k] -= factor * A[i][k]

    # Calculate determinant from diagonal
    for i in range(n):
        det *= A[i][i]

    return det


def handler(event):
    matrix = event.get("matrix")

    if matrix is None:
        return {"ok": False, "error": "matrix is required"}

    if not isinstance(matrix, list):
        return {"ok": False, "error": "matrix must be an array"}

    if not matrix:
        return {"ok": False, "error": "matrix cannot be empty"}

    # Check if matrix is valid (list of lists)
    if not all(isinstance(row, list) for row in matrix):
        return {"ok": False, "error": "matrix must be a 2D array"}

    # Get dimensions
    rows = len(matrix)
    cols = len(matrix[0]) if matrix else 0

    # Check if matrix is square
    if rows != cols:
        return {"ok": False, "error": "matrix must be square (same number of rows and columns)"}

    # Check if all rows have same length
    if not all(len(row) == cols for row in matrix):
        return {"ok": False, "error": "all rows in matrix must have the same length"}

    try:
        result = _determinant(matrix)
        return {"ok": True, "result": result}
    except (TypeError, ValueError, ZeroDivisionError) as e:
        return {"ok": False, "error": f"failed to calculate determinant: {str(e)}"}