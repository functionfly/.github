use std::collections::HashMap;

pub const MAX_NAME_LENGTH: usize = 256;
pub const MAX_METADATA_ENTRIES: usize = 64;
pub const MAX_METADATA_VALUE_LENGTH: usize = 1024;
pub const MAX_GRAPH_NODES: usize = 1000;
pub const MAX_GRAPH_EDGES: usize = 5000;
pub const MAX_INPUT_ENTRIES: usize = 100;
pub const MAX_INPUT_VALUE_SIZE: usize = 64 * 1024;
pub const MAX_TOTAL_INPUT_SIZE: usize = 1024 * 1024;
pub const MAX_CONCURRENT_CELLS: usize = 10000;
pub const MAX_GRACE_PERIOD_SECONDS: u64 = 3600;

pub const DEFAULT_MAX_CONCURRENT_CELLS: usize = 100;
pub const DEFAULT_PRIORITY: u8 = 2;

#[derive(Debug, Clone)]
pub enum ValidationError {
    NameTooLong { max: usize, actual: usize },
    NameEmpty,
    MetadataTooLarge { max: usize, actual: usize },
    MetadataValueTooLong { max: usize, actual: usize, key: String },
    GraphTooLarge { max_nodes: usize, actual_nodes: usize },
    GraphTooManyEdges { max: usize, actual: usize },
    InputTooLarge { max: usize, actual: usize },
    InputValueTooLarge { key: String, max: usize, actual: usize },
    TooManyInputEntries { max: usize, actual: usize },
    InvalidPriority { valid_range: String },
    InvalidMaxConcurrentCells { max: usize, actual: usize },
    InvalidGracePeriod { max_seconds: u64, actual_seconds: u64 },
    InvalidUuid { value: String },
    EmptyGraph,
}

impl std::fmt::Display for ValidationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::NameTooLong { max, actual } => {
                write!(f, "name too long: {} bytes (max {})", actual, max)
            }
            Self::NameEmpty => write!(f, "name cannot be empty"),
            Self::MetadataTooLarge { max, actual } => {
                write!(f, "metadata has {} entries (max {})", actual, max)
            }
            Self::MetadataValueTooLong { key, max, actual } => {
                write!(f, "metadata key '{}' value is {} bytes (max {})", key, actual, max)
            }
            Self::GraphTooLarge { max_nodes, actual_nodes } => {
                write!(f, "graph has {} nodes (max {})", actual_nodes, max_nodes)
            }
            Self::GraphTooManyEdges { max, actual } => {
                write!(f, "graph has {} edges (max {})", actual, max)
            }
            Self::InputTooLarge { max, actual } => {
                write!(f, "total input size {} bytes (max {})", actual, max)
            }
            Self::InputValueTooLarge { key, max, actual } => {
                write!(f, "input '{}' is {} bytes (max {})", key, actual, max)
            }
            Self::TooManyInputEntries { max, actual } => {
                write!(f, "input has {} entries (max {})", actual, max)
            }
            Self::InvalidPriority { valid_range } => {
                write!(f, "invalid priority (valid: {})", valid_range)
            }
            Self::InvalidMaxConcurrentCells { max, actual } => {
                write!(f, "max_concurrent_cells {} exceeds max {}", actual, max)
            }
            Self::InvalidGracePeriod { max_seconds, actual_seconds } => {
                write!(f, "grace_period {} seconds exceeds max {} seconds", actual_seconds, max_seconds)
            }
            Self::InvalidUuid { value } => {
                write!(f, "invalid UUID: {}", value)
            }
            Self::EmptyGraph => write!(f, "graph must have at least one node"),
        }
    }
}

impl std::error::Error for ValidationError {}

pub struct InputValidator;

impl InputValidator {
    pub fn validate_agent_name(name: &str) -> Result<(), ValidationError> {
        if name.is_empty() {
            return Err(ValidationError::NameEmpty);
        }
        if name.len() > MAX_NAME_LENGTH {
            return Err(ValidationError::NameTooLong {
                max: MAX_NAME_LENGTH,
                actual: name.len(),
            });
        }
        Ok(())
    }

    pub fn validate_metadata(
        metadata: &HashMap<String, String>,
    ) -> Result<(), ValidationError> {
        if metadata.len() > MAX_METADATA_ENTRIES {
            return Err(ValidationError::MetadataTooLarge {
                max: MAX_METADATA_ENTRIES,
                actual: metadata.len(),
            });
        }
        for (key, value) in metadata {
            if value.len() > MAX_METADATA_VALUE_LENGTH {
                return Err(ValidationError::MetadataValueTooLong {
                    key: key.clone(),
                    max: MAX_METADATA_VALUE_LENGTH,
                    actual: value.len(),
                });
            }
        }
        Ok(())
    }

    pub fn validate_priority(priority: u32) -> Result<u8, ValidationError> {
        if priority < 1 || priority > 4 {
            return Err(ValidationError::InvalidPriority {
                valid_range: "1-4".to_string(),
            });
        }
        Ok(priority as u8)
    }

    pub fn validate_max_concurrent_cells(cells: u32) -> Result<usize, ValidationError> {
        if cells as usize > MAX_CONCURRENT_CELLS {
            return Err(ValidationError::InvalidMaxConcurrentCells {
                max: MAX_CONCURRENT_CELLS,
                actual: cells as usize,
            });
        }
        Ok(cells as usize)
    }

    pub fn validate_grace_period(seconds: u64) -> Result<(), ValidationError> {
        if seconds > MAX_GRACE_PERIOD_SECONDS {
            return Err(ValidationError::InvalidGracePeriod {
                max_seconds: MAX_GRACE_PERIOD_SECONDS,
                actual_seconds: seconds,
            });
        }
        Ok(())
    }

    pub fn validate_uuid(value: &str) -> Result<(), ValidationError> {
        uuid::Uuid::parse_str(value)
            .map_err(|_| ValidationError::InvalidUuid { value: value.to_string() })?;
        Ok(())
    }

    pub fn validate_input_size(
        input: &HashMap<String, serde_json::Value>,
    ) -> Result<(), ValidationError> {
        if input.len() > MAX_INPUT_ENTRIES {
            return Err(ValidationError::TooManyInputEntries {
                max: MAX_INPUT_ENTRIES,
                actual: input.len(),
            });
        }

        let mut total_size = 0usize;
        for (key, value) in input {
            let json_size = serde_json::to_string(value)
                .map(|s| s.len())
                .unwrap_or(0);

            total_size += json_size;

            if json_size > MAX_INPUT_VALUE_SIZE {
                return Err(ValidationError::InputValueTooLarge {
                    key: key.clone(),
                    max: MAX_INPUT_VALUE_SIZE,
                    actual: json_size,
                });
            }
        }

        if total_size > MAX_TOTAL_INPUT_SIZE {
            return Err(ValidationError::InputTooLarge {
                max: MAX_TOTAL_INPUT_SIZE,
                actual: total_size,
            });
        }

        Ok(())
    }
}

pub fn sanitize_string(input: &str) -> String {
    input
        .chars()
        .filter(|c| !c.is_ascii_control())
        .take(MAX_NAME_LENGTH)
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_name() {
        assert!(InputValidator::validate_agent_name("valid").is_ok());
        assert!(InputValidator::validate_agent_name("").is_err());
        assert!(InputValidator::validate_agent_name(&"a".repeat(300)).is_err());
    }

    #[test]
    fn test_validate_priority() {
        assert!(InputValidator::validate_priority(1).is_ok());
        assert!(InputValidator::validate_priority(4).is_ok());
        assert!(InputValidator::validate_priority(0).is_err());
        assert!(InputValidator::validate_priority(5).is_err());
    }

    #[test]
    fn test_validate_grace_period() {
        assert!(InputValidator::validate_grace_period(60).is_ok());
        assert!(InputValidator::validate_grace_period(3600).is_ok());
        assert!(InputValidator::validate_grace_period(7200).is_err());
    }
}