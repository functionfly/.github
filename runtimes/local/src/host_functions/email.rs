//! Email host function implementation

use lettre::message::header::ContentType;
use lettre::transport::smtp::authentication::Credentials;
use lettre::{Message, SmtpTransport, Transport};
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Add the functionfly.email function for sending emails
pub fn add_email_function(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // functionfly.email(to_ptr: i32, to_len: i32, subject_ptr: i32, subject_len: i32,
    //                    body_ptr: i32, body_len: i32, from_ptr: i32, from_len: i32) -> i32
    // Returns 0 on success, negative values on error
    linker.func_wrap(
        "functionfly",
        "email",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              to_ptr: i32,
              to_len: i32,
              subject_ptr: i32,
              subject_len: i32,
              body_ptr: i32,
              body_len: i32,
              from_ptr: i32,
              from_len: i32|
              -> i32 {
            // Get email parameters from WASM memory
            let to = match memory_utils::read_string_from_memory(&mut caller, to_ptr, to_len) {
                Ok(t) => t,
                Err(_) => return -1, // Invalid recipient
            };

            let subject = match memory_utils::read_string_from_memory(
                &mut caller,
                subject_ptr,
                subject_len,
            ) {
                Ok(s) => s,
                Err(_) => return -2, // Invalid subject
            };

            let body = match memory_utils::read_string_from_memory(&mut caller, body_ptr, body_len)
            {
                Ok(b) => b,
                Err(_) => return -3, // Invalid body
            };

            let from = if from_len > 0 {
                match memory_utils::read_string_from_memory(&mut caller, from_ptr, from_len) {
                    Ok(f) => Some(f),
                    Err(_) => return -4, // Invalid sender
                }
            } else {
                None
            };

            // Send email
            let result = send_email(&to, &subject, &body, from.as_deref(), &config);

            match result {
                Ok(_) => 0,   // Success
                Err(_) => -5, // Email send failed
            }
        },
    )?;

    tracing::debug!("Added functionfly.email host function");
    Ok(())
}

/// Send an email
pub fn send_email(
    to: &str,
    subject: &str,
    body: &str,
    from: Option<&str>,
    config: &Config,
) -> anyhow::Result<()> {
    // Use provided from address or default to noreply@functionfly.local
    let from_addr = from.unwrap_or("noreply@functionfly.local");

    // Build the email message
    let email = Message::builder()
        .from(from_addr.parse()?)
        .to(to.parse()?)
        .subject(subject)
        .header(ContentType::TEXT_PLAIN)
        .body(body.to_string())?;

    // Create SMTP transport with proper TLS and timeout
    // Use SmtpTransport::relay for implicit TLS (port 465) or
    // SmtpTransport::starttls_relay for explicit TLS (port 587)
    let relay_builder = if config.smtp_port == 465 || config.smtp_use_tls {
        // Implicit TLS (SMTPS) - use relay() which creates TLS connection directly
        SmtpTransport::relay(&config.smtp_host)?
    } else {
        // Explicit TLS (STARTTLS) - use starttls_relay()
        SmtpTransport::starttls_relay(&config.smtp_host)?
    };

    // Configure the builder: port, timeout, and authentication
    let relay_builder = relay_builder
        .port(config.smtp_port)
        .timeout(Some(std::time::Duration::from_millis(1)));

    // Add authentication if credentials are provided
    let relay_builder =
        if let (Some(username), Some(password)) = (&config.smtp_username, &config.smtp_password) {
            let creds = Credentials::new(username.clone(), password.clone());
            relay_builder.credentials(creds)
        } else {
            relay_builder
        };

    // Build the transport
    let mailer = relay_builder.build();

    // Send the email
    mailer.send(&email)?;

    tracing::info!(
        "Email sent successfully: to={}, subject={}, from={}",
        to,
        subject,
        from_addr
    );

    Ok(())
}
