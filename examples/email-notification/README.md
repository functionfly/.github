# Email Notification Example

This example demonstrates how to use the `email` capability in FunctionFly functions.

## Overview

The email capability allows functions to send emails through configured SMTP servers. This is useful for notifications, alerts, and user communications.

## Configuration

Set the following environment variables:

- `FROM_EMAIL`: The sender email address (default: noreply@functionfly.dev)
- `NOTIFICATION_EMAIL`: The recipient email address

## Usage

```json
{
  "message": "Function executed successfully"
}
```

## Response

```json
{
  "status": "success",
  "message": "Email notification sent",
  "from": "noreply@functionfly.dev",
  "to": "admin@example.com",
  "subject": "FunctionFly Function Executed",
  "body_length": 150
}
```

## Security Notes

- Email sending is restricted to declared capabilities only
- SMTP configuration is managed by the platform
- Functions cannot send emails to arbitrary addresses without proper validation