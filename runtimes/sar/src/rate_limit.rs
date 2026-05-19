use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use parking_lot::RwLock;
use tokio::sync::mpsc;

#[derive(Clone)]
pub struct RateLimiter {
    global_limit: RateLimitConfig,
    per_agent_limit: RateLimitConfig,
    global_buckets: Arc<RwLock<HashMap<String, TokenBucket>>>,
    agent_buckets: Arc<RwLock<HashMap<String, AgentRateLimit>>>,
}

#[derive(Clone, Copy, Debug)]
pub struct RateLimitConfig {
    pub requests_per_second: u64,
    pub burst_size: u64,
    pub max_queue_size: usize,
}

impl Default for RateLimitConfig {
    fn default() -> Self {
        Self {
            requests_per_second: 1000,
            burst_size: 100,
            max_queue_size: 10000,
        }
    }
}

impl RateLimitConfig {
    pub fn strict() -> Self {
        Self {
            requests_per_second: 100,
            burst_size: 20,
            max_queue_size: 1000,
        }
    }

    pub fn permissive() -> Self {
        Self {
            requests_per_second: 10_000,
            burst_size: 1000,
            max_queue_size: 100_000,
        }
    }
}

#[derive(Clone)]
struct TokenBucket {
    tokens: f64,
    last_refill: Instant,
    capacity: f64,
    refill_rate: f64,
}

impl TokenBucket {
    fn new(capacity: u64, refill_rate: f64) -> Self {
        Self {
            tokens: capacity as f64,
            last_refill: Instant::now(),
            capacity: capacity as f64,
            refill_rate,
        }
    }

    fn try_consume(&mut self, tokens: u64) -> bool {
        self.refill();
        if self.tokens >= tokens as f64 {
            self.tokens -= tokens as f64;
            true
        } else {
            false
        }
    }

    fn refill(&mut self) {
        let now = Instant::now();
        let elapsed = now.duration_since(self.last_refill).as_secs_f64();
        let tokens_to_add = elapsed * self.refill_rate;
        self.tokens = self.tokens.min(self.capacity) + tokens_to_add;
        self.last_refill = now;
    }

    fn available_tokens(&self) -> f64 {
        let mut bucket = self.clone();
        bucket.refill();
        bucket.tokens
    }
}

#[derive(Clone)]
struct AgentRateLimit {
    buckets: HashMap<String, TokenBucket>,
}

impl AgentRateLimit {
    fn new(capacity: u64, refill_rate: f64) -> Self {
        let mut buckets = HashMap::new();
        buckets.insert("default".to_string(), TokenBucket::new(capacity, refill_rate));
        Self { buckets }
    }

    fn try_consume(&mut self, operation: &str, tokens: u64) -> bool {
        let bucket = self.buckets.entry(operation.to_string()).or_insert_with(|| {
            TokenBucket::new(50, 10.0)
        });
        bucket.try_consume(tokens)
    }
}

#[derive(Debug, Clone)]
pub enum RateLimitResult {
    Allowed,
    RateLimited { retry_after_ms: u64 },
    QueueFull,
}

impl RateLimiter {
    pub fn new(global: RateLimitConfig, per_agent: RateLimitConfig) -> Self {
        Self {
            global_limit: global,
            per_agent_limit: per_agent,
            global_buckets: Arc::new(RwLock::new(HashMap::new())),
            agent_buckets: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn default_rate_limiter() -> Self {
        Self::new(RateLimitConfig::default(), RateLimitConfig::default())
    }

    pub fn check_global(&self, key: &str) -> RateLimitResult {
        let mut buckets = self.global_buckets.write();

        let bucket = buckets.entry(key.to_string()).or_insert_with(|| {
            TokenBucket::new(
                self.global_limit.burst_size,
                self.global_limit.requests_per_second as f64,
            )
        });

        if bucket.try_consume(1) {
            RateLimitResult::Allowed
        } else {
            let wait_ms = ((1.0 - bucket.available_tokens()) / bucket.refill_rate * 1000.0) as u64;
            RateLimitResult::RateLimited {
                retry_after_ms: wait_ms.max(1),
            }
        }
    }

    pub fn check_agent(&self, agent_id: &str, operation: &str) -> RateLimitResult {
        let mut buckets = self.agent_buckets.write();

        let agent_limit = buckets.entry(agent_id.to_string()).or_insert_with(|| {
            AgentRateLimit::new(
                self.per_agent_limit.burst_size,
                self.per_agent_limit.requests_per_second as f64,
            )
        });

        if agent_limit.try_consume(operation, 1) {
            RateLimitResult::Allowed
        } else {
            RateLimitResult::RateLimited {
                retry_after_ms: 100,
            }
        }
    }

    pub fn check(&self, agent_id: &str, operation: &str) -> RateLimitResult {
        match self.check_global("global") {
            RateLimitResult::Allowed => {}
            other => return other,
        }

        let key = format!("{}:{}", agent_id, operation);
        match self.check_agent(&key, operation) {
            RateLimitResult::Allowed => RateLimitResult::Allowed,
            other => other,
        }
    }

    pub fn get_queue_depth(&self) -> usize {
        self.agent_buckets.read().len()
    }

    pub fn get_global_tokens(&self) -> f64 {
        self.global_buckets
            .read()
            .get("global")
            .map(|b| b.available_tokens())
            .unwrap_or(0.0)
    }
}

pub struct RateLimitResultExt {
    pub allowed: bool,
    pub retry_after_ms: Option<u64>,
    pub queue_position: Option<usize>,
}

impl From<RateLimitResult> for RateLimitResultExt {
    fn from(result: RateLimitResult) -> Self {
        match result {
            RateLimitResult::Allowed => Self {
                allowed: true,
                retry_after_ms: None,
                queue_position: None,
            },
            RateLimitResult::RateLimited { retry_after_ms } => Self {
                allowed: false,
                retry_after_ms: Some(retry_after_ms),
                queue_position: None,
            },
            RateLimitResult::QueueFull => Self {
                allowed: false,
                retry_after_ms: None,
                queue_position: None,
            },
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_token_bucket_basic() {
        let mut bucket = TokenBucket::new(10, 5.0);
        assert!(bucket.try_consume(5));
        assert!(bucket.try_consume(6)); // Should fail, only 5 left
        assert!(bucket.try_consume(3)); // Should succeed after refill
    }

    #[test]
    fn test_rate_limiter_global() {
        let limiter = RateLimiter::default_rate_limiter();
        let result = limiter.check_global("test");
        assert!(matches!(result, RateLimitResult::Allowed));
    }

    #[test]
    fn test_rate_limit_config_presets() {
        let strict = RateLimitConfig::strict();
        assert_eq!(strict.requests_per_second, 100);
        assert_eq!(strict.burst_size, 20);

        let permissive = RateLimitConfig::permissive();
        assert_eq!(permissive.requests_per_second, 10_000);
        assert_eq!(permissive.burst_size, 1000);
    }
}