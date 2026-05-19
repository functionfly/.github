//! CBOR Codec module for Prism Runtime
//!
//! Provides CBOR encoding/decoding using the ciborium crate.
//! CBOR is used for efficient binary serialization in distributed systems.

use serde::{Deserialize, Serialize};
use thiserror::Error;

use crate::core::PrismError;

/// CBOR encoding/decoding errors
#[derive(Error, Debug)]
pub enum CodecError {
    #[error("CBOR serialization error: {0}")]
    Serialization(#[from] ciborium::ser::Error<std::io::Error>),
    #[error("CBOR deserialization error: {0}")]
    Deserialization(#[from] ciborium::de::Error<std::io::Error>),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("Invalid tag: {0}")]
    InvalidTag(String),
    #[error("Prism error: {0}")]
    Prism(#[from] PrismError),
}

/// CBOR codec for encoding/decoding types
pub struct CborCodec;

impl CborCodec {
    /// Encode a value to CBOR bytes
    pub fn encode<T: Serialize>(value: &T) -> Result<Vec<u8>, CodecError> {
        let mut bytes = Vec::new();
        ciborium::ser::into_writer(value, &mut bytes)?;
        Ok(bytes)
    }

    /// Decode CBOR bytes to a value
    pub fn decode<T: for<'de> Deserialize<'de>>(bytes: &[u8]) -> Result<T, CodecError> {
        let value: T = ciborium::from_reader(bytes)?;
        Ok(value)
    }

    /// Encode a value to CBOR bytes in a boxed format suitable for storage
    pub fn encode_boxed<T: Serialize>(value: &T) -> Result<Box<[u8]>, CodecError> {
        let bytes = Self::encode(value)?;
        Ok(bytes.into_boxed_slice())
    }

    /// Get the encoded size in bytes
    pub fn encoded_size<T: Serialize>(value: &T) -> Result<usize, CodecError> {
        Ok(Self::encode(value)?.len())
    }
}

/// Tagged CBOR value for type identification
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaggedValue<T> {
    pub tag: String,
    pub value: T,
}

impl<T> TaggedValue<T> {
    pub fn new(tag: impl Into<String>, value: T) -> Self {
        Self {
            tag: tag.into(),
            value,
        }
    }
}

/// CBOR map with string keys
pub type CborMap<K, V> = std::collections::HashMap<K, V>;

/// CBOR list
pub type CborList<T> = Vec<T>;

/// Null-terminated CBOR string
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CborString {
    pub value: String,
}

impl CborString {
    pub fn new(value: impl Into<String>) -> Self {
        Self { value: value.into() }
    }
}

impl AsRef<str> for CborString {
    fn as_ref(&self) -> &str {
        &self.value
    }
}

impl From<String> for CborString {
    fn from(value: String) -> Self {
        Self::new(value)
    }
}

impl From<&str> for CborString {
    fn from(value: &str) -> Self {
        Self::new(value)
    }
}

/// CBOR bytes (wrapped for serialization)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CborBytes {
    pub value: Vec<u8>,
}

impl CborBytes {
    pub fn new(value: Vec<u8>) -> Self {
        Self { value }
    }

    pub fn as_slice(&self) -> &[u8] {
        &self.value
    }
}

impl AsRef<[u8]> for CborBytes {
    fn as_ref(&self) -> &[u8] {
        &self.value
    }
}

impl From<Vec<u8>> for CborBytes {
    fn from(value: Vec<u8>) -> Self {
        Self::new(value)
    }
}

impl From<&[u8]> for CborBytes {
    fn from(value: &[u8]) -> Self {
        Self::new(value.to_vec())
    }
}

/// CBOR encoding options
#[derive(Debug, Clone)]
pub struct EncodeOptions {
    /// Whether to canonicalize the output (sort keys)
    pub canonical: bool,
    /// Maximum depth for nested structures
    pub max_depth: usize,
}

impl Default for EncodeOptions {
    fn default() -> Self {
        Self {
            canonical: false,
            max_depth: 64,
        }
    }
}

/// CBOR decoding options
#[derive(Debug, Clone)]
pub struct DecodeOptions {
    /// Whether to allow invalid UTF-8 strings
    pub allow_invalid_utf8: bool,
    /// Maximum depth for nested structures
    pub max_depth: usize,
}

impl Default for DecodeOptions {
    fn default() -> Self {
        Self {
            allow_invalid_utf8: false,
            max_depth: 64,
        }
    }
}

/// Streaming CBOR decoder for large values
pub struct CborStream<R: std::io::Read> {
    reader: R,
}

impl<R: std::io::Read> CborStream<R> {
    pub fn new(reader: R) -> Self {
        Self { reader }
    }

    /// Read the next CBOR value as a boxed trait object
    pub fn next_value(&mut self) -> Result<Option<ciborium::Value>, CodecError> {
        // Try to read one value
        let result: Result<ciborium::Value, _> = ciborium::from_reader(&mut self.reader);
        match result {
            Ok(value) => Ok(Some(value)),
            Err(ciborium::de::Error::Io(ref e)) if e.kind() == std::io::ErrorKind::UnexpectedEof => {
                Ok(None)
            }
            Err(e) => Err(CodecError::Deserialization(e)),
        }
    }
}

/// Helper to convert a type to CBOR hex string for debugging
pub fn to_hex<T: Serialize>(value: &T) -> Result<String, CodecError> {
    let bytes = CborCodec::encode(value)?;
    Ok(bytes.iter().map(|b| format!("{:02x}", b)).collect())
}

/// Helper to parse a CBOR hex string
pub fn from_hex<T: for<'de> Deserialize<'de>>(hex: &str) -> Result<T, CodecError> {
    let bytes: Vec<u8> = hex
        .as_bytes()
        .chunks(2)
        .map(|chunk| {
            let s = std::str::from_utf8(chunk).unwrap();
            u8::from_str_radix(s, 16).unwrap()
        })
        .collect();
    CborCodec::decode(&bytes)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encode_decode_basic() {
        let value = "hello world";
        let encoded = CborCodec::encode(&value).unwrap();
        let decoded: String = CborCodec::decode(&encoded).unwrap();
        assert_eq!(value, decoded);
    }

    #[test]
    fn test_encode_decode_map() {
        use std::collections::HashMap;
        let mut map = HashMap::new();
        map.insert("key1".to_string(), 42u32);
        map.insert("key2".to_string(), 100u32);

        let encoded = CborCodec::encode(&map).unwrap();
        let decoded: HashMap<String, u32> = CborCodec::decode(&encoded).unwrap();
        assert_eq!(map, decoded);
    }

    #[test]
    fn test_encode_decode_nested() {
        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Inner {
            value: i32,
        }

        #[derive(Debug, Serialize, Deserialize, PartialEq)]
        struct Outer {
            name: String,
            inner: Inner,
            items: Vec<u64>,
        }

        let outer = Outer {
            name: "test".to_string(),
            inner: Inner { value: -10 },
            items: vec![1, 2, 3, 4, 5],
        };

        let encoded = CborCodec::encode(&outer).unwrap();
        let decoded: Outer = CborCodec::decode(&encoded).unwrap();
        assert_eq!(outer, decoded);
    }

    #[test]
    fn test_cbor_bytes() {
        let original = vec![0x01, 0x02, 0x03, 0x04];
        let cb = CborBytes::new(original.clone());
        let encoded = CborCodec::encode(&cb).unwrap();
        let decoded: CborBytes = CborCodec::decode(&encoded).unwrap();
        assert_eq!(original, decoded.as_ref());
    }

    #[test]
    fn test_tagged_value() {
        let tagged = TaggedValue::new("my-type", vec![1u8, 2, 3]);
        let encoded = CborCodec::encode(&tagged).unwrap();
        let decoded: TaggedValue<Vec<u8>> = CborCodec::decode(&encoded).unwrap();
        assert_eq!(tagged.tag, decoded.tag);
        assert_eq!(tagged.value, decoded.value);
    }

    #[test]
    fn test_to_hex() {
        let value = 42u32;
        let hex = to_hex(&value).unwrap();
        assert_eq!(hex, "18 2a".replace(" ", "")); // 0x18 0x2a = 42 in CBOR
    }

    #[test]
    fn test_encode_size() {
        let value = "small";
        let size = CborCodec::encoded_size(&value).unwrap();
        assert_eq!(size, CborCodec::encode(&value).unwrap().len());
    }
}
