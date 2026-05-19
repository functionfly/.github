//! Compression utilities for Quantum Snapshotting
//!
//! Uses zstd (level 3) for general-purpose compression and lz4 for
//! speed-critical paths.  Both crates are direct dependencies of the
//! prism runtime.

use crate::core::PrismResult;
use super::snapshot::CompressionAlgorithm;

/// Compress snapshot data using the specified algorithm.
pub fn compress_snapshot(
    data: &[u8],
    algorithm: CompressionAlgorithm,
) -> PrismResult<Vec<u8>> {
    match algorithm {
        CompressionAlgorithm::None => Ok(data.to_vec()),
        CompressionAlgorithm::Zstd => zstd_encode(data),
        CompressionAlgorithm::Lz4 => lz4_encode(data),
    }
}

/// Decompress snapshot data using the specified algorithm.
pub fn decompress_snapshot(
    data: &[u8],
    algorithm: CompressionAlgorithm,
) -> PrismResult<Vec<u8>> {
    match algorithm {
        CompressionAlgorithm::None => Ok(data.to_vec()),
        CompressionAlgorithm::Zstd => zstd_decode(data),
        CompressionAlgorithm::Lz4 => lz4_decode(data),
    }
}

// ── zstd ─────────────────────────────────────────────────────────────

/// Compress with zstd at level 3 (good balance of ratio and speed).
fn zstd_encode(data: &[u8]) -> PrismResult<Vec<u8>> {
    zstd::encode_all(data, 3).map_err(|e| {
        crate::core::PrismError::Internal(format!("zstd compression failed: {}", e))
    })
}

/// Decompress zstd-compressed data.
fn zstd_decode(data: &[u8]) -> PrismResult<Vec<u8>> {
    zstd::decode_all(data).map_err(|e| {
        crate::core::PrismError::Internal(format!("zstd decompression failed: {}", e))
    })
}

// ── lz4 ──────────────────────────────────────────────────────────────

/// Compress with lz4 in high-compression mode (level 3).
fn lz4_encode(data: &[u8]) -> PrismResult<Vec<u8>> {
    lz4::block::compress(data, Some(lz4::block::CompressionMode::HIGHCOMPRESSION(3)), true)
        .map_err(|e| crate::core::PrismError::Internal(format!("lz4 compression failed: {}", e)))
}

/// Decompress lz4-compressed data.
fn lz4_decode(data: &[u8]) -> PrismResult<Vec<u8>> {
    lz4::block::decompress(data, None)
        .map_err(|e| crate::core::PrismError::Internal(format!("lz4 decompression failed: {}", e)))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_zstd_roundtrip() {
        let original = b"Hello, world! This is a test of zstd compression. ".repeat(100);
        let compressed = compress_snapshot(original.as_slice(), CompressionAlgorithm::Zstd).unwrap();
        assert!(compressed.len() < original.len(), "zstd should compress repetitive data");
        let decompressed = decompress_snapshot(&compressed, CompressionAlgorithm::Zstd).unwrap();
        assert_eq!(original.as_slice(), decompressed.as_slice());
    }

    #[test]
    fn test_lz4_roundtrip() {
        let original = b"LZ4 compression test data. ".repeat(50);
        let compressed = compress_snapshot(original.as_slice(), CompressionAlgorithm::Lz4).unwrap();
        let decompressed = decompress_snapshot(&compressed, CompressionAlgorithm::Lz4).unwrap();
        assert_eq!(original.as_slice(), decompressed.as_slice());
    }

    #[test]
    fn test_none_roundtrip() {
        let original = b"no compression";
        let result = compress_snapshot(original, CompressionAlgorithm::None).unwrap();
        assert_eq!(original.as_slice(), result.as_slice());
    }

    #[test]
    fn test_zstd_empty_input() {
        let compressed = compress_snapshot(b"", CompressionAlgorithm::Zstd).unwrap();
        let decompressed = decompress_snapshot(&compressed, CompressionAlgorithm::Zstd).unwrap();
        assert!(decompressed.is_empty());
    }
}
