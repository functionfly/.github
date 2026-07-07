//! API handlers for StateFabric

mod routes;
mod handlers;
pub mod auth;        // JWT and API key validation
pub mod middleware;  // Auth middleware and rate limiting

pub use routes::*;
pub use handlers::*;
