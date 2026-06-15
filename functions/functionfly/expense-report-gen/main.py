"""Expense Report Generator - Generate formatted expense reports."""
import re
from datetime import datetime
from typing import Any
import hashlib


VALID_CATEGORIES = [
    "travel", "meals", "accommodation", "transportation", "office supplies",
    "equipment", "software", "subscriptions", "client entertainment",
    "professional development", "communication", "utilities", "other"
]


def validate_expense(expense: dict, index: int) -> str:
    """Validate a single expense entry."""
    if "date" not in expense:
        return f"Expense {index + 1}: date is required"
    if "category" not in expense:
        return f"Expense {index + 1}: category is required"
    if "amount" not in expense:
        return f"Expense {index + 1}: amount is required"

    try:
        datetime.strptime(str(expense["date"]), "%Y-%m-%d")
    except ValueError:
        return f"Expense {index + 1}: date must be in YYYY-MM-DD format"

    category = expense["category"].lower()
    if category not in VALID_CATEGORIES:
        return f"Expense {index + 1}: category must be one of: {', '.join(VALID_CATEGORIES)}"

    try:
        amount = float(expense["amount"])
        if amount <= 0:
            return f"Expense {index + 1}: amount must be positive"
        if amount > 1000000:
            return f"Expense {index + 1}: amount seems unreasonably large"
    except (ValueError, TypeError):
        return f"Expense {index + 1}: amount must be a valid number"

    return None


def handler(event: dict) -> dict:
    """Generate an expense report."""
    try:
        employee_name = event.get("employee_name")
        expenses = event.get("expenses", [])
        start_date = event.get("start_date")
        end_date = event.get("end_date")

        if not employee_name:
            return {"ok": False, "error": "employee_name is required"}
        if not expenses or len(expenses) == 0:
            return {"ok": False, "error": "expenses list is required and must not be empty"}
        if not start_date:
            return {"ok": False, "error": "start_date is required"}
        if not end_date:
            return {"ok": False, "error": "end_date is required"}

        for i, expense in enumerate(expenses):
            if not isinstance(expense, dict):
                return {"ok": False, "error": f"expense at index {i} must be an object"}

        for i, expense in enumerate(expenses):
            error = validate_expense(expense, i)
            if error:
                return {"ok": False, "error": error}

        try:
            parsed_start = datetime.strptime(start_date, "%Y-%m-%d")
            parsed_end = datetime.strptime(end_date, "%Y-%m-%d")
            if parsed_end < parsed_start:
                return {"ok": False, "error": "end_date must be after start_date"}
        except ValueError:
            return {"ok": False, "error": "dates must be in YYYY-MM-DD format"}

        hash_input = f"{employee_name}{start_date}{end_date}{datetime.now().isoformat()}"
        report_id = f"EXP-{datetime.now().strftime('%Y%m%d')}-{hash(hash_input) % 10000:04d}"

        by_category = {}
        total_amount = 0.0

        for expense in expenses:
            category = expense["category"].lower()
            amount = float(expense["amount"])

            if category not in by_category:
                by_category[category] = {"total": 0.0, "count": 0, "items": []}
            by_category[category]["total"] += amount
            by_category[category]["count"] += 1
            by_category[category]["items"].append(expense)
            total_amount += amount

        for category in by_category:
            by_category[category]["total"] = round(by_category[category]["total"], 2)

        expenses_list = []
        for expense in expenses:
            expenses_list.append({
                "date": expense["date"],
                "category": expense["category"].lower(),
                "amount": float(expense["amount"]),
                "description": expense.get("description", ""),
                "receipt_attached": expense.get("receipt_attached", False)
            })

        expenses_list.sort(key=lambda x: x["date"])

        return {
            "ok": True,
            "report_id": report_id,
            "employee_name": employee_name,
            "start_date": start_date,
            "end_date": end_date,
            "total_amount": round(total_amount, 2),
            "by_category": by_category,
            "expenses_list": expenses_list,
            "approval_status": "pending",
            "num_expenses": len(expenses),
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate expense report: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "employee_name": "John Smith",
        "start_date": "2026-06-01",
        "end_date": "2026-06-15",
        "expenses": [
            {"date": "2026-06-02", "category": "travel", "amount": 250.00, "description": "Flight to NYC"},
            {"date": "2026-06-03", "category": "accommodation", "amount": 180.00, "description": "Hotel NYC"},
            {"date": "2026-06-04", "category": "meals", "amount": 65.50, "description": "Client dinner"},
            {"date": "2026-06-05", "category": "transportation", "amount": 45.00, "description": "Uber"},
        ]
    })
    print(result)
