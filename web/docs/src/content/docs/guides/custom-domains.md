---
title: Custom Domains
description: Configure custom domains for your functions and applications on FunctionFly.
sidebar:
  order: 14
---



This guide shows how to configure custom domains for your functions, enabling branded URLs for your serverless endpoints.

## Overview

By default, your functions are accessible at:
```
https://functionfly.com/fx/{author}/{name}
```

With custom domains, you can use your own domain:
```
https://api.yourdomain.com/{function-path}
```

---

## Prerequisites

Before setting up a custom domain:

- [ ] Domain name you own (e.g., `yourdomain.com`)
- [ ] DNS access to manage domain records
- [ ] SSL certificate (we provide free auto-managed certificates)
- [ ] Professional plan or higher (required for custom domains)

---

## Step 1: Add Your Domain

### In the Dashboard

1. Go to **Settings → Domains**
2. Click **Add Custom Domain**
3. Enter your domain name (e.g., `api.yourdomain.com`)
4. Click **Add Domain**

### Domain Verification

After adding your domain, you'll need to verify ownership by adding a DNS record:

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| `TXT` | `yourdomain.com` | `ffly-verification=abc123xyz...` | Prove you own the domain |

**Wait for verification** — This can take anywhere from a few minutes to 48 hours.

### Verification Status

| Status | Meaning |
|--------|---------|
| **Pending** | DNS record added, waiting for propagation |
| **Verified** | Domain ownership confirmed |
| **Failed** | Verification failed — check DNS records |

---

## Step 2: Configure DNS

Once your domain is verified, add the DNS records to point to FunctionFly:

### For subdomain APIs (recommended)

To use `api.yourdomain.com` for your functions:

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| `CNAME` | `api` | `edge.functionfly.com` | Route API calls to FunctionFly |

### For wildcard subdomains

To use `*.yourdomain.com`:

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| `CNAME` | `*` | `edge.functionfly.com` | Route all subdomains |

### For apex/naked domains

Apex domains (e.g., `yourdomain.com`) cannot use CNAME records. Use A records:

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| `A` | `@` | `<FunctionFly IP>` | See dashboard for IP address |

**Note:** We recommend using a subdomain (e.g., `api.yourdomain.com`) instead of the apex domain for better reliability.

---

## Step 3: SSL Certificate

FunctionFly automatically provisions and manages SSL certificates via Let's Encrypt.

### Certificate Details

- **Auto-provisioned** after DNS verification
- **Auto-renewed** 30 days before expiration
- **Coverage** includes both `yourdomain.com` and `*.yourdomain.com`

### Viewing Certificate

1. Go to **Settings → Domains**
2. Click on your domain
3. View certificate details:
   - Issuer
   - Expiration date
   - Subject Alternative Names (SANs)

### Certificate Issues

If SSL fails after DNS propagation:

1. Check certificate status in dashboard
2. Verify DNS records are correct
3. Wait 5-10 minutes for auto-retry
4. If still failing, click **Recheck** to trigger reissuance

---

## Step 4: Route Traffic to Functions

After DNS and SSL are working, configure which functions use the domain.

### Route All Functions

Route all function calls through your custom domain:

```bash
# All functions at api.yourdomain.com/*
https://api.yourdomain.com/your-function

# Automatically routes to your-function
```

### Function Path Mapping

Map specific function paths:

1. Go to **Settings → Domains → yourdomain.com**
2. Click **Route Configuration**
3. Add path mappings:

| Path Pattern | Function | Version |
|-------------|----------|---------|
| `/users/*` | `user-service` | latest |
| `/payments/*` | `payment-service` | v2.1 |
| `/api/webhook` | `webhook-handler` | latest |

---

## Using Custom Domains in Functions

### Base URL Changes

When a request comes through your custom domain, your function sees:

| Header | Value |
|--------|-------|
| `Host` | `api.yourdomain.com` |
| `X-Forwarded-Host` | `api.yourdomain.com` |
| `X-Custom-Domain` | `api.yourdomain.com` |

### Detecting Custom Domain in Code

```python
import os

def handler(request):
    host = request.headers.get('Host', '')
    custom_domain = request.headers.get('X-Custom-Domain', '')
    
    if custom_domain:
        base_url = f"https://{custom_domain}"
    else:
        base_url = "https://functionfly.com"
    
    return {"base_url": base_url}
```

```javascript
export default async function handler(request) {
    const customDomain = request.headers['x-custom-domain'];
    const baseUrl = customDomain 
        ? `https://${customDomain}` 
        : 'https://functionfly.com';
    
    return { baseUrl };
}
```

---

## Multiple Custom Domains

You can configure multiple custom domains:

| Domain | Use Case |
|--------|----------|
| `api.yourdomain.com` | Production functions |
| `staging.yourdomain.com` | Staging/testing |
| `dev.yourdomain.com` | Development |

Each domain:
- Has its own SSL certificate
- Can have its own routing rules
- Can be assigned to specific functions

---

## Custom Domain Limits

| Plan | Custom Domains | Domains per Function |
|------|----------------|----------------------|
| Free | 0 | 0 |
| Starter | 1 | 2 |
| Professional | 5 | 10 |
| Enterprise | Unlimited | Unlimited |

---

## Troubleshooting

### Domain shows "Pending Verification"

**Cause:** DNS records haven't propagated yet.

**Solution:**
```bash
# Check DNS propagation
dig TXT yourdomain.com

# Or use an online DNS checker
# propagationchecker.com
```

Wait up to 48 hours. If still pending after 48 hours, verify the TXT record was entered correctly.

### SSL Certificate Not Issued

**Cause:** DNS not fully propagated, or CAA records blocking issuance.

**Solution:**
1. Check DNS A/CNAME records are correct
2. Verify no CAA records blocking Let's Encrypt
3. Click **Recheck** in dashboard

### Mixed Content Errors

**Cause:** Custom domain loads over HTTPS but resources reference HTTP URLs.

**Solution:** Ensure all asset URLs are protocol-relative or use HTTPS.

### Custom Domain Not Routing

**Cause:** Incorrect CNAME or routing rules not configured.

**Solution:**
1. Verify CNAME points to `edge.functionfly.com`
2. Go to **Settings → Domains → yourdomain.com**
3. Verify routing rules exist for your functions

### Apex Domain Not Working

**Cause:** Your DNS provider doesn't support ALIAS records.

**Solution:**
- Use a subdomain (recommended): `api.yourdomain.com`
- Or use your DNS provider's ALIAS/ANAME record type
- Or switch to a DNS provider that supports apex CNAME (Cloudflare)

---

## Advanced Configuration

### Redirect Root Domain to www

Configure redirects at the DNS level:

| Type | Name | Value |
|------|------|-------|
| `CNAME` | `www` | `yourdomain.com` |

Or redirect using Cloudflare Rules (or equivalent):

```
If: Host matches yourdomain.com
Then: Redirect to https://www.yourdomain.com
```

### Path-Based Routing

Route different paths to different functions:

```yaml
routes:
  - path: /users/*
    function: user-service
  - path: /products/*
    function: product-service
  - path: /*
    function: catch-all
```

### Wildcard Subdomains

Allow any subdomain to route dynamically:

| Pattern | Example Match |
|---------|--------------|
| `*.yourdomain.com` | `api.yourdomain.com`, `app.yourdomain.com` |

The function receives subdomain in headers:
```
X-Subdomain: api
```

---

## Removing a Custom Domain

To stop using a custom domain:

1. Go to **Settings → Domains**
2. Click on your domain
3. Click **Remove Domain**
4. Confirm removal

**Note:** DNS records should also be removed from your DNS provider to avoid confusion.

---

## Best Practices

1. **Use subdomains** — `api.yourdomain.com` instead of `yourdomain.com`
2. **Plan DNS propagation** — Add DNS records before starting the setup process
3. **Test locally first** — Use `ffly dev` to test before going live
4. **Monitor certificate expiration** — We auto-renew, but verify the process
5. **Set up monitoring** — Get alerts if custom domain becomes unavailable
6. **Keep records updated** — Remove domains you no longer use
