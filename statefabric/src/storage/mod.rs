//! Storage abstraction layer for object storage

pub mod api_keys;
mod object_store;
mod postgres;

pub use api_keys::*;
pub use object_store::*;
pub use postgres::*;

// Re-exports for use in AppState and main.rs
pub use object_store::{ObjectStoreMemory, create_object_store, EncryptedObjectStore, StorageConfig, StorageBackend};

// Re-export argon2 functions for use in middleware
pub use api_keys::{hash_api_key_argon2, verify_api_key_argon2};
