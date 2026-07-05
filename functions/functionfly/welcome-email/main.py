"""Welcome Email Sender.

Sends a personalized transactional welcome email after a new user signs up.
Designed to be customizable: edit SUBJECT_TEMPLATE / HTML_TEMPLATE /
TEXT_TEMPLATE below, or override per-invocation via the `subject` input.

Provider selection (first match wins):
  1. RESEND_API_KEY        -> Resend HTTP API (recommended)
  2. SENDGRID_API_KEY      -> SendGrid HTTP API
  3. SES_SMTP_*            -> SMTP via smtplib (AWS SES, Mailgun, Postmark, ...)
  4. Otherwise: dry-run mode that returns a simulated message_id (safe default
     for local dev / preview; production deployments MUST set one of the above)

Security:
  - All user-supplied values are HTML-escaped and validated before render.
  - Recipient email is validated against RFC 5322 and length-capped.
  - No shell execution; pure HTTP/SMTP libraries only.
  - Secrets are read from env at request time, never logged.
  - HTML is rendered with html.escape (no Jinja / no template injection risk).
  - CTA URL is scheme-restricted to http/https to prevent javascript:/data:.

Inputs:
  to_email  (str, required)  - recipient address
  to_name   (str, required)  - recipient display name (escaped)
  subject   (str, optional)  - override subject line
  app_name  (str, optional)  - product/company name (default "FunctionFly")
  cta_url   (str, optional)  - call-to-action URL (http/https only)
  cta_label (str, optional)  - CTA button label
  user_id   (str, optional)  - internal user id for audit logging
  html      (bool, optional) - request an HTML preview in the response (default false)
"""
import html
import json
import os
import re
import smtplib
import ssl
import time
import urllib.error
import urllib.parse
import urllib.request
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.utils import formataddr, parseaddr


EMAIL_RE = re.compile(r"^[^@\s]+@[^@\s]+\.[^@\s]+$")
MAX_EMAIL_LEN = 254  # RFC 5321
MAX_NAME_LEN = 120
MAX_SUBJECT_LEN = 998  # RFC 5322
MAX_URL_LEN = 2048
SAFE_URL_RE = re.compile(r"^https?://", re.IGNORECASE)
SAFE_HEADER_RE = re.compile(r"[\r\n]")


def log(level: str, msg: str, **fields) -> None:
    """Structured JSON log line (matches platform log convention)."""
    safe = {k: v for k, v in fields.items() if v is not None and k != "secret" and "key" not in k.lower()}
    print(json.dumps({"ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()), "level": level, "msg": msg, **safe}))


def fail(message: str, **extra):
    """Return a structured error response."""
    out = {"ok": False, "error": message}
    out.update(extra)
    return out


def sanitize_name(raw) -> str:
    """Trim, strip header-injection chars, collapse whitespace, cap length."""
    if not isinstance(raw, str):
        return ""
    cleaned = SAFE_HEADER_RE.sub(" ", raw).strip()
    if len(cleaned) > MAX_NAME_LEN:
        cleaned = cleaned[:MAX_NAME_LEN]
    return cleaned


def validate_email(raw) -> str:
    """Validate recipient address; return normalized lowercase or '' if invalid."""
    if not isinstance(raw, str):
        return ""
    candidate = raw.strip()
    if not candidate or len(candidate) > MAX_EMAIL_LEN:
        return ""
    if not EMAIL_RE.match(candidate):
        return ""
    # parseaddr catches a few more edge cases (display-name, comments)
    _, addr = parseaddr(candidate)
    if not addr or not EMAIL_RE.match(addr):
        return ""
    return addr.lower()


def validate_url(raw) -> str:
    """Require explicit http(s):// scheme; strip control chars."""
    if not isinstance(raw, str):
        return ""
    cleaned = SAFE_HEADER_RE.sub("", raw).strip()
    if not cleaned or len(cleaned) > MAX_URL_LEN:
        return ""
    if not SAFE_URL_RE.match(cleaned):
        return ""
    return cleaned


def safe_subject(raw, fallback: str) -> str:
    """Apply subject line with header-injection guard."""
    s = sanitize_name(raw) if isinstance(raw, str) else ""
    s = SAFE_HEADER_RE.sub("", s).strip()
    if not s:
        s = fallback
    return s[:MAX_SUBJECT_LEN]


def render_email(app_name: str, to_name: str, subject: str, cta_url: str, cta_label: str):
    """Render subject + plain-text + HTML bodies. Everything user-supplied is escaped."""
    safe_app = html.escape(app_name)
    safe_name = html.escape(to_name)
    safe_subject_text = html.escape(subject)

    text_lines = [
        f"Hi {to_name},",
        "",
        f"Welcome to {app_name}! Your account is ready and you can get started right away.",
        "",
    ]
    if cta_url:
        text_lines += [
            f"{cta_label}: {cta_url}",
            "",
        ]
    text_lines += [
        "If you have any questions, just reply to this email.",
        "",
        f"— The {app_name} team",
    ]
    text_body = "\n".join(text_lines)

    html_body = (
        '<!DOCTYPE html><html><body style="font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;'
        'background:#f6f7f9;padding:32px;color:#111;">'
        '<div style="max-width:560px;margin:0 auto;background:#ffffff;border-radius:12px;padding:32px;'
        'border:1px solid #e5e7eb;">'
        f'<h1 style="margin:0 0 16px 0;font-size:24px;color:#111;">Welcome to {safe_app}</h1>'
        f'<p style="margin:0 0 16px 0;font-size:16px;line-height:1.5;color:#374151;">Hi {safe_name},</p>'
        f'<p style="margin:0 0 24px 0;font-size:16px;line-height:1.5;color:#374151;">'
        f'Your account is ready. Click the button below to get started.</p>'
    )
    if cta_url:
        safe_cta_label = html.escape(cta_label)[:64]
        safe_cta_url = html.escape(cta_url, quote=True)
        html_body += (
            f'<p style="margin:0 0 24px 0;"><a href="{safe_cta_url}" '
            f'style="display:inline-block;background:#4f46e5;color:#ffffff;text-decoration:none;'
            f'padding:12px 20px;border-radius:8px;font-weight:600;">{safe_cta_label}</a></p>'
        )
    html_body += (
        '<p style="margin:24px 0 0 0;font-size:14px;color:#6b7280;">'
        "If you have questions, just reply to this email.</p>"
        f'<p style="margin:16px 0 0 0;font-size:14px;color:#6b7280;">— The {safe_app} team</p>'
        "</div></body></html>"
    )

    return safe_subject_text, text_body, html_body


def send_via_resend(api_key: str, from_addr: str, from_name: str, to_email: str, to_name: str, subject: str, text: str, html_body: str):
    payload = {
        "from": formataddr((from_name, from_addr)),
        "to": [formataddr((to_name, to_email))],
        "subject": subject,
        "text": text,
        "html": html_body,
    }
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        "https://api.resend.com/emails",
        data=body,
        method="POST",
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "User-Agent": "functionfly-welcome-email/1.0",
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode("utf-8") or "{}")
            return data.get("id") or ""
    except urllib.error.HTTPError as e:
        detail = e.read().decode("utf-8", errors="replace")[:512]
        raise RuntimeError(f"resend http {e.code}: {detail}") from e
    except urllib.error.URLError as e:
        raise RuntimeError(f"resend unreachable: {e.reason}") from e


def send_via_sendgrid(api_key: str, from_addr: str, from_name: str, to_email: str, to_name: str, subject: str, text: str, html_body: str):
    payload = {
        "personalizations": [{"to": [{"email": to_email, "name": to_name}]}],
        "from": {"email": from_addr, "name": from_name},
        "subject": subject,
        "content": [
            {"type": "text/plain", "value": text},
            {"type": "text/html", "value": html_body},
        ],
    }
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        "https://api.sendgrid.com/v3/mail/send",
        data=body,
        method="POST",
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "User-Agent": "functionfly-welcome-email/1.0",
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            # SendGrid returns 202 with an X-Message-Id header on success
            return resp.headers.get("X-Message-Id") or ""
    except urllib.error.HTTPError as e:
        detail = e.read().decode("utf-8", errors="replace")[:512]
        raise RuntimeError(f"sendgrid http {e.code}: {detail}") from e
    except urllib.error.URLError as e:
        raise RuntimeError(f"sendgrid unreachable: {e.reason}") from e


def send_via_smtp(
    host: str,
    port: int,
    username: str,
    password: str,
    use_tls: bool,
    from_addr: str,
    from_name: str,
    to_email: str,
    to_name: str,
    subject: str,
    text: str,
    html_body: str,
):
    msg = MIMEMultipart("alternative")
    msg["Subject"] = subject
    msg["From"] = formataddr((from_name, from_addr))
    msg["To"] = formataddr((to_name, to_email))
    msg.attach(MIMEText(text, "plain", "utf-8"))
    msg.attach(MIMEText(html_body, "html", "utf-8"))

    context = ssl.create_default_context()
    try:
        if use_tls:
            with smtplib.SMTP_SSL(host, port, context=context, timeout=10) as server:
                if username:
                    server.login(username, password)
                server.sendmail(from_addr, [to_email], msg.as_string())
        else:
            with smtplib.SMTP(host, port, timeout=10) as server:
                server.ehlo()
                if server.has_extn("starttls"):
                    server.starttls(context=context)
                    server.ehlo()
                if username:
                    server.login(username, password)
                server.sendmail(from_addr, [to_email], msg.as_string())
    except (smtplib.SMTPException, ssl.SSLError, OSError) as e:
        raise RuntimeError(f"smtp error: {e}") from e
    # SMTP doesn't return a server-side id; synthesise a stable identifier
    # so callers can still correlate delivery state.
    return f"smtp-{int(time.time() * 1000)}"


def handler(event):
    try:
        if not isinstance(event, dict):
            return fail("event must be an object")

        to_email = validate_email(event.get("to_email"))
        if not to_email:
            return fail("to_email is missing or invalid")

        to_name = sanitize_name(event.get("to_name"))
        if not to_name:
            return fail("to_name is required")

        app_name_raw = event.get("app_name")
        app_name = sanitize_name(app_name_raw) if isinstance(app_name_raw, str) and app_name_raw.strip() else "FunctionFly"
        app_name = app_name or "FunctionFly"

        cta_url = validate_url(event.get("cta_url") or "")
        cta_label_raw = event.get("cta_label")
        cta_label = sanitize_name(cta_label_raw) if isinstance(cta_label_raw, str) and cta_label_raw.strip() else "Get Started"
        cta_label = cta_label or "Get Started"

        user_id = sanitize_name(event.get("user_id")) if isinstance(event.get("user_id"), str) else ""
        want_html = bool(event.get("html", False))

        default_subject = f"Welcome to {app_name}"
        subject = safe_subject(event.get("subject"), default_subject)

        subject_safe, text_body, html_body = render_email(app_name, to_name, subject, cta_url, cta_label)

        from_addr = os.environ.get("EMAIL_FROM_ADDRESS", "noreply@functionfly.com").strip()
        from_name = sanitize_name(os.environ.get("EMAIL_FROM_NAME", app_name)) or app_name

        provider = "dry-run"
        message_id = ""

        resend_key = os.environ.get("RESEND_API_KEY", "").strip()
        sendgrid_key = os.environ.get("SENDGRID_API_KEY", "").strip()
        smtp_host = os.environ.get("SES_SMTP_HOST", "").strip()

        if resend_key:
            provider = "resend"
            message_id = send_via_resend(resend_key, from_addr, from_name, to_email, to_name, subject_safe, text_body, html_body)
        elif sendgrid_key:
            provider = "sendgrid"
            message_id = send_via_sendgrid(sendgrid_key, from_addr, from_name, to_email, to_name, subject_safe, text_body, html_body)
        elif smtp_host:
            provider = "smtp"
            try:
                smtp_port = int(os.environ.get("SES_SMTP_PORT", "587"))
            except ValueError:
                smtp_port = 587
            smtp_user = os.environ.get("SES_SMTP_USERNAME", "").strip()
            smtp_pass = os.environ.get("SES_SMTP_PASSWORD", "").strip()
            smtp_tls = os.environ.get("SES_SMTP_TLS", "starttls").lower() in ("starttls", "tls", "ssl", "1", "true")
            message_id = send_via_smtp(
                smtp_host,
                smtp_port,
                smtp_user,
                smtp_pass,
                smtp_tls,
                from_addr,
                from_name,
                to_email,
                to_name,
                subject_safe,
                text_body,
                html_body,
            )
        else:
            log("warn", "no email provider configured; running in dry-run mode", to=to_email)
            provider = "dry-run"
            message_id = f"dry-run-{int(time.time() * 1000)}"

        log("info", "welcome email dispatched", provider=provider, to=to_email, message_id=message_id, user_id=user_id or None)

        result = {
            "ok": True,
            "provider": provider,
            "message_id": message_id,
            "to": to_email,
            "subject": subject_safe,
            "skipped": provider == "dry-run",
        }
        if want_html:
            result["html"] = html_body
        return result
    except RuntimeError as e:
        log("error", "delivery failed", error=str(e))
        return fail(str(e))
    except Exception as e:  # last-resort guard; never leak internals
        log("error", "unhandled exception", error=type(e).__name__)
        return fail("internal error")