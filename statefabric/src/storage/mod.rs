//! Storage abstraction layer for object storage

mod object_store;
mod postgres;

pub use object_store::*;
pub use postgres::*;
