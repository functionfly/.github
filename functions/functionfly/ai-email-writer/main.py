"""AI Email Writer - Write professional emails for various purposes."""
import random


GREETINGS = {
    "sales": "Dear {recipient_name},",
    "follow_up": "Hi {recipient_name},",
    "welcome": "Welcome, {recipient_name}!",
    "notification": "Hello {recipient_name},"
}

SIGNOFFS = {
    "sales": ["Looking forward to hearing from you.", "Best regards", "Warm regards"],
    "follow_up": ["Just wanted to follow up.", "Hope to connect soon.", "Looking forward to your response."],
    "welcome": ["We're excited to have you!", "Welcome aboard!", "Here's to great things ahead!"],
    "notification": ["Please let us know if you have questions.", "Thank you for your attention.", "We appreciate your understanding."]
}

EMAIL_BODIES = {
    "sales": """We believe our {product_service} could be exactly what you're looking for.

At {sender_name}, we've helped countless {recipient_name} achieve their goals with our proven solutions. Our approach combines expertise with dedication to customer success.

What sets us apart:
- Tailored solutions designed for your specific needs
- Proven track record of delivering results
- Dedicated support every step of the way

We'd love to show you what we can do. Would you be open to a brief conversation this week?

{cta_text}""",

    "follow_up": """I wanted to follow up on our recent conversation about {product_service}.

Have you had a chance to consider what we discussed? I'm happy to answer any questions or provide additional information.

If you're still interested, I'd love to schedule a quick call to explore how we can help you achieve your goals.

{cta_text}

Looking forward to hearing from you!""",

    "welcome": """Welcome to {sender_name}!

We're thrilled to have you on board. Here's what you need to know to get started:

1. Check your inbox for login credentials
2. Explore our getting started guide
3. Reach out to our support team anytime

{cta_text}

Again, welcome aboard! We're excited to have you with us.""",

    "notification": """We wanted to let you know about an important update regarding {product_service}.

{notification_details}

If you have any questions or concerns, please don't hesitate to reach out. We're here to help.

{cta_text}

Thank you for your continued trust in {sender_name}."""
}


def handler(event):
    try:
        purpose = event.get("purpose", "sales")
        recipient_name = event.get("recipient_name", "")
        sender_name = event.get("sender_name", "The Team")
        product_service = event.get("product_service", "our services")
        cta_text = event.get("cta_text", "Feel free to reach out")

        if not purpose:
            return {"ok": False, "error": "purpose is required"}
        if purpose not in ["sales", "follow_up", "welcome", "notification"]:
            return {"ok": False, "error": "purpose must be sales, follow_up, welcome, or notification"}
        if not recipient_name:
            return {"ok": False, "error": "recipient_name is required"}

        greeting = GREETINGS.get(purpose, GREETINGS["sales"]).format(recipient_name=recipient_name)
        signoff = random.choice(SIGNOFFS.get(purpose, SIGNOFFS["sales"]))

        notification_details = ""
        if purpose == "notification":
            notification_details = "Important changes have been made to your account settings. Please review the updates at your earliest convenience."

        body_template = EMAIL_BODIES.get(purpose, EMAIL_BODIES["sales"])
        body = body_template.format(
            recipient_name=recipient_name,
            sender_name=sender_name,
            product_service=product_service,
            cta_text=cta_text,
            notification_details=notification_details
        )

        subjects = {
            "sales": f"How {product_service} can help you achieve your goals",
            "follow_up": f"Following up on our conversation - {product_service}",
            "welcome": f"Welcome to {sender_name}!",
            "notification": f"Important update regarding your account"
        }
        subject = subjects.get(purpose, subjects["sales"])

        preview_texts = {
            "sales": f"Discover how {product_service} can transform your business...",
            "follow_up": f"Just checking in about {product_service}...",
            "welcome": f"Welcome! Here's everything you need to get started...",
            "notification": f"You have a new notification from {sender_name}..."
        }
        preview_text = preview_texts.get(purpose, preview_texts["notification"])

        return {
            "ok": True,
            "subject": subject,
            "body": f"{greeting}\n\n{body}\n\n{signoff},\n{sender_name}",
            "preview_text": preview_text,
            "purpose": purpose,
            "recipient_name": recipient_name,
            "sender_name": sender_name
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
