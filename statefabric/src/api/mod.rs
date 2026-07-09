//! API handlers for StateFabric

mod routes;
mod handlers;
pub mod auth;        // JWT and API key validation
pub mod middleware;  // Auth middleware, rate limiting, audit logging
pub mod metrics;     // Prometheus metrics

pub use routes::*;
pub use handlers::*;
pub use metrics::*;
pub use routes::create_router_with_repo;

// Re-export security types for use in other modules
pub use middleware::{
    AuditEvent, AuditLogger, PostgresAuditLogger, TracingAuditLogger, FailedLoginTracker,
    RedisRateLimiter, security_headers_middleware,
};
