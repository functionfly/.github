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
              from_len: i32| -> i32 {
            // Get email parameters from WASM memory
            let to = match memory_utils::read_string_from_memory(&mut caller, to_ptr, to_len) {
                Ok(t) => t,
                Err(_) => return -1, // Invalid recipient
            };

            let subject = match memory_utils::read_string_from_memory(&mut caller, subject_ptr, subject_len) {
                Ok(s) => s,
                Err(_) => return -2, // Invalid subject
            };

            let body = match memory_utils::read_string_from_memory(&mut caller, body_ptr, body_len) {
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
                Ok(_) => 0, // Success
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

    // Create SMTP transport
    let mut mailer = SmtpTransport::builder_dangerous(&config.smtp_host)
        .port(config.smtp_port);

    // Add authentication if credentials are provided
    if let (Some(username), Some(password)) = (&config.smtp_username, &config.smtp_password) {
        let creds = Credentials::new(username.clone(), password.clone());
        mailer = mailer.credentials(creds);
    }

    let mailer = mailer.build();

    // Send the email
    mailer.send(&email)?;

    tracing::info!(
        "Email sent successfully: to={}, subject={}, from={}",
        to, subject, from_addr
    );

    Ok(())
}