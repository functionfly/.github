@flypy.function(
    name="safe-division",
    description="Perform safe division with error handling",
    deterministic=True,
    enable_performance_monitoring=True
)
def safe_division(calculation_data: Dict[str, Any]) -> Dict[str, Any]:
    """Perform division with comprehensive error handling."""

    try:
        numerator = calculation_data.get("numerator", 0)
        denominator = calculation_data.get("denominator", 1)

        # Validate inputs
        if not isinstance(numerator, (int, float)):
            raise ValueError("Numerator must be a number")
        if not isinstance(denominator, (int, float)):
            raise ValueError("Denominator must be a number")
        if denominator == 0:
            raise ZeroDivisionError("Division by zero")

        result = numerator / denominator

        return {
            "result": result,
            "status": "success",
            "error": None
        }

    except ZeroDivisionError:
        return {
            "result": None,
            "status": "error",
            "error": "division_by_zero"
        }
    except ValueError as e:
        return {
            "result": None,
            "status": "error",
            "error": str(e)
        }
    except Exception as e:
        return {
            "result": None,
            "status": "error",
            "error": f"unexpected_error: {str(e)}"
        }
