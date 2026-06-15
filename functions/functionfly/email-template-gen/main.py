"""Email Template Generator - Generate professional email templates."""
import re
from datetime import datetime
from typing import Any
from html import escape


def generate_welcome_email(company_name: str, recipient_name: str, custom_fields: dict = None) -> dict:
    """Generate welcome email template."""
    first_name = recipient_name.split()[0] if recipient_name else "Valued Customer"

    subject = f"Welcome to {company_name}!"

    cta_button = custom_fields.get("cta_button", "Get Started") if custom_fields else "Get Started"
    cta_url = custom_fields.get("cta_url", "#") if custom_fields else "#"

    body_html = f"""
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body {{ font-family: Arial, sans-serif; line-height: 1.6; color: #333; }}
        .container {{ max-width: 600px; margin: 0 auto; padding: 20px; }}
        .header {{ background: #4A90D9; color: white; padding: 20px; text-align: center; }}
        .content {{ padding: 20px; background: #f9f9f9; }}
        .cta-button {{ display: inline-block; background: #4A90D9; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }}
        .footer {{ padding: 20px; text-align: center; font-size: 12px; color: #666; }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {escape(company_name)}!</h1>
        </div>
        <div class="content">
            <p>Hi {escape(first_name)},</p>
            <p>Thank you for joining us! We're thrilled to have you on board.</p>
            <p>Here's what you can expect:</p>
            <ul>
                <li>Access to exclusive features and content</li>
                <li>Regular updates and newsletters</li>
                <li>Special offers for members</li>
            </ul>
            <p><a href="{escape(cta_url)}" class="cta-button">{escape(cta_button)}</a></p>
            <p>If you have any questions, feel free to reply to this email.</p>
            <p>Best regards,<br>The {escape(company_name)} Team</p>
        </div>
        <div class="footer">
            <p>{escape(company_name)} | © {datetime.now().year}</p>
        </div>
    </div>
</body>
</html>"""

    plain_text = f"""Welcome to {company_name}!

Hi {first_name},

Thank you for joining us! We're thrilled to have you on board.

Here's what you can expect:
- Access to exclusive features and content
- Regular updates and newsletters
- Special offers for members

Get started by visiting our website.

If you have any questions, feel free to reply to this email.

Best regards,
The {company_name} Team

---
{company_name} | © {datetime.now().year}"""

    return {
        "subject": subject,
        "body": body_html.strip(),
        "plain_text": plain_text.strip(),
        "cta_button_text": cta_button
    }


def generate_reset_password_email(company_name: str, recipient_name: str, custom_fields: dict = None) -> dict:
    """Generate password reset email template."""
    first_name = recipient_name.split()[0] if recipient_name else "Valued Customer"
    reset_link = custom_fields.get("reset_link", "#") if custom_fields else "#"
    expiry_hours = custom_fields.get("expiry_hours", 24) if custom_fields else 24

    subject = f"Reset Your {company_name} Password"

    body_html = f"""
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body {{ font-family: Arial, sans-serif; line-height: 1.6; color: #333; }}
        .container {{ max-width: 600px; margin: 0 auto; padding: 20px; }}
        .header {{ background: #E74C3C; color: white; padding: 20px; text-align: center; }}
        .content {{ padding: 20px; background: #f9f9f9; }}
        .cta-button {{ display: inline-block; background: #E74C3C; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }}
        .warning {{ background: #FFF3CD; padding: 10px; border-radius: 4px; margin: 10px 0; }}
        .footer {{ padding: 20px; text-align: center; font-size: 12px; color: #666; }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hi {escape(first_name)},</p>
            <p>We received a request to reset your password for your {escape(company_name)} account.</p>
            <p><a href="{escape(reset_link)}" class="cta-button">Reset Password</a></p>
            <div class="warning">
                <p><strong>Important:</strong> This link will expire in {expiry_hours} hours.</p>
            </div>
            <p>If you didn't request this password reset, please ignore this email. Your password will remain unchanged.</p>
            <p>Best regards,<br>The {escape(company_name)} Team</p>
        </div>
        <div class="footer">
            <p>{escape(company_name)} | © {datetime.now().year}</p>
        </div>
    </div>
</body>
</html>"""

    plain_text = f"""Password Reset Request - {company_name}

Hi {first_name},

We received a request to reset your password for your {company_name} account.

To reset your password, visit: {reset_link}

Important: This link will expire in {expiry_hours} hours.

If you didn't request this password reset, please ignore this email. Your password will remain unchanged.

Best regards,
The {company_name} Team

---
{company_name} | © {datetime.now().year}"""

    return {
        "subject": subject,
        "body": body_html.strip(),
        "plain_text": plain_text.strip(),
        "cta_button_text": "Reset Password"
    }


def generate_invoice_email(company_name: str, recipient_name: str, custom_fields: dict = None) -> dict:
    """Generate invoice email template."""
    first_name = recipient_name.split()[0] if recipient_name else "Valued Customer"
    invoice_number = custom_fields.get("invoice_number", "INV-001") if custom_fields else "INV-001"
    amount_due = custom_fields.get("amount_due", "$0.00") if custom_fields else "$0.00"
    due_date = custom_fields.get("due_date", "N/A") if custom_fields else "N/A"
    pay_url = custom_fields.get("pay_url", "#") if custom_fields else "#"

    subject = f"Invoice {invoice_number} from {company_name}"

    body_html = f"""
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body {{ font-family: Arial, sans-serif; line-height: 1.6; color: #333; }}
        .container {{ max-width: 600px; margin: 0 auto; padding: 20px; }}
        .header {{ background: #27AE60; color: white; padding: 20px; text-align: center; }}
        .content {{ padding: 20px; background: #f9f9f9; }}
        .invoice-details {{ background: white; padding: 15px; border-radius: 4px; margin: 15px 0; }}
        .cta-button {{ display: inline-block; background: #27AE60; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }}
        .footer {{ padding: 20px; text-align: center; font-size: 12px; color: #666; }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Invoice</h1>
        </div>
        <div class="content">
            <p>Hi {escape(first_name)},</p>
            <p>Thank you for your business! Please find your invoice details below.</p>
            <div class="invoice-details">
                <p><strong>Invoice Number:</strong> {escape(invoice_number)}</p>
                <p><strong>Amount Due:</strong> {escape(amount_due)}</p>
                <p><strong>Due Date:</strong> {escape(due_date)}</p>
            </div>
            <p><a href="{escape(pay_url)}" class="cta-button">Pay Now</a></p>
            <p>If you have any questions about this invoice, please don't hesitate to contact us.</p>
            <p>Best regards,<br>The {escape(company_name)} Team</p>
        </div>
        <div class="footer">
            <p>{escape(company_name)} | © {datetime.now().year}</p>
        </div>
    </div>
</body>
</html>"""

    plain_text = f"""Invoice {invoice_number} from {company_name}

Hi {first_name},

Thank you for your business! Please find your invoice details below.

Invoice Number: {invoice_number}
Amount Due: {amount_due}
Due Date: {due_date}

Pay now: {pay_url}

If you have any questions about this invoice, please don't hesitate to contact us.

Best regards,
The {company_name} Team

---
{company_name} | © {datetime.now().year}"""

    return {
        "subject": subject,
        "body": body_html.strip(),
        "plain_text": plain_text.strip(),
        "cta_button_text": "Pay Now"
    }


def generate_notification_email(company_name: str, recipient_name: str, custom_fields: dict = None) -> dict:
    """Generate notification email template."""
    first_name = recipient_name.split()[0] if recipient_name else "Valued Customer"
    notification_title = custom_fields.get("notification_title", "Important Update") if custom_fields else "Important Update"
    notification_message = custom_fields.get("notification_message", "Please review the latest updates.") if custom_fields else "Please review the latest updates."
    action_url = custom_fields.get("action_url", "#") if custom_fields else "#"
    action_text = custom_fields.get("action_text", "View Details") if custom_fields else "View Details"

    subject = f"{company_name}: {notification_title}"

    body_html = f"""
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body {{ font-family: Arial, sans-serif; line-height: 1.6; color: #333; }}
        .container {{ max-width: 600px; margin: 0 auto; padding: 20px; }}
        .header {{ background: #9B59B6; color: white; padding: 20px; text-align: center; }}
        .content {{ padding: 20px; background: #f9f9f9; }}
        .notification-box {{ background: white; padding: 15px; border-radius: 4px; margin: 15px 0; border-left: 4px solid #9B59B6; }}
        .cta-button {{ display: inline-block; background: #9B59B6; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }}
        .footer {{ padding: 20px; text-align: center; font-size: 12px; color: #666; }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{escape(notification_title)}</h1>
        </div>
        <div class="content">
            <p>Hi {escape(first_name)},</p>
            <div class="notification-box">
                <p>{escape(notification_message)}</p>
            </div>
            <p><a href="{escape(action_url)}" class="cta-button">{escape(action_text)}</a></p>
            <p>Best regards,<br>The {escape(company_name)} Team</p>
        </div>
        <div class="footer">
            <p>{escape(company_name)} | © {datetime.now().year}</p>
        </div>
    </div>
</body>
</html>"""

    plain_text = f"""{notification_title} - {company_name}

Hi {first_name},

{notification_message}

View details: {action_url}

Best regards,
The {company_name} Team

---
{company_name} | © {datetime.now().year}"""

    return {
        "subject": subject,
        "body": body_html.strip(),
        "plain_text": plain_text.strip(),
        "cta_button_text": action_text
    }


def handler(event: dict) -> dict:
    """Generate an email template."""
    try:
        email_type = event.get("email_type")
        company_name = event.get("company_name")
        recipient_name = event.get("recipient_name")
        custom_fields = event.get("custom_fields", {})

        if not email_type:
            return {"ok": False, "error": "email_type is required (welcome/reset-password/invoice/notification)"}
        if not company_name:
            return {"ok": False, "error": "company_name is required"}
        if not recipient_name:
            return {"ok": False, "error": "recipient_name is required"}

        valid_types = ["welcome", "reset-password", "invoice", "notification"]
        if email_type not in valid_types:
            return {"ok": False, "error": f"email_type must be one of: {', '.join(valid_types)}"}

        generators = {
            "welcome": generate_welcome_email,
            "reset-password": generate_reset_password_email,
            "invoice": generate_invoice_email,
            "notification": generate_notification_email
        }

        result = generators[email_type](company_name, recipient_name, custom_fields)

        return {
            "ok": True,
            "email_type": email_type,
            "company_name": company_name,
            "recipient_name": recipient_name,
            "subject": result["subject"],
            "body": result["body"],
            "plain_text": result["plain_text"],
            "cta_button_text": result.get("cta_button_text"),
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate email template: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "email_type": "welcome",
        "company_name": "Acme Corp",
        "recipient_name": "John Smith",
        "custom_fields": {"cta_button": "Complete Profile", "cta_url": "https://example.com/setup"}
    })
    print(result)
