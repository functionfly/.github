"""Invoice Generator - Generate professional invoices."""
from datetime import datetime, timedelta
from typing import Any


def format_currency(amount: float) -> str:
    """Format amount as USD currency."""
    return f"${amount:,.2f}"


def calculate_line_items(line_items: list) -> dict:
    """Calculate subtotal, tax, and total."""
    subtotal = 0.0
    calculated_items = []

    for item in line_items:
        quantity = float(item.get("quantity", 1))
        unit_price = float(item.get("unit_price", 0))
        line_total = quantity * unit_price
        subtotal += line_total

        calculated_items.append({
            "description": item.get("description", "Service"),
            "quantity": quantity,
            "unit_price": unit_price,
            "line_total": round(line_total, 2)
        })

    subtotal = round(subtotal, 2)
    tax_rate = 0.0
    tax_amount = 0.0
    total = subtotal

    return {
        "items": calculated_items,
        "subtotal": subtotal,
        "tax_rate": tax_rate,
        "tax_amount": tax_amount,
        "total": total
    }


def handler(event: dict) -> dict:
    """Generate a formatted invoice."""
    try:
        invoice_number = event.get("invoice_number")
        client_name = event.get("client_name")
        client_address = event.get("client_address")
        line_items = event.get("line_items", [])
        due_date = event.get("due_date")
        notes = event.get("notes", "")

        if not invoice_number:
            return {"ok": False, "error": "invoice_number is required"}
        if not client_name:
            return {"ok": False, "error": "client_name is required"}
        if not client_address:
            return {"ok": False, "error": "client_address is required"}
        if not line_items or len(line_items) == 0:
            return {"ok": False, "error": "line_items is required and must not be empty"}
        if not due_date:
            return {"ok": False, "error": "due_date is required"}

        for item in line_items:
            if not isinstance(item, dict):
                return {"ok": False, "error": "each line_item must be an object"}
            if "description" not in item:
                return {"ok": False, "error": "each line_item must have a description"}

        try:
            parsed_due = datetime.strptime(due_date, "%Y-%m-%d")
        except ValueError:
            return {"ok": False, "error": "due_date must be in YYYY-MM-DD format"}

        today = datetime.now()
        issue_date = today.strftime("%Y-%m-%d")

        calc = calculate_line_items(line_items)

        invoice_lines = []
        invoice_lines.append("=" * 70)
        invoice_lines.append(f"{'INVOICE':^70}")
        invoice_lines.append("=" * 70)
        invoice_lines.append(f"Invoice Number: {invoice_number}")
        invoice_lines.append(f"Issue Date: {issue_date}")
        invoice_lines.append(f"Due Date: {due_date}")
        invoice_lines.append("")
        invoice_lines.append("-" * 70)
        invoice_lines.append("BILL TO:")
        invoice_lines.append(f"  {client_name}")
        for line in client_address.split('\n'):
            invoice_lines.append(f"  {line}")
        invoice_lines.append("-" * 70)
        invoice_lines.append("")
        invoice_lines.append(f"{'Description':<40} {'Qty':>6} {'Unit Price':>12} {'Amount':>12}")
        invoice_lines.append("-" * 70)

        for item in calc["items"]:
            desc = item["description"][:38] if len(item["description"]) > 38 else item["description"]
            invoice_lines.append(f"{desc:<40} {item['quantity']:>6.0f} {format_currency(item['unit_price']):>12} {format_currency(item['line_total']):>12}")

        invoice_lines.append("-" * 70)
        invoice_lines.append(f"{'Subtotal:':<60} {format_currency(calc['subtotal']):>10}")
        if calc['tax_amount'] > 0:
            invoice_lines.append(f"{'Tax (' + str(int(calc['tax_rate'] * 100)) + '%):':<60} {format_currency(calc['tax_amount']):>10}")
        invoice_lines.append("=" * 70)
        invoice_lines.append(f"{'TOTAL DUE:':<60} {format_currency(calc['total']):>10}")
        invoice_lines.append("=" * 70)

        if notes:
            invoice_lines.append("")
            invoice_lines.append("Notes:")
            invoice_lines.append(notes)

        invoice_lines.append("")
        invoice_lines.append("Payment Terms: Net 30")
        invoice_lines.append("Please make payment within 30 days of the invoice date.")
        invoice_lines.append("")

        payment_terms = "Payment is due within 30 days of invoice date. Please include the invoice number with your payment."

        return {
            "ok": True,
            "invoice_number": invoice_number,
            "invoice_text": "\n".join(invoice_lines),
            "client_name": client_name,
            "client_address": client_address,
            "line_items": calc["items"],
            "subtotal": calc["subtotal"],
            "tax": calc["tax_amount"] if calc["tax_amount"] > 0 else None,
            "total": calc["total"],
            "issue_date": issue_date,
            "due_date": due_date,
            "notes": notes if notes else None,
            "payment_terms": payment_terms,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate invoice: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "invoice_number": "INV-2026-001",
        "client_name": "Acme Corp",
        "client_address": "123 Business Ave\nSan Francisco, CA 94102",
        "line_items": [
            {"description": "Web Development Services", "quantity": 40, "unit_price": 150},
            {"description": "Design Consultation", "quantity": 10, "unit_price": 125},
        ],
        "due_date": "2026-07-15",
        "notes": "Thank you for your business!"
    })
    print(result)
