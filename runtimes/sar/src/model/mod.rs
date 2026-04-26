use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tracing::info;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum LlmTrafficType {
    Realtime,
    Structured,
    FunctionCalling,
    Background,
    General,
}

impl Default for LlmTrafficType {
    fn default() -> Self {
        LlmTrafficType::General
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelConfig {
    pub name: String,
    pub provider: String,
    pub endpoint: Option<String>,
    pub api_key: Option<String>,
    pub max_tokens: Option<u32>,
    pub temperature: f32,
    pub cost_per_token: CostPerToken,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostPerToken {
    pub input: f64,
    pub output: f64,
}

impl Default for CostPerToken {
    fn default() -> Self {
        Self { input: 0.0, output: 0.0 }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelResponse {
    pub content: String,
    pub model: String,
    pub provider: String,
    pub usage: Usage,
    pub finish_reason: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Usage {
    pub prompt_tokens: u32,
    pub completion_tokens: u32,
    pub total_tokens: u32,
}

impl Usage {
    pub fn total_cost(&self, cost: &CostPerToken) -> f64 {
        (self.prompt_tokens as f64 * cost.input) + (self.completion_tokens as f64 * cost.output)
    }
}

#[derive(Debug, Clone)]
pub struct RoutingContext {
    pub complexity: TaskComplexity,
    pub latency_budget_ms: u64,
    pub cost_budget_usd: f64,
    pub preferred_provider: Option<String>,
}

#[derive(Debug, Clone, Copy)]
pub enum TaskComplexity {
    Low,
    Medium,
    High,
}

impl Default for TaskComplexity {
    fn default() -> Self {
        TaskComplexity::Medium
    }
}

pub struct ModelRouter {
    models: Arc<RwLock<HashMap<String, ModelConfig>>>,
    fallback_chain: Vec<String>,
    config: ModelRouterConfig,
}

impl ModelRouter {
    pub fn new(config: ModelRouterConfig) -> Self {
        let mut models = HashMap::new();
        for model in &config.models {
            models.insert(model.name.clone(), model.clone());
        }
        let fallback_chain = config.fallback_chain.clone();
        Self {
            models: Arc::new(RwLock::new(models)),
            fallback_chain,
            config,
        }
    }

    pub fn add_model(&self, model: ModelConfig) {
        let mut models = self.models.write();
        models.insert(model.name.clone(), model);
    }

    pub fn remove_model(&self, name: &str) {
        let mut models = self.models.write();
        models.remove(name);
    }

    pub async fn route(
        &self,
        prompt: &str,
        traffic_type: LlmTrafficType,
        context: &RoutingContext,
    ) -> anyhow::Result<ModelResponse> {
        let model_name = self.select_model(traffic_type, context)
            .ok_or_else(|| anyhow::anyhow!("No suitable model found"))?;

        let model = self.models.read().get(&model_name).cloned();

        match model {
            Some(m) => {
                info!(model = %m.name, provider = %m.provider, "Routing to model");
                Ok(ModelResponse {
                    content: format!("[Stub response from {}]", m.name),
                    model: m.name,
                    provider: m.provider,
                    usage: Usage {
                        prompt_tokens: (prompt.len() / 4) as u32,
                        completion_tokens: 50,
                        total_tokens: ((prompt.len() / 4) + 50) as u32,
                    },
                    finish_reason: "stop".to_string(),
                })
            }
            None => Err(anyhow::anyhow!("Model {} not found", model_name)),
        }
    }

    fn select_model(&self, _traffic_type: LlmTrafficType, context: &RoutingContext) -> Option<String> {
        let models = self.models.read();
        let suitable: Vec<_> = models.values()
            .filter(|m| context.preferred_provider.as_ref().map_or(true, |p| &m.provider == p))
            .collect();

        if suitable.is_empty() {
            return self.fallback_chain.first().cloned();
        }
        suitable.first().map(|m| m.name.clone())
    }

    pub fn list_models(&self) -> Vec<ModelConfig> {
        self.models.read().values().cloned().collect()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModelRouterConfig {
    pub models: Vec<ModelConfig>,
    pub fallback_chain: Vec<String>,
    pub default_traffic_type: LlmTrafficType,
    pub max_latency_ms: u64,
    pub max_cost_usd: f64,
}

impl Default for ModelRouterConfig {
    fn default() -> Self {
        Self {
            models: vec![
                ModelConfig {
                    name: "gpt-4".to_string(),
                    provider: "openai".to_string(),
                    endpoint: None,
                    api_key: None,
                    max_tokens: Some(4096),
                    temperature: 0.7,
                    cost_per_token: CostPerToken { input: 0.00003, output: 0.00006 },
                },
                ModelConfig {
                    name: "gpt-3.5-turbo".to_string(),
                    provider: "openai".to_string(),
                    endpoint: None,
                    api_key: None,
                    max_tokens: Some(4096),
                    temperature: 0.7,
                    cost_per_token: CostPerToken { input: 0.0000015, output: 0.000002 },
                },
            ],
            fallback_chain: vec!["gpt-3.5-turbo".to_string()],
            default_traffic_type: LlmTrafficType::General,
            max_latency_ms: 5000,
            max_cost_usd: 0.50,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_model_router_route_default() {
        let config = ModelRouterConfig::default();
        let router = ModelRouter::new(config);

        let ctx = RoutingContext {
            complexity: TaskComplexity::Medium,
            latency_budget_ms: 5000,
            cost_budget_usd: 0.50,
            preferred_provider: None,
        };

        let response = router.route("Hello world", LlmTrafficType::General, &ctx).await.unwrap();
        assert_eq!(response.model, "gpt-4");
        assert_eq!(response.provider, "openai");
        assert!(response.usage.prompt_tokens > 0);
    }

    #[tokio::test]
    async fn test_model_router_preferred_provider() {
        let config = ModelRouterConfig::default();
        let router = ModelRouter::new(config);

        let ctx = RoutingContext {
            complexity: TaskComplexity::Low,
            latency_budget_ms: 1000,
            cost_budget_usd: 0.01,
            preferred_provider: Some("openai".to_string()),
        };

        let response = router.route("Quick task", LlmTrafficType::Realtime, &ctx).await.unwrap();
        assert_eq!(response.provider, "openai");
    }

    #[test]
    fn test_usage_cost_calculation() {
        let usage = Usage {
            prompt_tokens: 1000,
            completion_tokens: 500,
            total_tokens: 1500,
        };
        let cost = CostPerToken { input: 0.00003, output: 0.00006 };
        let total = usage.total_cost(&cost);
        assert!((total - 0.06).abs() < 0.001);
    }

    #[test]
    fn test_model_router_add_remove() {
        let config = ModelRouterConfig::default();
        let router = ModelRouter::new(config);

        router.add_model(ModelConfig {
            name: "claude-3".to_string(),
            provider: "anthropic".to_string(),
            endpoint: None,
            api_key: None,
            max_tokens: Some(4096),
            temperature: 0.7,
            cost_per_token: CostPerToken::default(),
        });

        let models = router.list_models();
        assert_eq!(models.len(), 3);

        router.remove_model("claude-3");
        let models = router.list_models();
        assert_eq!(models.len(), 2);
    }
}
