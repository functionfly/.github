//! Native Module Restrictions
//! 
//! This module provides lists of allowed and blocked Node.js native modules
//! for security purposes.

use std::collections::{HashMap, HashSet};

/// Native module categories
#[derive(Debug, Clone)]
pub enum ModuleCategory {
    /// Fully allowed
    Allowed,
    /// Allowed with limited functionality
    Limited(String),  // Reason for limitation
    /// Blocked for security
    Blocked(String),   // Reason for blocking
}

/// Native module configuration
pub struct NativeModules {
    /// Module categories
    categories: HashMap<String, ModuleCategory>,
}

impl NativeModules {
    /// Create a new native modules configuration
    pub fn new() -> Self {
        let mut categories = HashMap::new();
        
        // === BLOCKED MODULES (Security) ===
        categories.insert("child_process".to_string(), ModuleCategory::Blocked(
            "Allows arbitrary command execution".to_string()
        ));
        categories.insert("fs".to_string(), ModuleCategory::Blocked(
            "Allows file system access".to_string()
        ));
        categories.insert("net".to_string(), ModuleCategory::Blocked(
            "Allows network connections".to_string()
        ));
        categories.insert("tls".to_string(), ModuleCategory::Blocked(
            "Allows TLS connections".to_string()
        ));
        categories.insert("http".to_string(), ModuleCategory::Blocked(
            "Allows HTTP requests (use fetch instead)".to_string()
        ));
        categories.insert("https".to_string(), ModuleCategory::Blocked(
            "Allows HTTPS requests (use fetch instead)".to_string()
        ));
        categories.insert("dns".to_string(), ModuleCategory::Blocked(
            "Allows DNS lookups".to_string()
        ));
        categories.insert("dgram".to_string(), ModuleCategory::Blocked(
            "Allows UDP connections".to_string()
        ));
        categories.insert("repl".to_string(), ModuleCategory::Blocked(
            "Allows interactive console".to_string()
        ));
        categories.insert("vm".to_string(), ModuleCategory::Blocked(
            "Allows code compilation".to_string()
        ));
        categories.insert("worker_threads".to_string(), ModuleCategory::Blocked(
            "Allows multi-threading".to_string()
        ));
        categories.insert("perf_hooks".to_string(), ModuleCategory::Blocked(
            "Allows performance monitoring".to_string()
        ));
        categories.insert("inspector".to_string(), ModuleCategory::Blocked(
            "Allows debugging".to_string()
        ));
        categories.insert("readline".to_string(), ModuleCategory::Blocked(
            "Allows interactive input".to_string()
        ));
        categories.insert("zlib".to_string(), ModuleCategory::Blocked(
            "Allows compression (use built-in compression instead)".to_string()
        ));
        
        // === LIMITED MODULES ===
        categories.insert("console".to_string(), ModuleCategory::Limited(
            "Only console.log, console.error, console.warn allowed".to_string()
        ));
        categories.insert("crypto".to_string(), ModuleCategory::Limited(
            "Only hashing functions allowed (sha256, sha512, md5)".to_string()
        ));
        categories.insert("stream".to_string(), ModuleCategory::Limited(
            "Only Readable and Writable streams allowed".to_string()
        ));
        categories.insert("buffer".to_string(), ModuleCategory::Limited(
            "Only Buffer.from, Buffer.alloc allowed".to_string()
        ));
        categories.insert("util".to_string(), ModuleCategory::Limited(
            "Only util.types and util.format allowed".to_string()
        ));
        categories.insert("path".to_string(), ModuleCategory::Limited(
            "Only path.join, path.resolve, path.basename allowed".to_string()
        ));
        categories.insert("url".to_string(), ModuleCategory::Allowed);
        categories.insert("querystring".to_string(), ModuleCategory::Allowed);
        
        // === ALLOWED MODULES ===
        categories.insert("json".to_string(), ModuleCategory::Allowed);
        categories.insert("Math".to_string(), ModuleCategory::Allowed);
        categories.insert("Date".to_string(), ModuleCategory::Allowed);
        categories.insert("Array".to_string(), ModuleCategory::Allowed);
        categories.insert("Object".to_string(), ModuleCategory::Allowed);
        categories.insert("String".to_string(), ModuleCategory::Allowed);
        categories.insert("Number".to_string(), ModuleCategory::Allowed);
        categories.insert("Boolean".to_string(), ModuleCategory::Allowed);
        categories.insert("Promise".to_string(), ModuleCategory::Allowed);
        categories.insert("Map".to_string(), ModuleCategory::Allowed);
        categories.insert("Set".to_string(), ModuleCategory::Allowed);
        categories.insert("WeakMap".to_string(), ModuleCategory::Allowed);
        categories.insert("WeakSet".to_string(), ModuleCategory::Allowed);
        categories.insert("Proxy".to_string(), ModuleCategory::Allowed);
        categories.insert("Reflect".to_string(), ModuleCategory::Allowed);
        categories.insert("BigInt".to_string(), ModuleCategory::Allowed);
        categories.insert("Symbol".to_string(), ModuleCategory::Allowed);
        categories.insert("Error".to_string(), ModuleCategory::Allowed);
        categories.insert("RegExp".to_string(), ModuleCategory::Allowed);
        
        Self { categories }
    }

    /// Check if a module is allowed
    pub fn is_allowed(&self, module: &str) -> bool {
        match self.categories.get(module) {
            Some(ModuleCategory::Allowed) => true,
            Some(ModuleCategory::Limited(_)) => true,
            Some(ModuleCategory::Blocked(_)) => false,
            None => false,  // Unknown modules are blocked
        }
    }

    /// Check if a module is blocked
    pub fn is_blocked(&self, module: &str) -> bool {
        matches!(
            self.categories.get(module),
            Some(ModuleCategory::Blocked(_))
        )
    }

    /// Get the reason for blocking/limiting a module
    pub fn get_reason(&self, module: &str) -> Option<String> {
        match self.categories.get(module) {
            Some(ModuleCategory::Blocked(reason)) => Some(reason.clone()),
            Some(ModuleCategory::Limited(reason)) => Some(reason.clone()),
            _ => None,
        }
    }

    /// Get all allowed modules
    pub fn allowed_modules(&self) -> Vec<String> {
        self.categories
            .iter()
            .filter_map(|(name, cat)| {
                match cat {
                    ModuleCategory::Allowed => Some(name.clone()),
                    ModuleCategory::Limited(_) => Some(name.clone()),
                    _ => None,
                }
            })
            .collect()
    }

    /// Get all blocked modules
    pub fn blocked_modules(&self) -> Vec<(String, String)> {
        self.categories
            .iter()
            .filter_map(|(name, cat)| {
                match cat {
                    ModuleCategory::Blocked(reason) => Some((name.clone(), reason.clone())),
                    _ => None,
                }
            })
            .collect()
    }

    /// Default configuration
    pub fn default() -> Self {
        Self::new()
    }
}

impl Default for NativeModules {
    fn default() -> Self {
        Self::new()
    }
}
