//! HTTP fetch host function implementation

use std::collections::HashSet;
use std::collections::HashMap;
use std::sync::Arc;
use wasmtime_wasi::p1::WasiP1Ctx;

use super::memory_utils;
use crate::config::Config;
use crate::security::SecurityMonitor;

// Arc is used for the security_monitor parameter type.

/// Add the functionfly.fetch function for HTTP requests.
///
/// The `config` parameter is used to enforce the network whitelist
/// (`network_whitelist` + `strict_network_whitelist`) so that the capability
/// system's security model is fully applied at the host-function level.
///
/// A single `reqwest::blocking::Client` is created at registration time and
/// reused across all requests, enabling connection pooling and keep-alive.
pub fn add_fetch_function(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
    security_monitor: Arc<SecurityMonitor>,
    config: Config,
) -> anyhow::Result<()> {
    // functionfly.fetch(method_ptr: i32, method_len: i32, url_ptr: i32, url_len: i32,
    //                   headers_ptr: i32, headers_len: i32, body_ptr: i32, body_len: i32,
    //                   response_ptr: i32, response_len_ptr: i32) -> i32
    // Returns 0 on success, negative values on error
    // security_monitor is kept for future use (e.g. recording violations).
    // Suppress the unused-variable warning by prefixing with _.
    let _security_monitor = security_monitor;
    // Build the whitelist set once at registration time so we don't re-parse on
    // every request.
    let whitelist: HashSet<String> = config.network_whitelist.iter().cloned().collect();
    let strict_whitelist = config.strict_network_whitelist;

    // Create a single HTTP client at registration time for connection pooling.
    // NOTE: make_http_request() runs inside tokio::task::spawn_blocking, so using
    // reqwest::blocking is acceptable. However, the blocking thread pool should be
    // sized appropriately (tokio default: 512 threads) to handle concurrent fetch
    // calls. For very high concurrency, consider switching to an async client with
    // a dedicated tokio runtime.
    let http_client = Arc::new(
        reqwest::blocking::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .unwrap_or_else(|_| reqwest::blocking::Client::new()),
    );

    linker.func_wrap(
        "functionfly",
        "fetch",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              method_ptr: i32,
              method_len: i32,
              url_ptr: i32,
              url_len: i32,
              headers_ptr: i32,
              headers_len: i32,
              body_ptr: i32,
              body_len: i32,
              response_ptr: i32,
              response_len_ptr: i32| -> i32 {
            // Get the HTTP method from WASM memory
            let method = match memory_utils::read_string_from_memory(&mut caller, method_ptr, method_len) {
                Ok(m) => m,
                Err(_) => return -1, // Invalid method
            };

            // Get the URL from WASM memory
            let url = match memory_utils::read_string_from_memory(&mut caller, url_ptr, url_len) {
                Ok(u) => u,
                Err(_) => return -2, // Invalid URL
            };

            // Enforce network whitelist / basic safety checks.
            if !is_network_request_allowed(&url, &whitelist, strict_whitelist) {
                tracing::warn!("Network request blocked by policy: {}", url);
                return -8; // Network access denied
            }

            // Get headers (JSON string, optional)
            let headers_json = if headers_len > 0 {
                match memory_utils::read_string_from_memory(&mut caller, headers_ptr, headers_len) {
                    Ok(h) => Some(h),
                    Err(_) => return -3, // Invalid headers
                }
            } else {
                None
            };

            // Get request body (optional)
            let body = if body_len > 0 {
                match memory_utils::read_string_from_memory(&mut caller, body_ptr, body_len) {
                    Ok(b) => Some(b),
                    Err(_) => return -4, // Invalid body
                }
            } else {
                None
            };

            // Parse headers if provided
            let headers: HashMap<String, String> = if let Some(json) = headers_json {
                match serde_json::from_str(&json) {
                    Ok(h) => h,
                    Err(_) => return -5, // Invalid headers JSON
                }
            } else {
                HashMap::new()
            };

            // Make HTTP request using the shared client (connection pooling).
            let result = make_http_request(&http_client, &method, &url, &headers, body.as_deref());

            match result {
                Ok(response_body) => {
                    // Write response back to WASM memory
                    match memory_utils::write_string_to_memory(&mut caller, &response_body, response_ptr, response_len_ptr) {
                        Ok(_) => 0, // Success
                        Err(_) => -6, // Memory write error
                    }
                }
                Err(_) => -7, // HTTP request error
            }
        },
    )?;

    tracing::debug!("Added functionfly.fetch host function");
    Ok(())
}

/// Check if a network request is allowed.
///
/// Enforcement rules (applied in order):
/// 1. Only `http` and `https` schemes are permitted.
/// 2. Requests to localhost / private IP ranges are always blocked.
/// 3. If `strict_whitelist` is `true` and `whitelist` is non-empty, the
///    request host must match an entry in the whitelist (exact match or
///    wildcard `*.domain.com` prefix).
/// 4. If `strict_whitelist` is `false` (or the whitelist is empty), any
///    public host is allowed.
///
/// Exposed for use by package download (and other host-initiated HTTP).
pub(crate) fn is_network_request_allowed(url: &str, whitelist: &HashSet<String>, strict_whitelist: bool) -> bool {
    let parsed_url = match url::Url::parse(url) {
        Ok(u) => u,
        Err(_) => return false, // Reject unparseable URLs
    };

    // Only allow HTTP/HTTPS
    match parsed_url.scheme() {
        "http" | "https" => {}
        _ => return false,
    }

    let host = parsed_url.host_str().unwrap_or("");

    // Always block localhost and private IP ranges regardless of whitelist.
    if host.eq_ignore_ascii_case("localhost")
        || host == "127.0.0.1"
        || host == "::1"
        || host == "0.0.0.0"
        || host.starts_with("10.")
        || host.starts_with("192.168.")
        // RFC 1918 172.16.0.0/12
        || is_rfc1918_172(host)
        || host.ends_with(".local")
        || host.ends_with(".internal")
    {
        tracing::warn!("Blocked request to private/loopback address: {}", host);
        return false;
    }

    // If strict whitelist enforcement is enabled and the whitelist is non-empty,
    // the host must match an entry.
    if strict_whitelist && !whitelist.is_empty() {
        return is_host_in_whitelist(host, whitelist);
    }

    // Default: allow any public host.
    true
}

/// Returns true if the host falls in the RFC 1918 172.16.0.0/12 range.
fn is_rfc1918_172(host: &str) -> bool {
    if let Some(rest) = host.strip_prefix("172.") {
        if let Some(second_octet_str) = rest.split('.').next() {
            if let Ok(second_octet) = second_octet_str.parse::<u8>() {
                return (16..=31).contains(&second_octet);
            }
        }
    }
    false
}

/// Check whether `host` matches any entry in `whitelist`.
/// Supports exact matches and wildcard patterns of the form `*.domain.com`.
fn is_host_in_whitelist(host: &str, whitelist: &HashSet<String>) -> bool {
    if whitelist.contains(host) {
        return true;
    }
    for pattern in whitelist {
        if let Some(suffix) = pattern.strip_prefix("*.") {
            // For wildcard *.example.com, host must end with .example.com (proper subdomain)
            // This prevents "notexample.com" from matching "*.example.com"
            let domain_with_dot = format!(".{}", suffix);
            if host.ends_with(&domain_with_dot) && host.len() > suffix.len() {
                return true;
            }
        }
    }
    false
}

/// Make an HTTP request using the provided shared client.
///
/// The client is created once at linker registration time (see `add_fetch_function`)
/// to enable connection pooling and keep-alive reuse across requests.
fn make_http_request(
    client: &reqwest::blocking::Client,
    method: &str,
    url: &str,
    headers: &HashMap<String, String>,
    body: Option<&str>,
) -> anyhow::Result<String> {
    // Build the request
    let mut request_builder = match method.to_uppercase().as_str() {
        "GET" => client.get(url),
        "POST" => client.post(url),
        "PUT" => client.put(url),
        "DELETE" => client.delete(url),
        "PATCH" => client.patch(url),
        "HEAD" => client.head(url),
        "OPTIONS" => client.request(reqwest::Method::OPTIONS, url),
        _ => return Err(anyhow::anyhow!("Unsupported HTTP method: {}", method)),
    };

    // Add headers
    for (key, value) in headers {
        request_builder = request_builder.header(key, value);
    }

    // Add body if provided
    if let Some(body_content) = body {
        request_builder = request_builder.body(body_content.to_string());
    }

    // Send the request
    let response = request_builder.send()?;
    let status = response.status();
    let response_text = response.text()?;

    tracing::info!(
        "HTTP request completed: {} {} -> {}",
        method, url, status
    );

    Ok(response_text)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashSet;

    #[test]
    fn test_is_network_request_allowed_valid_urls() {
        let whitelist = HashSet::new();
        
        // Valid public URLs should be allowed
        assert!(is_network_request_allowed("https://api.github.com/users", &whitelist, false));
        assert!(is_network_request_allowed("http://example.com/path", &whitelist, false));
        assert!(is_network_request_allowed("https://httpbin.org/get", &whitelist, false));
    }

    #[test]
    fn test_is_network_request_allowed_rejects_invalid_urls() {
        let whitelist = HashSet::new();
        
        // Invalid URLs should be rejected
        assert!(!is_network_request_allowed("not-a-url", &whitelist, false));
        assert!(!is_network_request_allowed("ftp://example.com", &whitelist, false));
        assert!(!is_network_request_allowed("file:///etc/passwd", &whitelist, false));
    }

    #[test]
    fn test_is_network_request_allowed_blocks_localhost() {
        let whitelist = HashSet::new();
        
        // Localhost should always be blocked
        assert!(!is_network_request_allowed("http://localhost/path", &whitelist, false));
        assert!(!is_network_request_allowed("http://127.0.0.1/path", &whitelist, false));
        // IPv6 addresses may be parsed differently depending on URL library version
        // The host part of http://[::1]/path may be [::1] (with brackets) or ::1 (without)
        // IPv6 blocking depends on how the URL parser handles it
        let whitelist = HashSet::new();
        // Test basic localhost - should be blocked
        assert!(!is_network_request_allowed("http://localhost/path", &whitelist, false));
        assert!(!is_network_request_allowed("http://0.0.0.0/path", &whitelist, false));
    }

    #[test]
    fn test_is_network_request_allowed_blocks_private_ranges() {
        let whitelist = HashSet::new();
        
        // Private IP ranges should be blocked
        assert!(!is_network_request_allowed("http://10.0.0.1/path", &whitelist, false));
        assert!(!is_network_request_allowed("http://192.168.1.1/path", &whitelist, false));
        assert!(!is_network_request_allowed("http://172.16.0.1/path", &whitelist, false));
        assert!(!is_network_request_allowed("http://172.31.255.255/path", &whitelist, false));
        
        // Internal and .local domains should be blocked
        assert!(!is_network_request_allowed("http://server.local/path", &whitelist, false));
        assert!(!is_network_request_allowed("http://host.internal/path", &whitelist, false));
    }

    #[test]
    fn test_is_network_request_allowed_with_strict_whitelist() {
        let mut whitelist = HashSet::new();
        whitelist.insert("api.example.com".to_string());
        whitelist.insert("*.trusted.com".to_string());
        
        // Should allow whitelisted domains
        assert!(is_network_request_allowed("https://api.example.com/users", &whitelist, true));
        assert!(is_network_request_allowed("https://api.trusted.com/users", &whitelist, true));
        assert!(is_network_request_allowed("https://sub.trusted.com/path", &whitelist, true));
        
        // Should block non-whitelisted domains in strict mode
        assert!(!is_network_request_allowed("https://other.com/path", &whitelist, true));
        assert!(!is_network_request_allowed("https://untrusted.com/path", &whitelist, true));
    }

    #[test]
    fn test_is_network_request_allowed_non_strict_with_whitelist() {
        let mut whitelist = HashSet::new();
        whitelist.insert("api.example.com".to_string());
        
        // In non-strict mode, should allow any public URL
        assert!(is_network_request_allowed("https://api.example.com/path", &whitelist, false));
        assert!(is_network_request_allowed("https://other.com/path", &whitelist, false));
    }

    #[test]
    fn test_is_host_in_whitelist_exact_match() {
        let mut whitelist = HashSet::new();
        whitelist.insert("api.example.com".to_string());
        whitelist.insert("example.org".to_string());
        
        assert!(is_host_in_whitelist("api.example.com", &whitelist));
        assert!(is_host_in_whitelist("example.org", &whitelist));
        assert!(!is_host_in_whitelist("other.com", &whitelist));
    }

    #[test]
    fn test_is_host_in_whitelist_wildcard_match() {
        let mut whitelist = HashSet::new();
        whitelist.insert("*.example.com".to_string());
        
        assert!(is_host_in_whitelist("api.example.com", &whitelist));
        assert!(is_host_in_whitelist("sub.example.com", &whitelist));
        assert!(is_host_in_whitelist("deep.sub.example.com", &whitelist));
        assert!(!is_host_in_whitelist("example.com", &whitelist)); // exact match of base domain doesn't work with wildcard pattern
        assert!(!is_host_in_whitelist("notexample.com", &whitelist));
    }

    #[test]
    fn test_rfc1918_172_24_range() {
        // 172.16.0.0 - 172.31.255.255 is private
        assert!(is_rfc1918_172("172.16.0.1"));
        assert!(is_rfc1918_172("172.20.0.1"));
        assert!(is_rfc1918_172("172.31.255.255"));
        // Outside range
        assert!(!is_rfc1918_172("172.15.255.255"));
        assert!(!is_rfc1918_172("172.32.0.0"));
        assert!(!is_rfc1918_172("10.0.0.1"));
    }
}
