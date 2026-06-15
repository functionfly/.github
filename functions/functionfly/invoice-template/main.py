"""Invoice Template Generator - Generate reusable invoice templates."""
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate an invoice template."""
    try:
        business_name = event.get("business_name")
        business_address = event.get("business_address")
        logo_url = event.get("logo_url")
        payment_terms = event.get("payment_terms", "Net 30")

        if not business_name:
            return {"ok": False, "error": "business_name is required"}
        if not business_address:
            return {"ok": False, "error": "business_address is required"}

        template_html = f"""<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Invoice - {business_name}</title>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{ font-family: 'Helvetica Neue', Arial, sans-serif; font-size: 14px; line-height: 1.6; color: #333; padding: 40px; }}
        .invoice-container {{ max-width: 800px; margin: 0 auto; background: white; }}
        .header {{ display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 40px; padding-bottom: 20px; border-bottom: 2px solid #4A90D9; }}
        .logo {{ max-height: 80px; max-width: 200px; }}
        .business-info {{ text-align: right; }}
        .invoice-title {{ font-size: 36px; font-weight: 300; color: #4A90D9; margin-bottom: 10px; }}
        .invoice-meta {{ display: flex; justify-content: space-between; margin-bottom: 30px; }}
        .bill-to {{ background: #f9f9f9; padding: 20px; border-radius: 4px; }}
        .bill-to h3 {{ color: #666; font-size: 12px; text-transform: uppercase; margin-bottom: 10px; }}
        .invoice-details {{ text-align: right; }}
        .invoice-details p {{ margin: 5px 0; }}
        table {{ width: 100%; border-collapse: collapse; margin-bottom: 30px; }}
        th {{ background: #4A90D9; color: white; padding: 12px; text-align: left; font-weight: 500; }}
        th:last-child {{ text-align: right; }}
        td {{ padding: 12px; border-bottom: 1px solid #eee; }}
        td:last-child {{ text-align: right; }}
        .totals {{ margin-left: auto; width: 300px; }}
        .totals-row {{ display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }}
        .totals-row.total {{ border-bottom: none; border-top: 2px solid #4A90D9; font-size: 18px; font-weight: bold; color: #4A90D9; margin-top: 10px; padding-top: 15px; }}
        .notes {{ background: #f9f9f9; padding: 20px; border-radius: 4px; margin-bottom: 30px; }}
        .notes h4 {{ margin-bottom: 10px; color: #666; }}
        .footer {{ text-align: center; padding-top: 30px; border-top: 1px solid #eee; color: #666; font-size: 12px; }}
        .payment-info {{ margin-top: 20px; }}
        .payment-info h4 {{ color: #4A90D9; margin-bottom: 10px; }}
    </style>
</head>
<body>
    <div class="invoice-container">
        <div class="header">
            <div class="logo-section">
                {"<img src='" + logo_url + "' alt='Logo' class='logo' />" if logo_url else "<h1 style='font-size: 24px; color: #4A90D9;'>" + business_name + "</h1>"}
            </div>
            <div class="business-info">
                {business_address.replace(chr(10), '<br>')}
            </div>
        </div>

        <div class="invoice-meta">
            <div class="bill-to">
                <h3>Bill To</h3>
                <p><strong>[Client Name]</strong></p>
                <p>[Client Address Line 1]</p>
                <p>[Client Address Line 2]</p>
                <p>[City, State ZIP]</p>
            </div>
            <div class="invoice-details">
                <p><strong>Invoice Number:</strong> [INV-XXXX]</p>
                <p><strong>Invoice Date:</strong> [DATE]</p>
                <p><strong>Due Date:</strong> [DUE DATE]</p>
                <p><strong>Payment Terms:</strong> {payment_terms}</p>
            </div>
        </div>

        <table>
            <thead>
                <tr>
                    <th>Description</th>
                    <th>Quantity</th>
                    <th>Unit Price</th>
                    <th>Amount</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>[Service/Product Description]</td>
                    <td>[Qty]</td>
                    <td>$[Price]</td>
                    <td>$[Total]</td>
                </tr>
                <tr>
                    <td>[Service/Product Description]</td>
                    <td>[Qty]</td>
                    <td>$[Price]</td>
                    <td>$[Total]</td>
                </tr>
            </tbody>
        </table>

        <div class="totals">
            <div class="totals-row">
                <span>Subtotal:</span>
                <span>$[SUBTOTAL]</span>
            </div>
            <div class="totals-row">
                <span>Tax (%):</span>
                <span>$[TAX]</span>
            </div>
            <div class="totals-row total">
                <span>Total Due:</span>
                <span>$[TOTAL]</span>
            </div>
        </div>

        <div class="notes">
            <h4>Notes</h4>
            <p>[Add any additional notes or payment instructions here]</p>
        </div>

        <div class="payment-info">
            <h4>Payment Information</h4>
            <p>Bank: [Bank Name]</p>
            <p>Account Name: {business_name}</p>
            <p>Account Number: [XXXX XXXX XXXX]</p>
            <p>Routing Number: [XXXXXXXXX]</p>
        </div>

        <div class="footer">
            <p>Thank you for your business!</p>
            <p>{business_name} | {business_address.replace(chr(10), ' | ')}</p>
        </div>
    </div>
</body>
</html>"""

        template_text = f"""
{business_name}
{business_address}
================================================================================

INVOICE

Invoice Number: [INV-XXXX]
Invoice Date: [DATE]
Due Date: [DUE DATE]
Payment Terms: {payment_terms}

--------------------------------------------------------------------------------
BILL TO:
--------------------------------------------------------------------------------
[Client Name]
[Client Address Line 1]
[Client Address Line 2]
[City, State ZIP]

================================================================================
DESCRIPTION                          QTY    UNIT PRICE    AMOUNT
================================================================================
[Service/Product Description]         [Qty]  $[Price]      $[Total]
[Service/Product Description]         [Qty]  $[Price]      $[Total]

================================================================================
                                  Subtotal: $[SUBTOTAL]
                                      Tax (%): $[TAX]
                                    ─────────────────
                               TOTAL DUE: $[TOTAL]
================================================================================

NOTES:
[Add any additional notes or payment instructions here]

PAYMENT INFORMATION:
Bank: [Bank Name]
Account Name: {business_name}
Account Number: [XXXX XXXX XXXX]
Routing Number: [XXXXXXXXX]

--------------------------------------------------------------------------------
Thank you for your business!
{business_name} | {business_address.replace(chr(10), ' | ')}
"""

        return {
            "ok": True,
            "business_name": business_name,
            "template_html": template_html.strip(),
            "template_text": template_text.strip(),
            "placeholders": [
                "[Client Name]", "[Client Address]", "[INV-XXXX]", "[DATE]",
                "[DUE DATE]", "[Service/Product Description]", "[Qty]", "[Price]",
                "[Total]", "[SUBTOTAL]", "[TAX]", "[TOTAL]"
            ],
            "payment_terms": payment_terms,
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate invoice template: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "business_name": "Acme Services LLC",
        "business_address": "456 Commerce St\nNew York, NY 10001",
        "logo_url": "https://example.com/logo.png",
        "payment_terms": "Net 30"
    })
    print(result)
