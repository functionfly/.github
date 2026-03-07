# Email Configuration Guide

This guide explains how to configure email functionality in FunctionFly using Resend for production/staging and Mailpit for local development.

## Overview

FunctionFly uses [Resend](https://resend.com) for sending transactional emails in production and staging environments. For local development, it automatically falls back to [Mailpit](https://github.com/axllent/mailpit) for zero-configuration email testing.

### Email Types

The platform sends the following types of emails:

- **Email Verification**: Sent during user registration with a 24-hour expiry token
- **Password Reset**: Sent when users request a password reset with a 1-hour expiry token  
- **GDPR Breach Notifications**: Sent to affected users for Article 33 compliance

### Features

- ✅ **Retry Logic**: Exponential backoff (1s, 2s, 4s) with up to 3 attempts
- ✅ **Smart Retry**: Only retries on rate limits (429) and server errors (5xx)
- ✅ **Batch Sending**: Optimized API calls for multiple recipients
- ✅ **Webhook Events**: Real-time tracking of delivered, bounced, opened, clicked, and complained events
- ✅ **Bounce Management**: Admin dashboard for reviewing and managing bounce events
- ✅ **Email Analytics**: Delivery rates, bounce rates, and event statistics
- ✅ **Custom Domain**: Support for branded email sending (e.g., noreply@functionfly.com)

## Local Development

Local development uses Mailpit, which provides:

- **SMTP Server**: localhost:1025
- **Web Interface**: http://localhost:8025
- **Zero Configuration**: No API keys or DNS setup required
- **Email Testing**: View sent emails in the browser

### Setup

1. Start the development environment:

```bash
docker-compose up -d
```

2. Mailpit will automatically start and be available at http://localhost:8025

3. The application will automatically use Mailpit when `RESEND_API_KEY` is not set

## Production/Staging Setup

### 1. Create Resend Account

1. Sign up at https://resend.com
2. Navigate to the [API Keys](https://resend.com/api-keys) page
3. Create a new API key with full access
4. Copy the API key (starts with `re_`)

### 2. Configure Environment Variables

Add the following to your `.env.production` or `.env.staging` file:

```bash
# Resend Configuration
RESEND_API_KEY=re_your_api_key_here
RESEND_WEBHOOK_SECRET=whsec_your_webhook_secret_here
RESEND_DOMAIN=functionfly.com

# Email Settings
FROM_EMAIL=noreply@functionfly.com
FROM_NAME=FunctionFly
BASE_URL=https://functionfly.com
```

**Important**: Keep your API key secure. Never commit it to version control.

### 3. Set Up Custom Domain

To send emails from your own domain (e.g., functionfly.com), you need to verify domain ownership with DNS records.

#### Step-by-Step DNS Configuration

1. **Add Domain in Resend Dashboard**
   - Go to https://resend.com/domains
   - Click "Add Domain"
   - Enter your domain (e.g., `functionfly.com`)

2. **Add DNS Records**

   Resend will provide you with DNS records to add. You need to configure three types of records:

   **DKIM Record** (Required for email authentication):
   ```
   Type: TXT
   Host: resend._domainkey
   Value: p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC... (provided by Resend)
   TTL: 3600
   ```

   **SPF Record** (Required for sender verification):
   ```
   Type: TXT
   Host: @
   Value: v=spf1 include:resend.io ~all
   TTL: 3600
   ```

   If you already have an SPF record, add `include:resend.io` to it:
   ```
   v=spf1 include:existing-provider.com include:resend.io ~all
   ```

   **DMARC Record** (Recommended for email security):
   ```
   Type: TXT
   Host: _dmarc
   Value: v=DMARC1; p=quarantine; pct=100; rua=mailto:dmarc@functionfly.com
   TTL: 3600
   ```

3. **Verify DNS Records**

   DNS propagation can take up to 48 hours, but typically completes within 1-2 hours.

   Check DNS propagation:
   ```bash
   # Check DKIM record
   dig TXT resend._domainkey.functionfly.com +short

   # Check SPF record
   dig TXT functionfly.com +short | grep spf

   # Check DMARC record
   dig TXT _dmarc.functionfly.com +short
   ```

4. **Verify Domain in Resend**
   - Return to https://resend.com/domains
   - Click "Verify" next to your domain
   - Resend will check the DNS records
   - Once verified, you'll see a green checkmark

### 4. Configure Webhooks

Webhooks allow FunctionFly to track email events (delivered, bounced, opened, clicked, complained).

#### Create Webhook in Resend

1. Go to https://resend.com/webhooks
2. Click "Create Webhook"
3. Configure:
   - **Endpoint URL**: `https://functionfly.com/webhooks/resend`
   - **Events**: Select all events:
     - `email.sent`
     - `email.delivered`
     - `email.bounced`
     - `email.complained`
     - `email.opened`
     - `email.clicked`
   - **Status**: Active

4. **Copy Webhook Secret**
   - After creating, copy the webhook signing secret (starts with `whsec_`)
   - Add it to your environment variables as `RESEND_WEBHOOK_SECRET`

#### Webhook Security

Webhooks are secured with HMAC-SHA256 signatures in Svix format. The webhook handler automatically verifies signatures to prevent unauthorized requests.

### 5. Run Database Migration

The email events tracking system requires a database migration:

```bash
# Production
make migrate-up

# Or manually
migrate -database "postgres://user:pass@host:5432/functionfly?sslmode=require" -path migrations up
```

This creates the `email_events` table with indexes for efficient querying.

## Email Event Tracking

All email events are stored in the database and accessible via the admin dashboard.

### Event Types

- **email.sent**: Email was sent to Resend API
- **email.delivered**: Email was successfully delivered to recipient's inbox
- **email.bounced**: Email bounced (permanent or temporary failure)
- **email.complained**: Recipient marked email as spam
- **email.opened**: Recipient opened the email (requires tracking pixel)
- **email.clicked**: Recipient clicked a link in the email

### Admin Dashboard

Access email analytics and bounce management at:

```
https://functionfly.com/admin/email-events
```

#### Available Endpoints

**List All Email Events** (with filters):
```
GET /admin/email-events?event_type=email.bounced&limit=50&offset=0
```

Query parameters:
- `event_type`: Filter by event type
- `user_email`: Filter by recipient email
- `user_id`: Filter by user UUID
- `reviewed`: Filter by review status (true/false)
- `limit`: Number of results (default: 50, max: 100)
- `offset`: Pagination offset

**List Pending Bounce Reviews**:
```
GET /admin/email-events/bounces
```

Returns all bounced and complained emails that haven't been reviewed yet.

**Mark Event as Reviewed**:
```
POST /admin/email-events/{id}/review
Content-Type: application/json

{
  "admin_id": "uuid-of-admin-user"
}
```

**Get Email Statistics**:
```
GET /admin/email-events/stats?user_email=user@example.com
```

Returns:
- Total events count
- Events breakdown by type
- Total bounces
- Pending reviews count
- Delivery rate percentage

## Resend Free Tier Limits

The free tier includes:

- **3,000 emails per month**
- **100 emails per day**
- **1 custom domain**
- **Webhook support**
- **Email analytics**

For higher volumes, upgrade to a paid plan at https://resend.com/pricing

## Testing Email Configuration

### Test Email Sending

```bash
# Test verification email
curl -X POST https://functionfly.com/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "name": "Test User"
  }'

# Check email events
curl https://functionfly.com/admin/email-events \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Test Webhook Delivery

1. Send a test email through the platform
2. Check Resend dashboard: https://resend.com/emails
3. View webhook events: https://resend.com/webhooks
4. Verify event was stored in database:

```bash
curl https://functionfly.com/admin/email-events/stats \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## Troubleshooting

### Emails Not Sending

**Check API Key**:
```bash
# Verify API key is set
echo $RESEND_API_KEY

# Should start with 're_'
```

**Check Logs**:
```bash
# Docker logs
docker-compose logs -f functionfly-server

# Look for errors like:
# - "Failed to send email"
# - "Resend API error"
# - "Invalid API key"
```

**Verify Service Configuration**:
```bash
# Check which email service is initialized
# Look for log entry:
# "Email service initialized: Resend (domain: functionfly.com)"
# or
# "Email service initialized: Mailpit (UI: http://localhost:8025)"
```

### DNS Verification Failing

**Check DNS Propagation**:
```bash
# DKIM
dig TXT resend._domainkey.functionfly.com +short

# SPF  
dig TXT functionfly.com +short | grep spf

# DMARC
dig TXT _dmarc.functionfly.com +short
```

**Common Issues**:
- DNS records not yet propagated (wait 1-2 hours)
- Incorrect host/subdomain format (check Resend dashboard for exact values)
- TTL too high (set to 3600 or lower)
- Cloudflare proxy enabled (should be DNS only for TXT records)

### Webhooks Not Receiving Events

**Verify Webhook URL**:
- Must be publicly accessible (not localhost)
- Must use HTTPS in production
- Path should be `/webhooks/resend`

**Check Webhook Secret**:
```bash
# Verify secret is set
echo $RESEND_WEBHOOK_SECRET

# Should start with 'whsec_'
```

**Test Webhook Endpoint**:
```bash
# Should return 200 OK
curl -X POST https://functionfly.com/webhooks/resend \
  -H "Content-Type: application/json" \
  -H "svix-id: test" \
  -H "svix-timestamp: $(date +%s)" \
  -H "svix-signature: test" \
  -d '{"type":"email.test"}'
```

**View Webhook Logs in Resend**:
- Go to https://resend.com/webhooks
- Click on your webhook
- View recent webhook deliveries and response codes

### High Bounce Rate

**Causes**:
- Invalid email addresses
- Inbox full
- Email marked as spam
- Domain reputation issues

**Solutions**:
1. Validate email addresses before sending
2. Review bounce reasons in admin dashboard
3. Remove chronically bouncing addresses
4. Ensure compliance with anti-spam laws (CAN-SPAM, GDPR)
5. Set up DMARC policy
6. Monitor sender reputation

### Rate Limiting

Resend enforces rate limits:
- Free tier: 100 emails/day, 10 emails/second
- Paid tiers: Higher limits based on plan

The retry logic automatically handles 429 rate limit errors with exponential backoff.

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RESEND_API_KEY` | Production | - | Resend API key (starts with `re_`) |
| `RESEND_WEBHOOK_SECRET` | Production | - | Webhook signing secret (starts with `whsec_`) |
| `RESEND_DOMAIN` | No | - | Custom domain for sending (e.g., `functionfly.com`) |
| `FROM_EMAIL` | No | `noreply@functionfly.com` | Default sender email address |
| `FROM_NAME` | No | `FunctionFly` | Default sender name |
| `BASE_URL` | Yes | `http://localhost:8080` | Application base URL for email links |

## Migration from SMTP

If migrating from the old SMTP system:

1. **Update environment variables** - Remove SMTP_* vars, add RESEND_* vars
2. **Run database migration** - Creates email_events table
3. **Configure DNS records** - Set up DKIM, SPF, DMARC
4. **Set up webhooks** - Configure webhook in Resend dashboard  
5. **Test email sending** - Verify emails are delivered successfully
6. **Monitor bounce rates** - Check admin dashboard for any issues

The old SMTP code has been removed. Mailpit remains available for local development.

## Additional Resources

- [Resend Documentation](https://resend.com/docs)
- [Resend API Reference](https://resend.com/docs/api-reference/introduction)
- [Resend Go SDK](https://github.com/resend/resend-go)
- [Mailpit Documentation](https://github.com/axllent/mailpit)
- [SPF Record Checker](https://mxtoolbox.com/spf.aspx)
- [DKIM Record Checker](https://mxtoolbox.com/dkim.aspx)
- [DMARC Record Checker](https://mxtoolbox.com/dmarc.aspx)

## Support

For issues or questions:

- **Resend Support**: https://resend.com/support
- **FunctionFly Issues**: https://github.com/functionfly/functionfly/issues
- **Email Configuration**: Check logs with `docker-compose logs -f functionfly-server`
