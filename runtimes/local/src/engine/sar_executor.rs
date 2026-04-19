//! Integrated SAR Node Executor — wires real implementations instead of stubs.
//!
//! This executor connects the graph engine to:
//! - **MemoryLayer** (Phase 3) — hot/warm/cold tier memory for Memory nodes
//! - **FlyMindClient** (Phase 4) — routes LLM nodes to Python ai-service
//! - **WasmCellExecutor** (Phase 2) — isolates Tool nodes in WASM cells
//! - **ActionConnectors** (Phase 8) — external service integrations (Stripe, Resend, Shopify, HTTP)
//!
//! ## Execution Flow
//!
//! ```text
//! GraphExecutor (Kahn's DAG traversal)
//!         │
//!         ▼
//! ┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
//! │  LLM Node       │────►│  FlyMindClient   │────►│  Python service │
//! │  (temperature,  │     │  (port 8081)     │     │  (9 providers)  │
//! │   model_hint)   │     └──────────────────┘     └─────────────────┘
//! └─────────────────┘
//!         │
//! ┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
//! │  Memory Node    │────►│  MemoryLayer     │────►│  Hot/Warm/Cold  │
//! │  (read/write)   │     │  (tier cascade)  │     │  tier stores    │
//! └─────────────────┘     └──────────────────┘     └─────────────────┘
//!         │
//! ┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
//! │  Tool Node      │────►│  WasmCellExecutor│────►│  WASM sandbox  │
//! │  (isolated exec)│     │  (pool + linker) │     │  (host fns)     │
//! └─────────────────┘     └──────────────────┘     └─────────────────┘
//!         │
//! ┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
//! │  Action Node    │────►│  ActionConnector │────►│  Stripe/Resend │
//! │  (external svc) │     │  (idempotency)   │     │  Shopify/HTTP  │
//! └─────────────────┘     └──────────────────┘     └─────────────────┘
//! ```
//!
//! ## Error Handling
//!
//! - LLM errors are retryable (provider flakiness)
//! - Memory errors are non-retryable (data consistency)
//! - Tool errors depend on WASM execution (retryable if timeout/OOM)
//! - Action errors are classified by connector (retryable for transient network issues)

use std::collections::HashMap;
use std::sync::Arc;

use tracing::{debug, info, instrument, warn};

use crate::actions::connector::{ActionError, ActionResult, IdempotencyCache, execute_with_idempotency};
use crate::actions::stripe::StripeConnector;
use crate::actions::resend::ResendConnector;
use crate::actions::shopify::ShopifyConnector;
use crate::actions::http::HttpConnector;
use crate::engine::graph::{ExecutionContext, Node, NodeExecutionError, NodeExecutor, NodeType, MemoryOp, LlmTrafficType};
use crate::memory::layer::MemoryLayer;
use crate::router::flymind::FlyMindClient;
use crate::engine::wasm_cell::WasmCellExecutor;

/// Production node executor for the SAR runtime.
///
/// Holds references to all real service implementations and routes
/// node execution to the appropriate backend.
pub struct SarNodeExecutor {
    /// Multi-tier memory for Memory node operations.
    memory_layer: Arc<MemoryLayer>,
    /// FlyMind client for LLM node routing.
    flymind: Arc<FlyMindClient>,
    /// WASM cell executor for Tool node isolation (optional — falls back to direct execution).
    wasm_cell_executor: Option<Arc<WasmCellExecutor>>,
    /// Stripe connector for billing/payment actions (optional).
    stripe_connector: Option<Arc<StripeConnector>>,
    /// Resend connector for email actions (optional).
    resend_connector: Option<Arc<ResendConnector>>,
    /// Shopify connector for e-commerce actions (optional).
    shopify_connector: Option<Arc<ShopifyConnector>>,
    /// Generic HTTP connector for REST API calls (optional).
    http_connector: Option<Arc<HttpConnector>>,
    /// Idempotency cache for action connectors.
    action_idempotency_cache: Arc<IdempotencyCache>,
}

impl SarNodeExecutor {
    /// Create a new executor with all required services.
    pub fn new(
        memory_layer: Arc<MemoryLayer>,
        flymind: Arc<FlyMindClient>,
        wasm_cell_executor: Option<Arc<WasmCellExecutor>>,
    ) -> Self {
        Self {
            memory_layer,
            flymind,
            wasm_cell_executor,
            stripe_connector: None,
            resend_connector: None,
            shopify_connector: None,
            http_connector: None,
            action_idempotency_cache: Arc::new(IdempotencyCache::default()),
        }
    }

    /// Create a new executor with all action connectors enabled.
    pub fn with_all_connectors(
        memory_layer: Arc<MemoryLayer>,
        flymind: Arc<FlyMindClient>,
        wasm_cell_executor: Option<Arc<WasmCellExecutor>>,
        stripe_connector: Option<Arc<StripeConnector>>,
        resend_connector: Option<Arc<ResendConnector>>,
        shopify_connector: Option<Arc<ShopifyConnector>>,
        http_connector: Option<Arc<HttpConnector>>,
    ) -> Self {
        let mut executor = Self::new(memory_layer, flymind, wasm_cell_executor);
        executor.stripe_connector = stripe_connector;
        executor.resend_connector = resend_connector;
        executor.shopify_connector = shopify_connector;
        executor.http_connector = http_connector;
        executor
    }

    /// Create from AppState components (convenience constructor).
    pub fn from_app_state(
        memory_layer: Option<Arc<MemoryLayer>>,
        flymind: Option<Arc<FlyMindClient>>,
        wasm_cell_executor: Option<Arc<WasmCellExecutor>>,
    ) -> anyhow::Result<Self> {
        let memory = memory_layer.ok_or_else(|| anyhow::anyhow!("MemoryLayer not configured"))?;
        let flymind = flymind.ok_or_else(|| anyhow::anyhow!("FlyMindClient not configured"))?;

        // Try to initialize Stripe connector if API key is present
        let stripe_connector = std::env::var("STRIPE_SECRET_KEY").ok().map(|api_key| {
            Arc::new(StripeConnector::new(api_key))
        });

        if stripe_connector.is_some() {
            tracing::info!("Stripe connector initialized");
        }

        // Try to initialize Resend connector if API key and from email are present
        let resend_connector = match (
            std::env::var("RESEND_API_KEY").ok(),
            std::env::var("RESEND_FROM_EMAIL").ok(),
            std::env::var("RESEND_FROM_NAME").ok(),
        ) {
            (Some(api_key), Some(from_email), from_name) => {
                let from_name = from_name.unwrap_or_else(|| "FunctionFly".to_string());
                tracing::info!("Resend connector initialized");
                Some(Arc::new(ResendConnector::new(api_key, from_email, from_name)))
            }
            _ => None,
        };

        // Try to initialize Shopify connector if credentials are present
        let shopify_connector = ShopifyConnector::from_env().map(|conn| {
            tracing::info!("Shopify connector initialized");
            Arc::new(conn)
        });

        // HTTP connector is always available (no credentials needed for basic usage)
        let http_connector = Some(Arc::new(HttpConnector::from_env()));
        tracing::info!("HTTP connector initialized");

        Ok(Self::with_all_connectors(
            memory,
            flymind,
            wasm_cell_executor,
            stripe_connector,
            resend_connector,
            shopify_connector,
            http_connector,
        ))
    }
}

impl NodeExecutor for SarNodeExecutor {
    #[instrument(skip(self, input, ctx), fields(node_id = %node.id, node_name = %node.name, node_type = ?node.node_type))]
    async fn execute_node(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        match &node.node_type {
            NodeType::LLM { model, prompt, temperature, max_tokens, traffic_type } => {
                execute_llm_node(
                    &self.flymind,
                    node,
                    input,
                    prompt.clone(),
                    *temperature,
                    *max_tokens,
                    *traffic_type,
                ).await
            }
            NodeType::Memory { operation, key } => {
                execute_memory_node(
                    &self.memory_layer,
                    node,
                    *operation,
                    key.clone(),
                    input,
                    ctx,
                ).await
            }
            NodeType::Tool { name, params } => {
                execute_tool_node(
                    self.wasm_cell_executor.as_ref(),
                    node,
                    name.clone(),
                    params.clone(),
                    input,
                ).await
            }
            NodeType::Control { kind, condition } => {
                execute_control_node(node, *kind, condition.clone(), input).await
            }
            NodeType::Optimization { strategy } => {
                execute_optimization_node(node, *strategy, input).await
            }
            NodeType::Action { connector, action, params } => {
                execute_action_node(
                    self.stripe_connector.as_ref(),
                    self.resend_connector.as_ref(),
                    self.shopify_connector.as_ref(),
                    self.http_connector.as_ref(),
                    &self.action_idempotency_cache,
                    node,
                    connector.clone(),
                    action.clone(),
                    params.clone(),
                    input,
                    ctx.tenant_id.as_deref(),
                ).await
            }
            NodeType::Passthrough => {
                // Passthrough just returns the input as output
                Ok(serde_json::Value::Object(input.into_iter().collect()))
            }
        }
    }
}

// Implement NodeExecutor for &SarNodeExecutor so GraphExecutor can use Arc<SarNodeExecutor>
impl NodeExecutor for &SarNodeExecutor {
    async fn execute_node(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        // Delegate to the owned implementation
        (**self).execute_node(node, input, ctx).await
    }
}

// -----------------------------------------------------------------------------
// LLM Node Execution (Phase 4: FlyMind Router)
// -----------------------------------------------------------------------------

#[instrument(skip(flymind, input), fields(prompt_len = prompt.len()))]
async fn execute_llm_node(
    flymind: &Arc<FlyMindClient>,
    node: &Node,
    input: HashMap<String, serde_json::Value>,
    prompt: String,
    temperature: f32,
    max_tokens: Option<u32>,
    traffic_type: LlmTrafficType,
) -> Result<serde_json::Value, NodeExecutionError> {
    debug!(traffic_type = ?traffic_type, "Routing LLM node to FlyMind");

    // Build messages from input — look for 'system' and 'user' keys,
    // or construct a single user message from the 'input' field.
    let mut messages: HashMap<String, String> = HashMap::new();

    // If input has explicit role keys, use them
    if let Some(system) = input.get("system").and_then(|v| v.as_str()) {
        messages.insert("system".to_string(), system.to_string());
    }
    if let Some(user) = input.get("user").and_then(|v| v.as_str()) {
        messages.insert("user".to_string(), format!("{}\n\n{}", prompt, user));
    } else if let Some(input_val) = input.get("input").and_then(|v| v.as_str()) {
        messages.insert("user".to_string(), format!("{}\n\n{}", prompt, input_val));
    } else {
        // Fallback: just send the prompt as user message
        messages.insert("user".to_string(), prompt);
    }

    // Call FlyMind
    let result = flymind.complete(
        &messages,
        traffic_type,
        None, // model hint — let FlyMind decide based on traffic_type
        temperature,
        max_tokens,
    ).await;

    match result {
        Ok(route_result) => {
            info!(
                provider = %route_result.provider,
                model = %route_result.model,
                latency_ms = %route_result.latency_ms,
                tokens = route_result.usage.total_tokens,
                "LLM completion successful"
            );

            let output = serde_json::json!({
                "content": route_result.content,
                "provider": route_result.provider,
                "model": route_result.model,
                "usage": {
                    "prompt_tokens": route_result.usage.prompt_tokens,
                    "completion_tokens": route_result.usage.completion_tokens,
                    "total_tokens": route_result.usage.total_tokens,
                },
                "latency_ms": route_result.latency_ms,
            });

            Ok(output)
        }
        Err(e) => {
            warn!(error = %e, "LLM completion failed");
            // LLM errors are retryable (provider flakiness)
            Err(NodeExecutionError::new(node.id, format!("LLM call failed: {}", e)))
        }
    }
}

// -----------------------------------------------------------------------------
// Memory Node Execution (Phase 3: Memory Layer)
// -----------------------------------------------------------------------------

#[instrument(skip(memory_layer), fields(operation = ?operation, key = %key))]
async fn execute_memory_node(
    memory_layer: &Arc<MemoryLayer>,
    node: &Node,
    operation: MemoryOp,
    key: String,
    input: HashMap<String, serde_json::Value>,
    ctx: &ExecutionContext,
) -> Result<serde_json::Value, NodeExecutionError> {
    let tenant_id = ctx.tenant_id.as_deref();

    match operation {
        MemoryOp::Read => {
            match memory_layer.read(tenant_id, &key).await {
                Ok(Some(value)) => {
                    debug!(tier = "hit", "Memory read successful");
                    Ok(serde_json::json!({
                        "key": key,
                        "value": value,
                        "found": true,
                    }))
                }
                Ok(None) => {
                    debug!(tier = "miss", "Memory key not found");
                    Ok(serde_json::json!({
                        "key": key,
                        "value": null,
                        "found": false,
                    }))
                }
                Err(e) => {
                    warn!(error = %e, "Memory read error");
                    // Memory errors are non-retryable (data consistency issues)
                    Err(NodeExecutionError::non_retryable(node.id, format!("Memory read failed: {}", e)))
                }
            }
        }
        MemoryOp::Write => {
            // Extract value from input — look for 'value' field
            let value = input.get("value")
                .map(|v| v.to_string())
                .unwrap_or_default();

            match memory_layer.write(tenant_id, &key, value).await {
                Ok(()) => {
                    debug!("Memory write successful");
                    Ok(serde_json::json!({
                        "key": key,
                        "written": true,
                    }))
                }
                Err(e) => {
                    warn!(error = %e, "Memory write error");
                    Err(NodeExecutionError::non_retryable(node.id, format!("Memory write failed: {}", e)))
                }
            }
        }
        MemoryOp::Delete => {
            match memory_layer.delete(tenant_id, &key).await {
                Ok(deleted) => {
                    debug!(deleted = deleted, "Memory delete completed");
                    Ok(serde_json::json!({
                        "key": key,
                        "deleted": deleted,
                    }))
                }
                Err(e) => {
                    warn!(error = %e, "Memory delete error");
                    Err(NodeExecutionError::non_retryable(node.id, format!("Memory delete failed: {}", e)))
                }
            }
        }
        MemoryOp::List => {
            // List is not fully implemented — return empty for now
            debug!("Memory list not yet implemented");
            Ok(serde_json::json!({
                "key": key,
                "entries": [],
                "note": "List operation not yet implemented",
            }))
        }
    }
}

// -----------------------------------------------------------------------------
// Tool Node Execution (Phase 2: WASM Cell Isolation)
// -----------------------------------------------------------------------------

#[instrument(skip(wasm_cell_executor, input), fields(tool_name = %name))]
async fn execute_tool_node(
    wasm_cell_executor: Option<&Arc<WasmCellExecutor>>,
    node: &Node,
    name: String,
    params: serde_json::Value,
    input: HashMap<String, serde_json::Value>,
) -> Result<serde_json::Value, NodeExecutionError> {
    // Merge input and params into a single JSON payload
    let mut tool_input = serde_json::Map::new();
    tool_input.insert("tool_name".to_string(), serde_json::Value::String(name.clone()));
    tool_input.insert("params".to_string(), params.clone());
    tool_input.insert("input".to_string(), serde_json::Value::Object(input.into_iter().collect()));

    let tool_input_json = serde_json::to_string(&tool_input)
        .map_err(|e| NodeExecutionError::new(node.id, format!("Failed to serialize tool input: {}", e)))?;

    // If WASM cell executor is available, use it for isolation
    if let Some(executor) = wasm_cell_executor {
        // Look up function by name — assumes the tool WASM was registered
        let function_key = format!("tool:{}", name);

        match executor.execute_cell(&function_key, &tool_input_json).await {
            Ok(output) => {
                // Parse the output as JSON
                match serde_json::from_str::<serde_json::Value>(&output) {
                    Ok(json) => Ok(json),
                    Err(_) => {
                        // If not valid JSON, wrap as string result
                        Ok(serde_json::json!({
                            "tool": name,
                            "result": output,
                        }))
                    }
                }
            }
            Err(e) => {
                warn!(error = %e, "Tool WASM execution failed");
                // Tool errors may be retryable (OOM, timeout)
                Err(NodeExecutionError::new(node.id, format!("Tool execution failed: {}", e)))
            }
        }
    } else {
        // Fallback: return stub response if no WASM executor
        warn!("No WasmCellExecutor configured — returning stub tool result");
        Ok(serde_json::json!({
            "tool": name,
            "result": format!("[Tool stub - no WASM executor] {} called", name),
            "params": params,
        }))
    }
}

// -----------------------------------------------------------------------------
// Control Node Execution (Conditional Branching)
// -----------------------------------------------------------------------------

#[instrument(skip(input), fields(kind = ?kind))]
async fn execute_control_node(
    node: &Node,
    kind: crate::engine::graph::ControlKind,
    condition: crate::engine::graph::Expr,
    input: HashMap<String, serde_json::Value>,
) -> Result<serde_json::Value, NodeExecutionError> {
    let condition_met = condition.eval(&input);

    let output = match kind {
        crate::engine::graph::ControlKind::If => {
            serde_json::json!({
                "control": "if",
                "condition_met": condition_met,
                "branch": if condition_met { "then" } else { "else" },
            })
        }
        crate::engine::graph::ControlKind::Loop => {
            serde_json::json!({
                "control": "loop",
                "condition_met": condition_met,
                "continue": condition_met,
            })
        }
        crate::engine::graph::ControlKind::Switch => {
            serde_json::json!({
                "control": "switch",
                "condition_met": condition_met,
                "case": condition_met.to_string(),
            })
        }
    };

    Ok(output)
}

// -----------------------------------------------------------------------------
// Optimization Node Execution (Phase 7: Self-Optimization)
// -----------------------------------------------------------------------------

#[instrument(fields(strategy = ?strategy))]
async fn execute_optimization_node(
    node: &Node,
    strategy: crate::engine::graph::OptStrategy,
    _input: HashMap<String, serde_json::Value>,
) -> Result<serde_json::Value, NodeExecutionError> {
    // Optimization nodes execute strategy suggestions.
    // The actual graph mutation happens through the GraphMutator in the optimizer module.
    // This node records the optimization intent which can be applied via API call.
    info!(strategy = ?strategy, "Optimization node executed");

    let output = serde_json::json!({
        "optimization": format!("{:?}", strategy),
        "node_id": node.id.to_string(),
        "node_name": node.name,
        "suggestion": match strategy {
            crate::engine::graph::OptStrategy::AdjustTimeouts => {
                "Detected high timeout rate — suggest increasing node timeout"
            }
            crate::engine::graph::OptStrategy::EnableCaching => {
                "Detected stable high success rate — suggest enabling result caching"
            }
            crate::engine::graph::OptStrategy::IncreaseQuota => {
                "Suggest increasing tenant quota based on usage patterns"
            }
            crate::engine::graph::OptStrategy::SimplifyPath => {
                "Detected redundant nodes — suggest path simplification"
            }
        },
        "applied": false,
        "can_apply_via_api": true,
        "api_endpoint": "/api/graphs/{graph_id}/optimize",
        "note": "Optimization suggestion generated. Apply via GraphMutator API.",
    });

    Ok(output)
}

// -----------------------------------------------------------------------------
// Action Node Execution (Phase 8: External Service Connectors)
// -----------------------------------------------------------------------------

#[instrument(skip(stripe, resend, shopify, http, cache, input), fields(connector = %connector_name, action = %action_name))]
async fn execute_action_node(
    stripe: Option<&Arc<StripeConnector>>,
    resend: Option<&Arc<ResendConnector>>,
    shopify: Option<&Arc<ShopifyConnector>>,
    http: Option<&Arc<HttpConnector>>,
    cache: &Arc<IdempotencyCache>,
    node: &Node,
    connector_name: String,
    action_name: String,
    params: serde_json::Value,
    input: HashMap<String, serde_json::Value>,
    tenant_id: Option<&str>,
) -> Result<serde_json::Value, NodeExecutionError> {
    // Merge input into params (input values take precedence)
    let mut merged_params = params.clone();
    if let serde_json::Value::Object(ref mut p) = merged_params {
        for (k, v) in input {
            p.insert(k, v);
        }
    }

    // Helper to build success response
    fn build_action_response(
        connector: &str,
        action: &str,
        result: &ActionResult,
    ) -> serde_json::Value {
        let mut output = serde_json::json!({
            "connector": connector,
            "action": action,
            "success": result.success,
            "data": result.data,
            "latency_ms": result.latency_ms,
        });
        
        if let Some(provider_ref) = &result.provider_ref {
            output["provider_ref"] = serde_json::Value::String(provider_ref.clone());
        }
        if let Some(error) = &result.error {
            output["error"] = serde_json::Value::String(error.clone());
        }
        
        output
    }

    // Helper to handle action errors
    fn handle_action_error(
        node: &Node,
        action: &str,
        connector: &str,
        err: &ActionError,
    ) -> NodeExecutionError {
        if err.retryable {
            NodeExecutionError::new(
                node.id,
                format!("{} {} failed: {}", connector, action, err),
            )
        } else {
            NodeExecutionError::non_retryable(
                node.id,
                format!("{} {} failed: {}", connector, action, err),
            )
        }
    }

    // Route to the appropriate connector
    match connector_name.as_str() {
        "stripe" => {
            let Some(connector) = stripe else {
                return Err(NodeExecutionError::non_retryable(
                    node.id,
                    "Stripe connector not configured. Set STRIPE_SECRET_KEY".to_string(),
                ));
            };

            match execute_with_idempotency(
                connector.as_ref(),
                cache,
                tenant_id,
                &action_name,
                merged_params,
                3, // max retries
            ).await {
                Ok(result) => {
                    info!(
                        action = %action_name,
                        success = result.success,
                        provider_ref = ?result.provider_ref,
                        "Stripe action completed successfully"
                    );
                    Ok(build_action_response("stripe", &action_name, &result))
                }
                Err(err) => {
                    warn!(error = %err, "Stripe action failed");
                    Err(handle_action_error(node, &action_name, "Stripe", &err))
                }
            }
        }

        "resend" => {
            let Some(connector) = resend else {
                return Err(NodeExecutionError::non_retryable(
                    node.id,
                    "Resend connector not configured. Set RESEND_API_KEY and RESEND_FROM_EMAIL".to_string(),
                ));
            };

            match execute_with_idempotency(
                connector.as_ref(),
                cache,
                tenant_id,
                &action_name,
                merged_params,
                3, // max retries
            ).await {
                Ok(result) => {
                    info!(
                        action = %action_name,
                        success = result.success,
                        provider_ref = ?result.provider_ref,
                        "Resend action completed successfully"
                    );
                    Ok(build_action_response("resend", &action_name, &result))
                }
                Err(err) => {
                    warn!(error = %err, "Resend action failed");
                    Err(handle_action_error(node, &action_name, "Resend", &err))
                }
            }
        }

        "shopify" => {
            let Some(connector) = shopify else {
                return Err(NodeExecutionError::non_retryable(
                    node.id,
                    "Shopify connector not configured. Set SHOPIFY_SHOP_DOMAIN and SHOPIFY_ACCESS_TOKEN".to_string(),
                ));
            };

            match execute_with_idempotency(
                connector.as_ref(),
                cache,
                tenant_id,
                &action_name,
                merged_params,
                3, // max retries
            ).await {
                Ok(result) => {
                    info!(
                        action = %action_name,
                        success = result.success,
                        provider_ref = ?result.provider_ref,
                        "Shopify action completed successfully"
                    );
                    Ok(build_action_response("shopify", &action_name, &result))
                }
                Err(err) => {
                    warn!(error = %err, "Shopify action failed");
                    Err(handle_action_error(node, &action_name, "Shopify", &err))
                }
            }
        }

        "http" => {
            let Some(connector) = http else {
                return Err(NodeExecutionError::non_retryable(
                    node.id,
                    "HTTP connector not available".to_string(),
                ));
            };

            match execute_with_idempotency(
                connector.as_ref(),
                cache,
                tenant_id,
                &action_name,
                merged_params,
                3, // max retries
            ).await {
                Ok(result) => {
                    info!(
                        action = %action_name,
                        success = result.success,
                        "HTTP action completed successfully"
                    );
                    Ok(build_action_response("http", &action_name, &result))
                }
                Err(err) => {
                    warn!(error = %err, "HTTP action failed");
                    Err(handle_action_error(node, &action_name, "HTTP", &err))
                }
            }
        }

        _ => {
            Err(NodeExecutionError::non_retryable(
                node.id,
                format!("Unknown action connector: {}. Supported: stripe, resend, shopify, http", connector_name),
            ))
        }
    }
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_control_node_if_true() {
        let node = Node {
            id: crate::engine::graph::NodeId(uuid::Uuid::new_v4()),
            name: "test".to_string(),
            node_type: NodeType::Passthrough,
            timeout_ms: 1000,
            retry: Default::default(),
            input_schema: None,
            output_schema: None,
            metadata: HashMap::new(),
        };

        let mut input = HashMap::new();
        input.insert("condition".to_string(), serde_json::json!(true));

        let condition = crate::engine::graph::Expr::Var("condition".to_string());
        let result = execute_control_node(&node, crate::engine::graph::ControlKind::If, condition, input).await;

        assert!(result.is_ok());
        let json = result.unwrap();
        assert_eq!(json["control"], "if");
        assert_eq!(json["condition_met"], true);
        assert_eq!(json["branch"], "then");
    }

    #[tokio::test]
    async fn test_control_node_if_false() {
        let node = Node {
            id: crate::engine::graph::NodeId(uuid::Uuid::new_v4()),
            name: "test".to_string(),
            node_type: NodeType::Passthrough,
            timeout_ms: 1000,
            retry: Default::default(),
            input_schema: None,
            output_schema: None,
            metadata: HashMap::new(),
        };

        let mut input = HashMap::new();
        input.insert("condition".to_string(), serde_json::json!(false));

        let condition = crate::engine::graph::Expr::Var("condition".to_string());
        let result = execute_control_node(&node, crate::engine::graph::ControlKind::If, condition, input).await;

        assert!(result.is_ok());
        let json = result.unwrap();
        assert_eq!(json["branch"], "else");
    }
}
