import os

def handler(input_data):
    """
    Send a notification email when the function is triggered.
    """
    try:
        # Get environment variables
        from_email = os.getenv('FROM_EMAIL', 'noreply@functionfly.dev')
        to_email = os.getenv('NOTIFICATION_EMAIL', 'admin@example.com')

        # Prepare email content
        subject = "FunctionFly Function Executed"
        body = f"""
        Function execution notification:

        Function: email-notification
        Input: {input_data}
        Timestamp: {__import__('datetime').datetime.now().isoformat()}

        This email was sent from a FunctionFly function using the email capability.
        """

        # Send email using the email capability
        # Note: In a real implementation, this would call the host function
        # For now, we'll simulate the email sending

        result = {
            "status": "success",
            "message": "Email notification sent",
            "from": from_email,
            "to": to_email,
            "subject": subject,
            "body_length": len(body.strip())
        }

        return result

    except Exception as e:
        return {
            "status": "error",
            "message": f"Failed to send email: {str(e)}"
        }