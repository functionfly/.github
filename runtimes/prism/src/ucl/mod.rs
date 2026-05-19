//! Universal Capability Layer (UCL)
//!
//! Instead of exposing raw APIs, everything becomes a discoverable capability:
//! ```json
//! {
//!   "capability": "vision.detect",
//!   "latency": "12ms",
//!   "trust": 0.998,
//!   "gpu_required": true
//! }
//! ```
//!
//! This enables robots, AI agents, browsers, drones, SaaS apps, IDEs
//! to all discover compatible functions dynamically.
//!
//! Think of it as "The DNS of AI capabilities."

pub mod registry;
pub mod discovery;
pub mod capability;
pub mod matcher;

pub use registry::{CapabilityRegistry, RegistryConfig};
pub use discovery::{CapabilityDiscovery, DiscoveryQuery, DiscoveryResult};
pub use capability::{Capability, CapabilityId, CapabilityCategory, TrustLevel};
pub use matcher::{CapabilityMatcher, MatchScore};