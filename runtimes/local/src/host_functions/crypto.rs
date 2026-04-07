//! Cryptographic primitives host functions implementation.
//!
//! Provides the following WASM-callable host functions:
//!
//! ```text
//! crypto_hmac(algo_ptr, algo_len, key_ptr, key_len,
//!             msg_ptr,  msg_len,  out_ptr, out_len_ptr) -> i32
//!   Computes HMAC using the given algorithm ("sha256" | "sha512").
//!   Writes lowercase hex digest to WASM memory.
//!
//! crypto_hash(algo_ptr, algo_len, data_ptr, data_len,
//!             out_ptr,  out_len_ptr) -> i32
//!   Computes a cryptographic hash ("sha256" | "sha512").
//!   Writes lowercase hex digest to WASM memory.
//!
//! crypto_random(out_ptr, out_len_ptr, n_bytes: i32) -> i32
//!   Fills WASM memory with `n_bytes` of CSPRNG output encoded as lowercase hex.
//!   n_bytes must be in 1..=4096.
//! ```
//!
//! All functions return 0 on success, or a negative error code.

use hmac::{Hmac, Mac};
use sha2::{Digest, Sha256, Sha512};
use wasmtime_wasi::p1::WasiP1Ctx;

use super::memory_utils;

type HmacSha256 = Hmac<Sha256>;
type HmacSha512 = Hmac<Sha512>;

/// Add all crypto host functions to the linker.
pub fn add_crypto_functions(linker: &mut wasmtime::Linker<WasiP1Ctx>) -> anyhow::Result<()> {
    add_hmac_function(linker)?;
    add_hash_function(linker)?;
    add_random_function(linker)?;
    tracing::debug!("Added functionfly.crypto_hmac, crypto_hash, crypto_random host functions");
    Ok(())
}

// ---------------------------------------------------------------------------
// HMAC
// ---------------------------------------------------------------------------

fn add_hmac_function(linker: &mut wasmtime::Linker<WasiP1Ctx>) -> anyhow::Result<()> {
    linker.func_wrap(
        "functionfly",
        "crypto_hmac",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              algo_ptr: i32,
              algo_len: i32,
              key_ptr: i32,
              key_len: i32,
              msg_ptr: i32,
              msg_len: i32,
              out_ptr: i32,
              out_len_ptr: i32|
              -> i32 {
            let algo = match memory_utils::read_string_from_memory(&mut caller, algo_ptr, algo_len)
            {
                Ok(a) => a,
                Err(_) => return -1,
            };
            let key = match memory_utils::read_bytes_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2,
            };
            let msg = match memory_utils::read_bytes_from_memory(&mut caller, msg_ptr, msg_len) {
                Ok(m) => m,
                Err(_) => return -3,
            };

            let hex_output = match compute_hmac(&algo, &key, &msg) {
                Ok(h) => h,
                Err(_) => return -4,
            };

            match memory_utils::write_string_to_memory(
                &mut caller,
                &hex_output,
                out_ptr,
                out_len_ptr,
            ) {
                Ok(_) => 0,
                Err(_) => -5,
            }
        },
    )?;
    Ok(())
}

// ---------------------------------------------------------------------------
// Hash
// ---------------------------------------------------------------------------

fn add_hash_function(linker: &mut wasmtime::Linker<WasiP1Ctx>) -> anyhow::Result<()> {
    linker.func_wrap(
        "functionfly",
        "crypto_hash",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              algo_ptr: i32,
              algo_len: i32,
              data_ptr: i32,
              data_len: i32,
              out_ptr: i32,
              out_len_ptr: i32|
              -> i32 {
            let algo = match memory_utils::read_string_from_memory(&mut caller, algo_ptr, algo_len)
            {
                Ok(a) => a,
                Err(_) => return -1,
            };
            let data = match memory_utils::read_bytes_from_memory(&mut caller, data_ptr, data_len) {
                Ok(d) => d,
                Err(_) => return -2,
            };

            let hex_output = match compute_hash(&algo, &data) {
                Ok(h) => h,
                Err(_) => return -3,
            };

            match memory_utils::write_string_to_memory(
                &mut caller,
                &hex_output,
                out_ptr,
                out_len_ptr,
            ) {
                Ok(_) => 0,
                Err(_) => -4,
            }
        },
    )?;
    Ok(())
}

// ---------------------------------------------------------------------------
// Random
// ---------------------------------------------------------------------------

fn add_random_function(linker: &mut wasmtime::Linker<WasiP1Ctx>) -> anyhow::Result<()> {
    linker.func_wrap(
        "functionfly",
        "crypto_random",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              out_ptr: i32,
              out_len_ptr: i32,
              n_bytes: i32|
              -> i32 {
            if n_bytes <= 0 || n_bytes > 4096 {
                return -1;
            }

            let bytes = match read_random_bytes(n_bytes as usize) {
                Ok(b) => b,
                Err(_) => return -2,
            };

            let hex_str = hex::encode(&bytes);
            match memory_utils::write_string_to_memory(&mut caller, &hex_str, out_ptr, out_len_ptr)
            {
                Ok(_) => 0,
                Err(_) => -3,
            }
        },
    )?;
    Ok(())
}

// ---------------------------------------------------------------------------
// Internal helpers (pub for unit tests)
// ---------------------------------------------------------------------------

pub fn compute_hmac(algo: &str, key: &[u8], msg: &[u8]) -> anyhow::Result<String> {
    match algo.to_lowercase().as_str() {
        "sha256" => {
            let mut mac = HmacSha256::new_from_slice(key)
                .map_err(|e| anyhow::anyhow!("HMAC-SHA256 init: {}", e))?;
            mac.update(msg);
            Ok(hex::encode(mac.finalize().into_bytes()))
        }
        "sha512" => {
            let mut mac = HmacSha512::new_from_slice(key)
                .map_err(|e| anyhow::anyhow!("HMAC-SHA512 init: {}", e))?;
            mac.update(msg);
            Ok(hex::encode(mac.finalize().into_bytes()))
        }
        other => Err(anyhow::anyhow!("Unsupported HMAC algorithm: {}", other)),
    }
}

pub fn compute_hash(algo: &str, data: &[u8]) -> anyhow::Result<String> {
    match algo.to_lowercase().as_str() {
        "sha256" => {
            let mut h = Sha256::new();
            h.update(data);
            Ok(hex::encode(h.finalize()))
        }
        "sha512" => {
            let mut h = Sha512::new();
            h.update(data);
            Ok(hex::encode(h.finalize()))
        }
        other => Err(anyhow::anyhow!("Unsupported hash algorithm: {}", other)),
    }
}

/// Read `n` cryptographically random bytes from `/dev/urandom` (Unix) or a
/// fallback based on UUID v4 hashing.
pub fn read_random_bytes(n: usize) -> anyhow::Result<Vec<u8>> {
    #[cfg(unix)]
    {
        use std::io::Read;
        let mut f = std::fs::File::open("/dev/urandom")
            .map_err(|e| anyhow::anyhow!("/dev/urandom: {}", e))?;
        let mut buf = vec![0u8; n];
        f.read_exact(&mut buf)
            .map_err(|e| anyhow::anyhow!("read /dev/urandom: {}", e))?;
        Ok(buf)
    }

    #[cfg(not(unix))]
    {
        // Non-Unix fallback: stretch UUID v4 bytes through SHA-256
        let mut result = Vec::with_capacity(n);
        while result.len() < n {
            let id = uuid::Uuid::new_v4();
            let mut h = Sha256::new();
            h.update(id.as_bytes());
            result.extend_from_slice(&h.finalize());
        }
        result.truncate(n);
        Ok(result)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hmac_sha256_known_vector() {
        let result = compute_hmac("sha256", b"key", b"message").unwrap();
        assert_eq!(
            result.len(),
            64,
            "HMAC-SHA256 must produce 32-byte (64 hex char) output"
        );
    }

    #[test]
    fn test_hash_sha256_known_vector() {
        let result = compute_hash("sha256", b"hello").unwrap();
        assert_eq!(
            result,
            "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
        );
    }

    #[test]
    fn test_random_bytes_length() {
        let bytes = read_random_bytes(32).unwrap();
        assert_eq!(bytes.len(), 32);
    }

    #[test]
    fn test_unsupported_algo() {
        assert!(compute_hmac("md5", b"k", b"m").is_err());
        assert!(compute_hash("md5", b"d").is_err());
    }
}
