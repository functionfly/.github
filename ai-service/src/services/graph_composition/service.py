"""AI Graph Composition Service for FunctionFly.

This service generates graph topologies (nodes and edges) from natural language prompts.
It uses LLM to:
1. Parse intent (SaaS, marketplace, API backend, etc.)
2. Retrieve relevant function templates from RAG
3. Generate node topology (which functions, in what order)
4. Suggest connections based on input/output schema matching

Part of the "Backend as a Graph" vision - Phase 1 implementation.
"""

import json
import logging
import time
from typing import Optional, List, Dict, Any, Tuple
from dataclasses import dataclass

from ...models.schemas import (
    GraphCompositionRequest,
    GraphCompositionResponse,
    GraphCompositionExplanation,
    GraphDefinition,
    GraphNodeRef,
    GraphNodeInput,
    GraphNodeOutput,
    GraphEdge,
    GraphEdgeMapping,
    GraphTriggerConfig,
    TemplateCategory,
    GraphTemplateInfo,
    ChatMessage,
    MessageRole,
    FunctionGenerationRequest,
)
from ...providers.manager import get_provider_manager
from ...security.auth import APIKeyInfo

logger = logging.getLogger(__name__)


@dataclass
class CompositionAttempt:
    """Record of a composition attempt."""
    model: str
    provider: str
    success: bool
    cost_usd: float
    tokens_in: int
    tokens_out: int
    latency_ms: float
    errors: List[str]


class GraphCompositionService:
    """AI-powered graph composition for Backend as a Graph.

    Generates complete graph definitions (nodes + edges) from natural language.
    Supports templates for common patterns (SaaS, marketplace, API backend).
    """

    # Prebuilt template catalog - minimal viable templates for Phase 1
    TEMPLATES = {
        "saas_starter": {
            "id": "saas_starter",
            "name": "SaaS Starter Kit",
            "description": "Complete SaaS backend: user signup, Stripe billing, welcome email, dashboard API",
            "category": TemplateCategory.SAAS_STARTER,
            "tags": ["auth", "payments", "email", "saas"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 5,
            "popular_use_cases": [
                "MVP SaaS applications",
                "Subscription-based services",
                "User onboarding flows"
            ],
            "default_nodes": [
                {"node_id": "auth-signup", "author": "functionfly", "name": "auth-signup", "description": "Handle user registration with validation"},
                {"node_id": "stripe-create-customer", "author": "functionfly", "name": "stripe-create-customer", "description": "Create Stripe customer record"},
                {"node_id": "send-welcome-email", "author": "functionfly", "name": "send-welcome-email", "description": "Send welcome email via Resend"},
                {"node_id": "create-user-record", "author": "functionfly", "name": "create-user-record", "description": "Create user in database"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "auth-signup", "target_node_id": "stripe-create-customer"},
                {"id": "e2", "source_node_id": "stripe-create-customer", "target_node_id": "create-user-record"},
                {"id": "e3", "source_node_id": "create-user-record", "target_node_id": "send-welcome-email"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/signup", "method": "POST"}},
        },
        "ecommerce_checkout": {
            "id": "ecommerce_checkout",
            "name": "E-commerce Checkout",
            "description": "Complete checkout flow: validate cart, process payment, create order, send receipt",
            "category": TemplateCategory.MARKETPLACE,
            "tags": ["payments", "orders", "receipts", "ecommerce"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 10,
            "popular_use_cases": [
                "Online stores",
                "Digital product sales",
                "Service bookings"
            ],
            "default_nodes": [
                {"node_id": "validate-cart", "author": "functionfly", "name": "validate-cart", "description": "Validate cart items and inventory"},
                {"node_id": "calculate-totals", "author": "functionfly", "name": "calculate-totals", "description": "Calculate subtotal, tax, shipping"},
                {"node_id": "process-payment", "author": "functionfly", "name": "process-stripe-payment", "description": "Process Stripe payment"},
                {"node_id": "create-order", "author": "functionfly", "name": "create-order", "description": "Create order record in database"},
                {"node_id": "send-receipt", "author": "functionfly", "name": "send-receipt-email", "description": "Send order receipt email"},
                {"node_id": "update-inventory", "author": "functionfly", "name": "update-inventory", "description": "Decrement inventory counts"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-cart", "target_node_id": "calculate-totals"},
                {"id": "e2", "source_node_id": "calculate-totals", "target_node_id": "process-payment"},
                {"id": "e3", "source_node_id": "process-payment", "target_node_id": "create-order"},
                {"id": "e4", "source_node_id": "process-payment", "target_node_id": "update-inventory"},
                {"id": "e5", "source_node_id": "create-order", "target_node_id": "send-receipt"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/checkout", "method": "POST"}},
        },
        "api_backend": {
            "id": "api_backend",
            "name": "CRUD API Backend",
            "description": "RESTful API backend with CRUD operations, auth middleware, and caching",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["api", "crud", "auth", "caching"],
            "complexity": "simple",
            "estimated_setup_time_minutes": 3,
            "popular_use_cases": [
                "Mobile app backends",
                "Internal tools",
                "Content APIs"
            ],
            "default_nodes": [
                {"node_id": "auth-middleware", "author": "functionfly", "name": "auth-middleware", "description": "Validate JWT token"},
                {"node_id": "cache-check", "author": "functionfly", "name": "cache-get", "description": "Check Redis cache"},
                {"node_id": "db-query", "author": "functionfly", "name": "db-query", "description": "Query database"},
                {"node_id": "cache-set", "author": "functionfly", "name": "cache-set", "description": "Update cache"},
                {"node_id": "format-response", "author": "functionfly", "name": "format-response", "description": "Format JSON response"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "auth-middleware", "target_node_id": "cache-check"},
                {"id": "e2", "source_node_id": "cache-check", "target_node_id": "db-query", "condition": {"operator": "eq", "field": "$.cached", "value": False}},
                {"id": "e3", "source_node_id": "db-query", "target_node_id": "cache-set"},
                {"id": "e4", "source_node_id": "cache-check", "target_node_id": "format-response", "condition": {"operator": "eq", "field": "$.cached", "value": True}},
                {"id": "e5", "source_node_id": "cache-set", "target_node_id": "format-response"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/resource", "method": "GET"}},
        },
        "webhook_processor": {
            "id": "webhook_processor",
            "name": "Webhook Event Processor",
            "description": "Process incoming webhooks with validation, queuing, and retry logic",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["webhooks", "events", "queue", "retry"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 7,
            "popular_use_cases": [
                "Stripe webhook handling",
                "GitHub event processing",
                "Third-party integrations"
            ],
            "default_nodes": [
                {"node_id": "validate-signature", "author": "functionfly", "name": "validate-webhook-signature", "description": "Validate webhook HMAC signature"},
                {"node_id": "parse-event", "author": "functionfly", "name": "parse-webhook-event", "description": "Parse and validate event payload"},
                {"node_id": "queue-event", "author": "functionfly", "name": "queue-to-redis", "description": "Queue event for processing"},
                {"node_id": "process-event", "author": "functionfly", "name": "process-event", "description": "Process event business logic"},
                {"node_id": "notify-completion", "author": "functionfly", "name": "send-notification", "description": "Send completion notification"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-signature", "target_node_id": "parse-event"},
                {"id": "e2", "source_node_id": "parse-event", "target_node_id": "queue-event"},
                {"id": "e3", "source_node_id": "queue-event", "target_node_id": "process-event"},
                {"id": "e4", "source_node_id": "process-event", "target_node_id": "notify-completion"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/webhook", "method": "POST"}},
        },
        # ============================================================================
        # AUTHENTICATION & IDENTITY TEMPLATES
        # ============================================================================
        "passwordless_auth": {
            "id": "passwordless_auth",
            "name": "Passwordless Authentication",
            "description": "Magic link and OTP-based authentication without passwords",
            "category": TemplateCategory.SAAS_STARTER,
            "tags": ["auth", "passwordless", "magic-link", "otp", "security"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 8,
            "popular_use_cases": [
                "Passwordless login",
                "Email OTP verification",
                "SMS 2FA",
                "Secure user onboarding"
            ],
            "default_nodes": [
                {"node_id": "validate-email", "author": "functionfly", "name": "validate-email", "description": "Validate email format and domain"},
                {"node_id": "check-existing-user", "author": "functionfly", "name": "db-query", "description": "Check if user already exists"},
                {"node_id": "generate-otp", "author": "functionfly", "name": "generate-otp", "description": "Generate secure OTP code"},
                {"node_id": "store-otp", "author": "functionfly", "name": "cache-set", "description": "Store OTP in Redis with TTL"},
                {"node_id": "send-otp-email", "author": "functionfly", "name": "send-otp-email", "description": "Send OTP via email"},
                {"node_id": "verify-otp", "author": "functionfly", "name": "verify-otp", "description": "Verify OTP code"},
                {"node_id": "generate-jwt", "author": "functionfly", "name": "generate-jwt", "description": "Generate JWT token"},
                {"node_id": "create-session", "author": "functionfly", "name": "create-session", "description": "Create user session"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-email", "target_node_id": "check-existing-user"},
                {"id": "e2", "source_node_id": "check-existing-user", "target_node_id": "generate-otp"},
                {"id": "e3", "source_node_id": "generate-otp", "target_node_id": "store-otp"},
                {"id": "e4", "source_node_id": "store-otp", "target_node_id": "send-otp-email"},
                {"id": "e5", "source_node_id": "verify-otp", "target_node_id": "generate-jwt"},
                {"id": "e6", "source_node_id": "generate-jwt", "target_node_id": "create-session"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/auth/passwordless", "method": "POST"}},
        },
        "oauth_integration": {
            "id": "oauth_integration",
            "name": "OAuth 2.0 Integration",
            "description": "Social login with Google, GitHub, Apple, and other OAuth providers",
            "category": TemplateCategory.SAAS_STARTER,
            "tags": ["auth", "oauth", "social-login", "sso", "google", "github"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 10,
            "popular_use_cases": [
                "Google Sign-In",
                "GitHub OAuth",
                "Apple Sign-In",
                "Microsoft/Azure AD"
            ],
            "default_nodes": [
                {"node_id": "validate-oauth-code", "author": "functionfly", "name": "validate-oauth-code", "description": "Validate OAuth authorization code"},
                {"node_id": "exchange-token", "author": "functionfly", "name": "exchange-oauth-token", "description": "Exchange code for access token"},
                {"node_id": "fetch-user-profile", "author": "functionfly", "name": "fetch-oauth-profile", "description": "Fetch user profile from provider"},
                {"node_id": "check-existing-oauth", "author": "functionfly", "name": "db-query", "description": "Check for existing OAuth user"},
                {"node_id": "link-or-create-user", "author": "functionfly", "name": "upsert-user", "description": "Link to existing or create new user"},
                {"node_id": "generate-session", "author": "functionfly", "name": "generate-jwt", "description": "Generate session token"},
                {"node_id": "log-auth-event", "author": "functionfly", "name": "log-auth", "description": "Log authentication event"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-oauth-code", "target_node_id": "exchange-token"},
                {"id": "e2", "source_node_id": "exchange-token", "target_node_id": "fetch-user-profile"},
                {"id": "e3", "source_node_id": "fetch-user-profile", "target_node_id": "check-existing-oauth"},
                {"id": "e4", "source_node_id": "check-existing-oauth", "target_node_id": "link-or-create-user"},
                {"id": "e5", "source_node_id": "link-or-create-user", "target_node_id": "generate-session"},
                {"id": "e6", "source_node_id": "generate-session", "target_node_id": "log-auth-event"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/auth/oauth/callback", "method": "POST"}},
        },
        "rbac_authorization": {
            "id": "rbac_authorization",
            "name": "RBAC Authorization System",
            "description": "Role-based access control with permissions, roles, and user assignments",
            "category": TemplateCategory.SAAS_STARTER,
            "tags": ["auth", "rbac", "permissions", "roles", "authorization"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 15,
            "popular_use_cases": [
                "Admin dashboards",
                "Multi-tenant apps",
                "Team-based permissions",
                "Feature flags"
            ],
            "default_nodes": [
                {"node_id": "validate-jwt", "author": "functionfly", "name": "auth-middleware", "description": "Validate JWT token"},
                {"node_id": "extract-user-roles", "author": "functionfly", "name": "extract-roles", "description": "Extract roles from token or DB"},
                {"node_id": "load-permissions", "author": "functionfly", "name": "load-permissions", "description": "Load role permissions from DB"},
                {"node_id": "check-permission", "author": "functionfly", "name": "check-permission", "description": "Check if user has required permission"},
                {"node_id": "permission-cache", "author": "functionfly", "name": "cache-set", "description": "Cache permission check result"},
                {"node_id": "audit-log", "author": "functionfly", "name": "log-auth", "description": "Log authorization check"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-jwt", "target_node_id": "extract-user-roles"},
                {"id": "e2", "source_node_id": "extract-user-roles", "target_node_id": "load-permissions"},
                {"id": "e3", "source_node_id": "load-permissions", "target_node_id": "check-permission"},
                {"id": "e4", "source_node_id": "check-permission", "target_node_id": "permission-cache"},
                {"id": "e5", "source_node_id": "check-permission", "target_node_id": "audit-log"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/authz/check", "method": "POST"}},
        },
        # ============================================================================
        # AI & ML PIPELINES
        # ============================================================================
        "rag_pipeline": {
            "id": "rag_pipeline",
            "name": "RAG (Retrieval-Augmented Generation)",
            "description": "AI pipeline: embed query, retrieve documents from vector DB, generate response",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["ai", "rag", "embeddings", "openai", "vector-db", "llm"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 12,
            "popular_use_cases": [
                "AI customer support",
                "Knowledge base Q&A",
                "Document search",
                "Chatbots"
            ],
            "default_nodes": [
                {"node_id": "sanitize-query", "author": "functionfly", "name": "sanitize-input", "description": "Sanitize and validate user query"},
                {"node_id": "embed-query", "author": "functionfly", "name": "openai-embed", "description": "Generate embeddings via OpenAI"},
                {"node_id": "vector-search", "author": "functionfly", "name": "pgvector-search", "description": "Search similar documents in vector DB"},
                {"node_id": "rerank-results", "author": "functionfly", "name": "rerank-documents", "description": "Rerank retrieved documents"},
                {"node_id": "build-context", "author": "functionfly", "name": "build-context", "description": "Build context from documents"},
                {"node_id": "generate-response", "author": "functionfly", "name": "openai-completion", "description": "Generate AI response"},
                {"node_id": "log-usage", "author": "functionfly", "name": "log-ai-usage", "description": "Log AI API usage"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "sanitize-query", "target_node_id": "embed-query"},
                {"id": "e2", "source_node_id": "embed-query", "target_node_id": "vector-search"},
                {"id": "e3", "source_node_id": "vector-search", "target_node_id": "rerank-results"},
                {"id": "e4", "source_node_id": "rerank-results", "target_node_id": "build-context"},
                {"id": "e5", "source_node_id": "build-context", "target_node_id": "generate-response"},
                {"id": "e6", "source_node_id": "generate-response", "target_node_id": "log-usage"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/ai/query", "method": "POST"}},
        },
        "ai_image_generation": {
            "id": "ai_image_generation",
            "name": "AI Image Generation Pipeline",
            "description": "Generate images from prompts with caching and rate limiting",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["ai", "image-generation", "dall-e", "stable-diffusion", "caching"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 8,
            "popular_use_cases": [
                "Product image generation",
                "Marketing assets",
                "Social media content",
                "Thumbnail creation"
            ],
            "default_nodes": [
                {"node_id": "validate-prompt", "author": "functionfly", "name": "moderate-content", "description": "Validate and moderate prompt"},
                {"node_id": "check-cache", "author": "functionfly", "name": "cache-get", "description": "Check for cached image"},
                {"node_id": "rate-limit", "author": "functionfly", "name": "rate-limit", "description": "Check rate limits"},
                {"node_id": "generate-image", "author": "functionfly", "name": "dalle-generate", "description": "Generate image via DALL-E"},
                {"node_id": "upload-storage", "author": "functionfly", "name": "upload-s3", "description": "Upload to cloud storage"},
                {"node_id": "cache-result", "author": "functionfly", "name": "cache-set", "description": "Cache image URL"},
                {"node_id": "log-generation", "author": "functionfly", "name": "log-ai-usage", "description": "Log generation metrics"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-prompt", "target_node_id": "check-cache"},
                {"id": "e2", "source_node_id": "check-cache", "target_node_id": "rate-limit", "condition": {"operator": "eq", "field": "$.cached", "value": False}},
                {"id": "e3", "source_node_id": "check-cache", "target_node_id": "log-generation", "condition": {"operator": "eq", "field": "$.cached", "value": True}},
                {"id": "e4", "source_node_id": "rate-limit", "target_node_id": "generate-image"},
                {"id": "e5", "source_node_id": "generate-image", "target_node_id": "upload-storage"},
                {"id": "e6", "source_node_id": "upload-storage", "target_node_id": "cache-result"},
                {"id": "e7", "source_node_id": "cache-result", "target_node_id": "log-generation"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/ai/image", "method": "POST"}},
        },
        "content_moderation": {
            "id": "content_moderation",
            "name": "AI Content Moderation",
            "description": "Moderate user-generated content with toxicity, spam, and policy checks",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["ai", "moderation", "toxicity", "spam", "content-safety"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 6,
            "popular_use_cases": [
                "Comment moderation",
                "UGC filtering",
                "Community management",
                "Brand safety"
            ],
            "default_nodes": [
                {"node_id": "preprocess-text", "author": "functionfly", "name": "preprocess-text", "description": "Normalize and clean text"},
                {"node_id": "toxicity-check", "author": "functionfly", "name": "moderate-content", "description": "Check toxicity with Perspective API"},
                {"node_id": "spam-check", "author": "functionfly", "name": "spam-detection", "description": "Detect spam patterns"},
                {"node_id": "custom-rules", "author": "functionfly", "name": "apply-rules", "description": "Apply custom moderation rules"},
                {"node_id": "decision", "author": "functionfly", "name": "moderation-decision", "description": "Make moderation decision"},
                {"node_id": "queue-review", "author": "functionfly", "name": "queue-moderation", "description": "Queue for human review if needed"},
                {"node_id": "log-decision", "author": "functionfly", "name": "log-moderation", "description": "Log moderation decision"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "preprocess-text", "target_node_id": "toxicity-check"},
                {"id": "e2", "source_node_id": "toxicity-check", "target_node_id": "spam-check"},
                {"id": "e3", "source_node_id": "spam-check", "target_node_id": "custom-rules"},
                {"id": "e4", "source_node_id": "custom-rules", "target_node_id": "decision"},
                {"id": "e5", "source_node_id": "decision", "target_node_id": "queue-review", "condition": {"operator": "eq", "field": "$.decision", "value": "review"}},
                {"id": "e6", "source_node_id": "decision", "target_node_id": "log-decision"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/moderate", "method": "POST"}},
        },
        # ============================================================================
        # REAL-TIME & COMMUNICATION
        # ============================================================================
        "realtime_notifications": {
            "id": "realtime_notifications",
            "name": "Multi-Channel Notification System",
            "description": "Send notifications via email, SMS, push, and Slack with fallback",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["notifications", "email", "sms", "push", "slack", "fallback"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 10,
            "popular_use_cases": [
                "Order updates",
                "Security alerts",
                "Marketing campaigns",
                "System notifications"
            ],
            "default_nodes": [
                {"node_id": "user-preferences", "author": "functionfly", "name": "get-notify-prefs", "description": "Get user notification preferences"},
                {"node_id": "priority-check", "author": "functionfly", "name": "check-priority", "description": "Determine notification priority"},
                {"node_id": "deduplicate", "author": "functionfly", "name": "dedupe-check", "description": "Check for duplicate notifications"},
                {"node_id": "rate-limit-check", "author": "functionfly", "name": "rate-limit", "description": "Check rate limits per channel"},
                {"node_id": "send-primary", "author": "functionfly", "name": "send-notification", "description": "Send via primary channel"},
                {"node_id": "fallback-check", "author": "functionfly", "name": "check-delivery", "description": "Check if primary succeeded"},
                {"node_id": "send-fallback", "author": "functionfly", "name": "send-fallback", "description": "Send via fallback channel"},
                {"node_id": "track-delivery", "author": "functionfly", "name": "track-delivery", "description": "Track delivery status"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "user-preferences", "target_node_id": "priority-check"},
                {"id": "e2", "source_node_id": "priority-check", "target_node_id": "deduplicate"},
                {"id": "e3", "source_node_id": "deduplicate", "target_node_id": "rate-limit-check"},
                {"id": "e4", "source_node_id": "rate-limit-check", "target_node_id": "send-primary"},
                {"id": "e5", "source_node_id": "send-primary", "target_node_id": "fallback-check"},
                {"id": "e6", "source_node_id": "fallback-check", "target_node_id": "send-fallback", "condition": {"operator": "eq", "field": "$.delivered", "value": False}},
                {"id": "e7", "source_node_id": "fallback-check", "target_node_id": "track-delivery", "condition": {"operator": "eq", "field": "$.delivered", "value": True}},
                {"id": "e8", "source_node_id": "send-fallback", "target_node_id": "track-delivery"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/notify", "method": "POST"}},
        },
        "websocket_realtime": {
            "id": "websocket_realtime",
            "name": "WebSocket Real-Time Updates",
            "description": "Real-time WebSocket connections with presence and broadcasting",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["websockets", "realtime", "presence", "broadcast", "collaboration"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 15,
            "popular_use_cases": [
                "Live collaboration",
                "Real-time dashboards",
                "Chat applications",
                "Live notifications"
            ],
            "default_nodes": [
                {"node_id": "auth-connection", "author": "functionfly", "name": "ws-auth", "description": "Authenticate WebSocket connection"},
                {"node_id": "join-room", "author": "functionfly", "name": "ws-join", "description": "Join user to room/channel"},
                {"node_id": "broadcast-presence", "author": "functionfly", "name": "ws-presence", "description": "Broadcast user presence"},
                {"node_id": "process-message", "author": "functionfly", "name": "ws-message", "description": "Process incoming message"},
                {"node_id": "persist-message", "author": "functionfly", "name": "db-insert", "description": "Persist message to DB"},
                {"node_id": "broadcast-message", "author": "functionfly", "name": "ws-broadcast", "description": "Broadcast to room"},
                {"node_id": "delivery-confirm", "author": "functionfly", "name": "ws-ack", "description": "Send delivery confirmation"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "auth-connection", "target_node_id": "join-room"},
                {"id": "e2", "source_node_id": "join-room", "target_node_id": "broadcast-presence"},
                {"id": "e3", "source_node_id": "process-message", "target_node_id": "persist-message"},
                {"id": "e4", "source_node_id": "persist-message", "target_node_id": "broadcast-message"},
                {"id": "e5", "source_node_id": "broadcast-message", "target_node_id": "delivery-confirm"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/ws/message", "method": "POST"}},
        },
        # ============================================================================
        # DATA & ANALYTICS
        # ============================================================================
        "etl_pipeline": {
            "id": "etl_pipeline",
            "name": "ETL Data Pipeline",
            "description": "Extract, transform, and load data between sources with validation",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["etl", "data-pipeline", "csv", "json", "transformation", "batch"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 20,
            "popular_use_cases": [
                "CSV imports",
                "Data migration",
                "Report generation",
                "Third-party sync"
            ],
            "default_nodes": [
                {"node_id": "fetch-source", "author": "functionfly", "name": "fetch-data", "description": "Fetch data from source"},
                {"node_id": "validate-schema", "author": "functionfly", "name": "validate-schema", "description": "Validate data schema"},
                {"node_id": "clean-data", "author": "functionfly", "name": "clean-data", "description": "Clean and normalize data"},
                {"node_id": "transform-fields", "author": "functionfly", "name": "transform-data", "description": "Transform field mappings"},
                {"node_id": "enrich-data", "author": "functionfly", "name": "enrich-data", "description": "Enrich with additional data"},
                {"node_id": "validate-business-rules", "author": "functionfly", "name": "validate-rules", "description": "Validate business rules"},
                {"node_id": "batch-insert", "author": "functionfly", "name": "batch-write", "description": "Batch write to destination"},
                {"node_id": "log-metrics", "author": "functionfly", "name": "log-metrics", "description": "Log ETL metrics"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "fetch-source", "target_node_id": "validate-schema"},
                {"id": "e2", "source_node_id": "validate-schema", "target_node_id": "clean-data"},
                {"id": "e3", "source_node_id": "clean-data", "target_node_id": "transform-fields"},
                {"id": "e4", "source_node_id": "transform-fields", "target_node_id": "enrich-data"},
                {"id": "e5", "source_node_id": "enrich-data", "target_node_id": "validate-business-rules"},
                {"id": "e6", "source_node_id": "validate-business-rules", "target_node_id": "batch-insert"},
                {"id": "e7", "source_node_id": "batch-insert", "target_node_id": "log-metrics"},
            ],
            "trigger": {"type": "schedule", "config": {"cron": "0 */6 * * *"}},  # Every 6 hours
        },
        "analytics_aggregation": {
            "id": "analytics_aggregation",
            "name": "Real-Time Analytics Aggregation",
            "description": "Aggregate events into time-series analytics with real-time dashboards",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["analytics", "metrics", "timeseries", "aggregation", "dashboards"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 15,
            "popular_use_cases": [
                "Product analytics",
                "User behavior tracking",
                "Performance monitoring",
                "Business metrics"
            ],
            "default_nodes": [
                {"node_id": "validate-event", "author": "functionfly", "name": "validate-event", "description": "Validate analytics event"},
                {"node_id": "enrich-event", "author": "functionfly", "name": "enrich-event", "description": "Add geo, device, session data"},
                {"node_id": "write-stream", "author": "functionfly", "name": "write-stream", "description": "Write to event stream"},
                {"node_id": "update-counters", "author": "functionfly", "name": "increment-counter", "description": "Update real-time counters"},
                {"node_id": "update-timeseries", "author": "functionfly", "name": "timeseries-write", "description": "Write to timeseries DB"},
                {"node_id": "trigger-alerts", "author": "functionfly", "name": "check-thresholds", "description": "Check alert thresholds"},
                {"node_id": "update-dashboard", "author": "functionfly", "name": "push-update", "description": "Push update to live dashboard"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-event", "target_node_id": "enrich-event"},
                {"id": "e2", "source_node_id": "enrich-event", "target_node_id": "write-stream"},
                {"id": "e3", "source_node_id": "write-stream", "target_node_id": "update-counters"},
                {"id": "e4", "source_node_id": "update-counters", "target_node_id": "update-timeseries"},
                {"id": "e5", "source_node_id": "update-timeseries", "target_node_id": "trigger-alerts"},
                {"id": "e6", "source_node_id": "update-counters", "target_node_id": "update-dashboard"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/analytics/event", "method": "POST"}},
        },
        "scheduled_reports": {
            "id": "scheduled_reports",
            "name": "Scheduled Report Generation",
            "description": "Generate and deliver periodic reports with data aggregation",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["reports", "scheduled", "pdf", "email", "analytics"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 12,
            "popular_use_cases": [
                "Daily sales reports",
                "Weekly analytics summaries",
                "Monthly invoices",
                "Performance reports"
            ],
            "default_nodes": [
                {"node_id": "fetch-data", "author": "functionfly", "name": "query-report-data", "description": "Query data for report"},
                {"node_id": "aggregate-metrics", "author": "functionfly", "name": "aggregate-data", "description": "Aggregate and calculate metrics"},
                {"node_id": "generate-pdf", "author": "functionfly", "name": "generate-pdf", "description": "Generate PDF report"},
                {"node_id": "generate-csv", "author": "functionfly", "name": "generate-csv", "description": "Generate CSV attachment"},
                {"node_id": "personalize", "author": "functionfly", "name": "personalize-report", "description": "Personalize for recipient"},
                {"node_id": "send-report", "author": "functionfly", "name": "send-email", "description": "Send report via email"},
                {"node_id": "track-delivery", "author": "functionfly", "name": "track-email", "description": "Track email delivery"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "fetch-data", "target_node_id": "aggregate-metrics"},
                {"id": "e2", "source_node_id": "aggregate-metrics", "target_node_id": "generate-pdf"},
                {"id": "e3", "source_node_id": "generate-pdf", "target_node_id": "generate-csv"},
                {"id": "e4", "source_node_id": "generate-csv", "target_node_id": "personalize"},
                {"id": "e5", "source_node_id": "personalize", "target_node_id": "send-report"},
                {"id": "e6", "source_node_id": "send-report", "target_node_id": "track-delivery"},
            ],
            "trigger": {"type": "schedule", "config": {"cron": "0 9 * * 1"}},  # Mondays at 9am
        },
        # ============================================================================
        # ECOMMERCE & MARKETPLACE
        # ============================================================================
        "subscription_management": {
            "id": "subscription_management",
            "name": "Subscription Lifecycle Management",
            "description": "Manage subscription billing cycles, upgrades, downgrades, and cancellations",
            "category": TemplateCategory.MARKETPLACE,
            "tags": ["subscriptions", "billing", "stripe", "upgrades", "saas"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 18,
            "popular_use_cases": [
                "SaaS billing",
                "Membership sites",
                "Content subscriptions",
                "Usage-based billing"
            ],
            "default_nodes": [
                {"node_id": "validate-action", "author": "functionfly", "name": "validate-sub-action", "description": "Validate subscription action"},
                {"node_id": "load-subscription", "author": "functionfly", "name": "get-subscription", "description": "Load subscription from DB"},
                {"node_id": "stripe-api", "author": "functionfly", "name": "stripe-api", "description": "Call Stripe API"},
                {"node_id": "proration-calc", "author": "functionfly", "name": "calculate-proration", "description": "Calculate prorated amounts"},
                {"node_id": "update-db", "author": "functionfly", "name": "update-subscription", "description": "Update subscription in DB"},
                {"node_id": "notify-customer", "author": "functionfly", "name": "send-billing-email", "description": "Send billing notification"},
                {"node_id": "audit-log", "author": "functionfly", "name": "log-billing", "description": "Log billing event"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-action", "target_node_id": "load-subscription"},
                {"id": "e2", "source_node_id": "load-subscription", "target_node_id": "stripe-api"},
                {"id": "e3", "source_node_id": "stripe-api", "target_node_id": "proration-calc"},
                {"id": "e4", "source_node_id": "proration-calc", "target_node_id": "update-db"},
                {"id": "e5", "source_node_id": "update-db", "target_node_id": "notify-customer"},
                {"id": "e6", "source_node_id": "notify-customer", "target_node_id": "audit-log"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/billing/subscription", "method": "POST"}},
        },
        "inventory_management": {
            "id": "inventory_management",
            "name": "Inventory Management System",
            "description": "Track stock levels, handle reservations, and trigger reorder alerts",
            "category": TemplateCategory.MARKETPLACE,
            "tags": ["inventory", "stock", "reservations", "ecommerce", "alerts"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 15,
            "popular_use_cases": [
                "Stock tracking",
                "Pre-orders",
                "Backorders",
                "Multi-warehouse"
            ],
            "default_nodes": [
                {"node_id": "validate-sku", "author": "functionfly", "name": "validate-sku", "description": "Validate SKU exists"},
                {"node_id": "check-stock", "author": "functionfly", "name": "get-stock-level", "description": "Check current stock level"},
                {"node_id": "reserve-stock", "author": "functionfly", "name": "atomic-reserve", "description": "Atomically reserve inventory"},
                {"node_id": "update-reservations", "author": "functionfly", "name": "update-reserved", "description": "Update reservation count"},
                {"node_id": "check-threshold", "author": "functionfly", "name": "check-reorder-threshold", "description": "Check if below reorder point"},
                {"node_id": "trigger-reorder", "author": "functionfly", "name": "send-reorder-alert", "description": "Send reorder alert"},
                {"node_id": "audit-inventory", "author": "functionfly", "name": "log-inventory", "description": "Log inventory change"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-sku", "target_node_id": "check-stock"},
                {"id": "e2", "source_node_id": "check-stock", "target_node_id": "reserve-stock"},
                {"id": "e3", "source_node_id": "reserve-stock", "target_node_id": "update-reservations"},
                {"id": "e4", "source_node_id": "update-reservations", "target_node_id": "check-threshold"},
                {"id": "e5", "source_node_id": "check-threshold", "target_node_id": "trigger-reorder", "condition": {"operator": "eq", "field": "$.below_threshold", "value": True}},
                {"id": "e6", "source_node_id": "check-threshold", "target_node_id": "audit-inventory"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/inventory/reserve", "method": "POST"}},
        },
        "refund_processor": {
            "id": "refund_processor",
            "name": "Refund and Return Processor",
            "description": "Process refunds, returns, and exchanges with approval workflows",
            "category": TemplateCategory.MARKETPLACE,
            "tags": ["refunds", "returns", "stripe", "workflow", "approvals"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 12,
            "popular_use_cases": [
                "Return processing",
                "Refund automation",
                "Exchange handling",
                "Quality checks"
            ],
            "default_nodes": [
                {"node_id": "validate-request", "author": "functionfly", "name": "validate-refund", "description": "Validate refund request"},
                {"node_id": "load-order", "author": "functionfly", "name": "get-order", "description": "Load order details"},
                {"node_id": "check-policy", "author": "functionfly", "name": "check-refund-policy", "description": "Check refund policy"},
                {"node_id": "approve-refund", "author": "functionfly", "name": "approve-refund", "description": "Approve or reject refund"},
                {"node_id": "stripe-refund", "author": "functionfly", "name": "process-refund", "description": "Process Stripe refund"},
                {"node_id": "update-order", "author": "functionfly", "name": "update-order-status", "description": "Update order status"},
                {"node_id": "notify-customer", "author": "functionfly", "name": "send-refund-email", "description": "Send refund confirmation"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-request", "target_node_id": "load-order"},
                {"id": "e2", "source_node_id": "load-order", "target_node_id": "check-policy"},
                {"id": "e3", "source_node_id": "check-policy", "target_node_id": "approve-refund"},
                {"id": "e4", "source_node_id": "approve-refund", "target_node_id": "stripe-refund", "condition": {"operator": "eq", "field": "$.approved", "value": True}},
                {"id": "e5", "source_node_id": "stripe-refund", "target_node_id": "update-order"},
                {"id": "e6", "source_node_id": "update-order", "target_node_id": "notify-customer"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/refund", "method": "POST"}},
        },
        # ============================================================================
        # FILE & MEDIA HANDLING
        # ============================================================================
        "file_upload_pipeline": {
            "id": "file_upload_pipeline",
            "name": "Secure File Upload Pipeline",
            "description": "Handle file uploads with virus scanning, image optimization, and CDN distribution",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["files", "upload", "s3", "cdn", "virus-scan", "optimization"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 10,
            "popular_use_cases": [
                "Image uploads",
                "Document uploads",
                "Video processing",
                "Asset management"
            ],
            "default_nodes": [
                {"node_id": "validate-upload", "author": "functionfly", "name": "validate-upload", "description": "Validate file type and size"},
                {"node_id": "auth-check", "author": "functionfly", "name": "auth-middleware", "description": "Check upload permissions"},
                {"node_id": "virus-scan", "author": "functionfly", "name": "virus-scan", "description": "Scan for malware"},
                {"node_id": "generate-thumbnail", "author": "functionfly", "name": "generate-thumbnail", "description": "Generate image thumbnail"},
                {"node_id": "optimize-image", "author": "functionfly", "name": "optimize-image", "description": "Optimize for web"},
                {"node_id": "upload-s3", "author": "functionfly", "name": "upload-s3", "description": "Upload to S3"},
                {"node_id": "invalidate-cdn", "author": "functionfly", "name": "invalidate-cdn", "description": "Invalidate CDN cache"},
                {"node_id": "save-metadata", "author": "functionfly", "name": "save-file-metadata", "description": "Save file metadata to DB"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "auth-check", "target_node_id": "validate-upload"},
                {"id": "e2", "source_node_id": "validate-upload", "target_node_id": "virus-scan"},
                {"id": "e3", "source_node_id": "virus-scan", "target_node_id": "generate-thumbnail"},
                {"id": "e4", "source_node_id": "generate-thumbnail", "target_node_id": "optimize-image"},
                {"id": "e5", "source_node_id": "optimize-image", "target_node_id": "upload-s3"},
                {"id": "e6", "source_node_id": "upload-s3", "target_node_id": "invalidate-cdn"},
                {"id": "e7", "source_node_id": "invalidate-cdn", "target_node_id": "save-metadata"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/upload", "method": "POST"}},
        },
        "video_processing": {
            "id": "video_processing",
            "name": "Video Processing Pipeline",
            "description": "Transcode videos, generate thumbnails, and create adaptive bitrate streams",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["video", "transcoding", "ffmpeg", "streaming", "thumbnails"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 20,
            "popular_use_cases": [
                "Video platforms",
                "Course content",
                "Social video",
                "Live streaming"
            ],
            "default_nodes": [
                {"node_id": "validate-video", "author": "functionfly", "name": "validate-video", "description": "Validate video format"},
                {"node_id": "extract-metadata", "author": "functionfly", "name": "extract-metadata", "description": "Extract video metadata"},
                {"node_id": "transcode-720p", "author": "functionfly", "name": "transcode-video", "description": "Transcode to 720p"},
                {"node_id": "transcode-1080p", "author": "functionfly", "name": "transcode-video", "description": "Transcode to 1080p"},
                {"node_id": "generate-thumbnail", "author": "functionfly", "name": "video-thumbnail", "description": "Generate poster image"},
                {"node_id": "upload-variants", "author": "functionfly", "name": "upload-variants", "description": "Upload all variants"},
                {"node_id": "create-manifest", "author": "functionfly", "name": "create-hls-manifest", "description": "Create HLS manifest"},
                {"node_id": "notify-complete", "author": "functionfly", "name": "notify-complete", "description": "Notify processing complete"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "validate-video", "target_node_id": "extract-metadata"},
                {"id": "e2", "source_node_id": "extract-metadata", "target_node_id": "transcode-720p"},
                {"id": "e3", "source_node_id": "transcode-720p", "target_node_id": "transcode-1080p"},
                {"id": "e4", "source_node_id": "extract-metadata", "target_node_id": "generate-thumbnail"},
                {"id": "e5", "source_node_id": "transcode-1080p", "target_node_id": "upload-variants"},
                {"id": "e6", "source_node_id": "generate-thumbnail", "target_node_id": "upload-variants"},
                {"id": "e7", "source_node_id": "upload-variants", "target_node_id": "create-manifest"},
                {"id": "e8", "source_node_id": "create-manifest", "target_node_id": "notify-complete"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/video/process", "method": "POST"}},
        },
        # ============================================================================
        # SEARCH & DISCOVERY
        # ============================================================================
        "search_engine": {
            "id": "search_engine",
            "name": "Full-Text Search Engine",
            "description": "Implement search with autocomplete, faceting, and relevance ranking",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["search", "elasticsearch", "algolia", "facets", "autocomplete"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 15,
            "popular_use_cases": [
                "Product search",
                "Content discovery",
                "Document search",
                "User directory"
            ],
            "default_nodes": [
                {"node_id": "parse-query", "author": "functionfly", "name": "parse-search-query", "description": "Parse and sanitize query"},
                {"node_id": "check-cache", "author": "functionfly", "name": "cache-get", "description": "Check search cache"},
                {"node_id": "execute-search", "author": "functionfly", "name": "elasticsearch-query", "description": "Execute Elasticsearch query"},
                {"node_id": "apply-facets", "author": "functionfly", "name": "apply-facets", "description": "Apply facet filters"},
                {"node_id": "rank-results", "author": "functionfly", "name": "rank-results", "description": "Apply relevance ranking"},
                {"node_id": "format-response", "author": "functionfly", "name": "format-search", "description": "Format search response"},
                {"node_id": "cache-results", "author": "functionfly", "name": "cache-set", "description": "Cache search results"},
                {"node_id": "log-search", "author": "functionfly", "name": "log-search", "description": "Log search analytics"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "parse-query", "target_node_id": "check-cache"},
                {"id": "e2", "source_node_id": "check-cache", "target_node_id": "execute-search", "condition": {"operator": "eq", "field": "$.cached", "value": False}},
                {"id": "e3", "source_node_id": "check-cache", "target_node_id": "format-response", "condition": {"operator": "eq", "field": "$.cached", "value": True}},
                {"id": "e4", "source_node_id": "execute-search", "target_node_id": "apply-facets"},
                {"id": "e5", "source_node_id": "apply-facets", "target_node_id": "rank-results"},
                {"id": "e6", "source_node_id": "rank-results", "target_node_id": "format-response"},
                {"id": "e7", "source_node_id": "format-response", "target_node_id": "cache-results"},
                {"id": "e8", "source_node_id": "cache-results", "target_node_id": "log-search"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/search", "method": "GET"}},
        },
        "recommendation_engine": {
            "id": "recommendation_engine",
            "name": "Personalized Recommendation Engine",
            "description": "Generate personalized recommendations using collaborative filtering",
            "category": TemplateCategory.API_BACKEND,
            "tags": ["recommendations", "personalization", "ml", "discovery", "collaborative-filtering"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 20,
            "popular_use_cases": [
                "Product recommendations",
                "Content suggestions",
                "Similar items",
                "Trending discovery"
            ],
            "default_nodes": [
                {"node_id": "load-user-profile", "author": "functionfly", "name": "get-user-profile", "description": "Load user preferences"},
                {"node_id": "fetch-history", "author": "functionfly", "name": "get-interactions", "description": "Fetch user interaction history"},
                {"node_id": "check-cache", "author": "functionfly", "name": "cache-get", "description": "Check recommendation cache"},
                {"node_id": "embed-user", "author": "functionfly", "name": "embed-user", "description": "Generate user embedding"},
                {"node_id": "similar-users", "author": "functionfly", "name": "find-similar", "description": "Find similar users"},
                {"node_id": "score-items", "author": "functionfly", "name": "score-candidates", "description": "Score candidate items"},
                {"node_id": "diversify", "author": "functionfly", "name": "diversify-results", "description": "Apply diversity to results"},
                {"node_id": "format-response", "author": "functionfly", "name": "format-recs", "description": "Format recommendations"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "load-user-profile", "target_node_id": "fetch-history"},
                {"id": "e2", "source_node_id": "fetch-history", "target_node_id": "check-cache"},
                {"id": "e3", "source_node_id": "check-cache", "target_node_id": "embed-user", "condition": {"operator": "eq", "field": "$.cached", "value": False}},
                {"id": "e4", "source_node_id": "check-cache", "target_node_id": "format-response", "condition": {"operator": "eq", "field": "$.cached", "value": True}},
                {"id": "e5", "source_node_id": "embed-user", "target_node_id": "similar-users"},
                {"id": "e6", "source_node_id": "similar-users", "target_node_id": "score-items"},
                {"id": "e7", "source_node_id": "score-items", "target_node_id": "diversify"},
                {"id": "e8", "source_node_id": "diversify", "target_node_id": "format-response"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/recommendations", "method": "GET"}},
        },
        # ============================================================================
        # MONITORING & OBSERVABILITY
        # ============================================================================
        "health_checks": {
            "id": "health_checks",
            "name": "System Health Check Aggregator",
            "description": "Aggregate health checks from multiple services with alerting",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["health", "monitoring", "alerts", "status", "observability"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 10,
            "popular_use_cases": [
                "Service status pages",
                "Health dashboards",
                "Incident detection",
                "SLA monitoring"
            ],
            "default_nodes": [
                {"node_id": "check-database", "author": "functionfly", "name": "db-health", "description": "Check database connectivity"},
                {"node_id": "check-redis", "author": "functionfly", "name": "redis-health", "description": "Check Redis connectivity"},
                {"node_id": "check-external", "author": "functionfly", "name": "external-health", "description": "Check external APIs"},
                {"node_id": "aggregate-status", "author": "functionfly", "name": "aggregate-health", "description": "Aggregate all health checks"},
                {"node_id": "check-degraded", "author": "functionfly", "name": "check-degraded", "description": "Check if any service degraded"},
                {"node_id": "send-alert", "author": "functionfly", "name": "send-pagerduty", "description": "Send PagerDuty alert if degraded"},
                {"node_id": "update-status", "author": "functionfly", "name": "update-status-page", "description": "Update status page"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "check-database", "target_node_id": "aggregate-status"},
                {"id": "e2", "source_node_id": "check-redis", "target_node_id": "aggregate-status"},
                {"id": "e3", "source_node_id": "check-external", "target_node_id": "aggregate-status"},
                {"id": "e4", "source_node_id": "aggregate-status", "target_node_id": "check-degraded"},
                {"id": "e5", "source_node_id": "check-degraded", "target_node_id": "send-alert", "condition": {"operator": "eq", "field": "$.degraded", "value": True}},
                {"id": "e6", "source_node_id": "aggregate-status", "target_node_id": "update-status"},
            ],
            "trigger": {"type": "schedule", "config": {"cron": "*/5 * * * *"}},  # Every 5 minutes
        },
        "log_aggregation": {
            "id": "log_aggregation",
            "name": "Log Aggregation and Analysis",
            "description": "Collect, parse, and analyze logs with alerting for errors",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["logs", "monitoring", "elasticsearch", "alerts", "parsing"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 12,
            "popular_use_cases": [
                "Centralized logging",
                "Error tracking",
                "Audit logging",
                "Compliance"
            ],
            "default_nodes": [
                {"node_id": "parse-log", "author": "functionfly", "name": "parse-log", "description": "Parse log entry"},
                {"node_id": "extract-fields", "author": "functionfly", "name": "extract-fields", "description": "Extract structured fields"},
                {"node_id": "enrich-context", "author": "functionfly", "name": "enrich-log", "description": "Add service context"},
                {"node_id": "check-error", "author": "functionfly", "name": "detect-error", "description": "Detect error level"},
                {"node_id": "index-log", "author": "functionfly", "name": "index-elasticsearch", "description": "Index to Elasticsearch"},
                {"node_id": "alert-error", "author": "functionfly", "name": "send-error-alert", "description": "Send error alert"},
                {"node_id": "archive-cold", "author": "functionfly", "name": "archive-s3", "description": "Archive to cold storage"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "parse-log", "target_node_id": "extract-fields"},
                {"id": "e2", "source_node_id": "extract-fields", "target_node_id": "enrich-context"},
                {"id": "e3", "source_node_id": "enrich-context", "target_node_id": "check-error"},
                {"id": "e4", "source_node_id": "check-error", "target_node_id": "index-log"},
                {"id": "e5", "source_node_id": "check-error", "target_node_id": "alert-error", "condition": {"operator": "eq", "field": "$.is_error", "value": True}},
                {"id": "e6", "source_node_id": "index-log", "target_node_id": "archive-cold"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/logs/ingest", "method": "POST"}},
        },
        # ============================================================================
        # WORKFLOW AUTOMATION
        # ============================================================================
        "approval_workflow": {
            "id": "approval_workflow",
            "name": "Multi-Stage Approval Workflow",
            "description": "Route requests through multi-level approval chains with delegation",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["workflow", "approvals", "bpm", "automation", "routing"],
            "complexity": "complex",
            "estimated_setup_time_minutes": 18,
            "popular_use_cases": [
                "Expense approvals",
                "Purchase orders",
                "Content publishing",
                "Access requests"
            ],
            "default_nodes": [
                {"node_id": "submit-request", "author": "functionfly", "name": "create-request", "description": "Create approval request"},
                {"node_id": "find-approvers", "author": "functionfly", "name": "find-approvers", "description": "Find approvers by rules"},
                {"node_id": "notify-approver", "author": "functionfly", "name": "notify-approver", "description": "Notify first approver"},
                {"node_id": "wait-approval", "author": "functionfly", "name": "wait-response", "description": "Wait for approval"},
                {"node_id": "check-decision", "author": "functionfly", "name": "check-decision", "description": "Check approval decision"},
                {"node_id": "next-level", "author": "functionfly", "name": "next-approver", "description": "Route to next level"},
                {"node_id": "execute-action", "author": "functionfly", "name": "execute-approved", "description": "Execute approved action"},
                {"node_id": "notify-result", "author": "functionfly", "name": "notify-requester", "description": "Notify requester"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "submit-request", "target_node_id": "find-approvers"},
                {"id": "e2", "source_node_id": "find-approvers", "target_node_id": "notify-approver"},
                {"id": "e3", "source_node_id": "notify-approver", "target_node_id": "wait-approval"},
                {"id": "e4", "source_node_id": "wait-approval", "target_node_id": "check-decision"},
                {"id": "e5", "source_node_id": "check-decision", "target_node_id": "next-level", "condition": {"operator": "eq", "field": "$.decision", "value": "approved"}},
                {"id": "e6", "source_node_id": "check-decision", "target_node_id": "notify-result", "condition": {"operator": "eq", "field": "$.decision", "value": "rejected"}},
                {"id": "e7", "source_node_id": "next-level", "target_node_id": "execute-action"},
                {"id": "e8", "source_node_id": "execute-action", "target_node_id": "notify-result"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/workflow/approval", "method": "POST"}},
        },
        "crm_integration": {
            "id": "crm_integration",
            "name": "CRM Integration Sync",
            "description": "Sync contacts and interactions between your app and CRM platforms",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["crm", "salesforce", "hubspot", "sync", "integration"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 12,
            "popular_use_cases": [
                "Contact sync",
                "Lead tracking",
                "Activity logging",
                "Sales pipeline"
            ],
            "default_nodes": [
                {"node_id": "normalize-contact", "author": "functionfly", "name": "normalize-contact", "description": "Normalize contact data"},
                {"node_id": "duplicate-check", "author": "functionfly", "name": "dedupe-contact", "description": "Check for duplicates"},
                {"node_id": "enrich-data", "author": "functionfly", "name": "enrich-contact", "description": "Enrich with Clearbit"},
                {"node_id": "map-fields", "author": "functionfly", "name": "map-crm-fields", "description": "Map to CRM fields"},
                {"node_id": "sync-crm", "author": "functionfly", "name": "sync-salesforce", "description": "Sync to Salesforce/HubSpot"},
                {"node_id": "local-save", "author": "functionfly", "name": "save-contact", "description": "Save to local DB"},
                {"node_id": "log-sync", "author": "functionfly", "name": "log-crm-sync", "description": "Log sync activity"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "normalize-contact", "target_node_id": "duplicate-check"},
                {"id": "e2", "source_node_id": "duplicate-check", "target_node_id": "enrich-data"},
                {"id": "e3", "source_node_id": "enrich-data", "target_node_id": "map-fields"},
                {"id": "e4", "source_node_id": "map-fields", "target_node_id": "sync-crm"},
                {"id": "e5", "source_node_id": "map-fields", "target_node_id": "local-save"},
                {"id": "e6", "source_node_id": "sync-crm", "target_node_id": "log-sync"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/crm/sync", "method": "POST"}},
        },
        "slack_automation": {
            "id": "slack_automation",
            "name": "Slack Bot Automation",
            "description": "Build interactive Slack bots with commands, notifications, and workflows",
            "category": TemplateCategory.WEBHOOK_PROCESSOR,
            "tags": ["slack", "bot", "chatops", "automation", "notifications"],
            "complexity": "moderate",
            "estimated_setup_time_minutes": 10,
            "popular_use_cases": [
                "DevOps commands",
                "Support tickets",
                "Team notifications",
                "Daily standups"
            ],
            "default_nodes": [
                {"node_id": "verify-request", "author": "functionfly", "name": "verify-slack-sig", "description": "Verify Slack signature"},
                {"node_id": "parse-command", "author": "functionfly", "name": "parse-slash-cmd", "description": "Parse slash command"},
                {"node_id": "validate-user", "author": "functionfly", "name": "check-slack-auth", "description": "Check user permissions"},
                {"node_id": "execute-command", "author": "functionfly", "name": "run-cmd", "description": "Execute command logic"},
                {"node_id": "format-response", "author": "functionfly", "name": "format-slack-msg", "description": "Format Slack message"},
                {"node_id": "send-immediate", "author": "functionfly", "name": "send-slack-msg", "description": "Send immediate response"},
                {"node_id": "async-followup", "author": "functionfly", "name": "delayed-msg", "description": "Send delayed follow-up"},
            ],
            "default_edges": [
                {"id": "e1", "source_node_id": "verify-request", "target_node_id": "parse-command"},
                {"id": "e2", "source_node_id": "parse-command", "target_node_id": "validate-user"},
                {"id": "e3", "source_node_id": "validate-user", "target_node_id": "execute-command"},
                {"id": "e4", "source_node_id": "execute-command", "target_node_id": "format-response"},
                {"id": "e5", "source_node_id": "format-response", "target_node_id": "send-immediate"},
                {"id": "e6", "source_node_id": "execute-command", "target_node_id": "async-followup"},
            ],
            "trigger": {"type": "webhook", "config": {"path": "/api/slack/command", "method": "POST"}},
        },
    }

    def __init__(self):
        self._provider_manager = None

    def _get_provider(self):
        """Lazy initialize provider manager."""
        if self._provider_manager is None:
            self._provider_manager = get_provider_manager()
        return self._provider_manager

    def list_templates(self) -> List[GraphTemplateInfo]:
        """List all available prebuilt templates."""
        templates = []
        for template_id, template in self.TEMPLATES.items():
            templates.append(GraphTemplateInfo(
                id=template_id,
                name=template["name"],
                description=template["description"],
                category=template["category"],
                tags=template["tags"],
                node_count=len(template["default_nodes"]),
                complexity=template["complexity"],
                estimated_setup_time_minutes=template["estimated_setup_time_minutes"],
                popular_use_cases=template["popular_use_cases"],
            ))
        return templates

    def get_template(self, template_id: str) -> Optional[Dict[str, Any]]:
        """Get a specific template by ID."""
        return self.TEMPLATES.get(template_id)

    async def compose_from_template(
        self,
        template_id: str,
        customization_prompt: Optional[str] = None,
        tenant_id: Optional[str] = None,
    ) -> GraphCompositionResponse:
        """Generate a graph from a prebuilt template with optional customization."""
        generation_id = f"tpl-{int(time.time() * 1000)}"
        start_time = time.time()

        template = self.TEMPLATES.get(template_id)
        if not template:
            return GraphCompositionResponse(
                success=False,
                generation_id=generation_id,
                latency_ms=(time.time() - start_time) * 1000,
                confidence=0.0,
                error=f"Template not found: {template_id}",
            )

        # Build nodes from template
        nodes = []
        for node_def in template["default_nodes"]:
            nodes.append(GraphNodeRef(
                node_id=node_def["node_id"],
                author=node_def["author"],
                name=node_def["name"],
                version="latest",
                config={},
                description=node_def.get("description"),
            ))

        # Build edges from template
        edges = []
        for edge_def in template["default_edges"]:
            edges.append(GraphEdge(
                id=edge_def["id"],
                source_node_id=edge_def["source_node_id"],
                target_node_id=edge_def["target_node_id"],
                mapping=GraphEdgeMapping(),
                condition=None,
                type="sync",
            ))

        # Apply customizations if provided
        if customization_prompt:
            # TODO: Use LLM to customize template based on prompt
            # For now, just add a note
            pass

        graph = GraphDefinition(
            name=template_id.replace("_", "-"),
            description=template["description"],
            execution_mode="event_driven" if template["trigger"]["type"] == "webhook" else "sync",
            nodes=nodes,
            edges=edges,
            trigger_config=GraphTriggerConfig(**template["trigger"]) if template.get("trigger") else None,
            visibility="public",
        )

        explanation = GraphCompositionExplanation(
            summary=f"Graph based on '{template['name']}' template",
            node_purposes={n.node_id: n.description or "" for n in nodes},
            data_flow_description=self._describe_data_flow(nodes, edges),
            trigger_explanation=f"Triggered by {template['trigger']['type']}" if template.get("trigger") else "Manual trigger",
            suggested_tests=[
                "Test successful flow end-to-end",
                "Test error handling at each node",
                "Test trigger webhook/schedule",
            ],
        )

        return GraphCompositionResponse(
            success=True,
            graph=graph,
            explanation=explanation,
            confidence=0.95,
            generation_id=generation_id,
            latency_ms=(time.time() - start_time) * 1000,
            tokens_used={"prompt": 0, "completion": 0, "total": 0},
        )

    async def compose_from_prompt(
        self,
        request: GraphCompositionRequest,
        api_key_info: Optional[APIKeyInfo] = None,
    ) -> GraphCompositionResponse:
        """Generate a graph from natural language prompt using LLM.

        This is the core AI composition feature:
        1. Parse intent from prompt
        2. Select relevant functions from RAG/catalog
        3. Generate node topology
        4. Suggest edge connections
        5. Return complete graph definition
        """
        generation_id = f"ai-{int(time.time() * 1000)}"
        start_time = time.time()
        attempts: List[CompositionAttempt] = []

        try:
            # Step 1: Check if prompt matches a template directly
            template_match = self._match_prompt_to_template(request.prompt)
            if template_match and not request.requirements:
                # Use template for common patterns
                return await self.compose_from_template(
                    template_match,
                    customization_prompt=request.prompt,
                    tenant_id=request.tenant_id,
                )

            # Step 2: Use LLM to generate graph topology
            provider_manager = self._get_provider()
            provider = provider_manager.get_provider("openai")

            if not provider or not provider.available:
                return GraphCompositionResponse(
                    success=False,
                    generation_id=generation_id,
                    latency_ms=(time.time() - start_time) * 1000,
                    confidence=0.0,
                    error="AI generation service not available",
                )

            # Build the composition prompt
            system_prompt = self._build_system_prompt(request.preferred_runtime)
            user_prompt = self._build_user_prompt(request)

            messages = [
                ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
                ChatMessage(role=MessageRole.USER, content=user_prompt),
            ]

            # Generate with GPT-4
            completion = await provider.complete(
                messages=messages,
                model="gpt-4o",
                temperature=0.3,  # Lower temperature for structured output
                max_tokens=4000,
            )

            # Parse the JSON response
            try:
                result_data = json.loads(completion.content)
            except json.JSONDecodeError:
                # Try to extract JSON from markdown
                content = completion.content
                if "```json" in content:
                    json_str = content.split("```json")[1].split("```")[0].strip()
                elif "```" in content:
                    json_str = content.split("```")[1].split("```")[0].strip()
                else:
                    # Try to find JSON object
                    start = content.find("{")
                    end = content.rfind("}")
                    if start != -1 and end != -1:
                        json_str = content[start:end+1]
                    else:
                        raise ValueError(f"No JSON found in response: {content[:200]}")
                result_data = json.loads(json_str)

            # Build graph from parsed data
            graph = self._build_graph_from_llm_output(result_data)
            explanation = self._build_explanation_from_llm_output(result_data)

            latency_ms = (time.time() - start_time) * 1000

            return GraphCompositionResponse(
                success=True,
                graph=graph,
                explanation=explanation,
                confidence=result_data.get("confidence", 0.8),
                generation_id=generation_id,
                latency_ms=latency_ms,
                tokens_used=completion.usage,
                suggestions=result_data.get("suggested_improvements", []),
            )

        except Exception as e:
            logger.error(f"Graph composition failed: {e}")
            latency_ms = (time.time() - start_time) * 1000

            return GraphCompositionResponse(
                success=False,
                generation_id=generation_id,
                latency_ms=latency_ms,
                confidence=0.0,
                error=str(e),
                suggestions=["Try rephrasing your prompt with more specific details"],
            )

    def _match_prompt_to_template(self, prompt: str) -> Optional[str]:
        """Check if prompt matches a known template pattern."""
        prompt_lower = prompt.lower()

        # Simple keyword matching (can be improved with embeddings)
        if any(kw in prompt_lower for kw in ["saas", "signup", "subscription", "billing", "stripe customer"]):
            return "saas_starter"
        elif any(kw in prompt_lower for kw in ["ecommerce", "checkout", "cart", "payment", "order", "receipt"]):
            return "ecommerce_checkout"
        elif any(kw in prompt_lower for kw in ["api", "crud", "rest", "backend", "database", "resource"]):
            return "api_backend"
        elif any(kw in prompt_lower for kw in ["webhook", "stripe webhook", "github webhook", "event processor"]):
            return "webhook_processor"

        return None

    def _build_system_prompt(self, preferred_runtime: str) -> str:
        """Build the system prompt for graph composition."""
        return f"""You are an expert backend architect specializing in "Backend as a Graph" - composing serverless functions into reactive workflows.

Your task is to generate a complete graph definition from a natural language description.

Key concepts:
- Graph = Directed acyclic graph (DAG) of function nodes connected by edges
- Node = A serverless function call (with author/name/version)
- Edge = Data flow between nodes with optional transformations
- Trigger = What starts the graph (webhook, schedule, database change)

Guidelines:
1. Break down workflows into logical steps (nodes)
2. Connect them with appropriate data flow (edges)
3. Use common patterns: validate → process → store → notify
4. Prefer existing functionfly functions when available
5. Add error handling nodes where appropriate

Respond with a JSON object containing:
{{
  "graph": {{
    "name": "url-friendly-name",
    "description": "What this graph does",
    "execution_mode": "sync|async|event_driven",
    "nodes": [
      {{
        "node_id": "unique-id",
        "author": "functionfly|user",
        "name": "function-name",
        "version": "latest",
        "description": "What this node does"
      }}
    ],
    "edges": [
      {{
        "id": "e1",
        "source_node_id": "node-a",
        "target_node_id": "node-b",
        "type": "sync"
      }}
    ],
    "trigger_config": {{
      "type": "webhook|schedule|manual",
      "config": {{}}
    }}
  }},
  "explanation": {{
    "summary": "Brief overview",
    "node_purposes": {{"node-id": "description"}},
    "data_flow_description": "How data flows through the graph",
    "trigger_explanation": "When this graph runs"
  }},
  "confidence": 0.85,
  "suggested_improvements": ["Optional improvement suggestions"]
}}

Use {preferred_runtime} as the preferred runtime when suggesting functions."""

    def _build_user_prompt(self, request: GraphCompositionRequest) -> str:
        """Build the user prompt from the composition request."""
        prompt_parts = [
            f"Create a graph for: {request.prompt}",
        ]

        if request.requirements:
            prompt_parts.append(f"\nRequirements: {', '.join(request.requirements)}")

        prompt_parts.append("\nGenerate the complete graph definition with nodes and edges.")

        return "\n".join(prompt_parts)

    def _build_graph_from_llm_output(self, data: Dict[str, Any]) -> GraphDefinition:
        """Convert LLM output to GraphDefinition model."""
        graph_data = data.get("graph", {})

        nodes = []
        for node_def in graph_data.get("nodes", []):
            nodes.append(GraphNodeRef(
                node_id=node_def["node_id"],
                author=node_def.get("author", "functionfly"),
                name=node_def["name"],
                version=node_def.get("version", "latest"),
                config=node_def.get("config", {}),
                description=node_def.get("description"),
            ))

        edges = []
        for edge_def in graph_data.get("edges", []):
            mapping_data = edge_def.get("mapping", {})
            mapping = GraphEdgeMapping(
                source_path=mapping_data.get("source_path"),
                target_path=mapping_data.get("target_path"),
                transform=mapping_data.get("transform"),
                script=mapping_data.get("script"),
            ) if mapping_data else GraphEdgeMapping()

            edges.append(GraphEdge(
                id=edge_def["id"],
                source_node_id=edge_def["source_node_id"],
                target_node_id=edge_def["target_node_id"],
                mapping=mapping,
                condition=None,  # TODO: Parse conditions
                type=edge_def.get("type", "sync"),
                fallback_node_id=edge_def.get("fallback_node_id"),
            ))

        trigger_data = graph_data.get("trigger_config")
        trigger_config = None
        if trigger_data:
            trigger_config = GraphTriggerConfig(
                type=trigger_data["type"],
                config=trigger_data.get("config", {}),
            )

        return GraphDefinition(
            name=graph_data.get("name", "generated-graph"),
            description=graph_data.get("description", "AI-generated graph"),
            execution_mode=graph_data.get("execution_mode", "sync"),
            nodes=nodes,
            edges=edges,
            input_schema=graph_data.get("input_schema"),
            output_schema=graph_data.get("output_schema"),
            trigger_config=trigger_config,
            visibility="public",
        )

    def _build_explanation_from_llm_output(self, data: Dict[str, Any]) -> GraphCompositionExplanation:
        """Convert LLM output to GraphCompositionExplanation model."""
        exp_data = data.get("explanation", {})

        return GraphCompositionExplanation(
            summary=exp_data.get("summary", "AI-generated graph"),
            node_purposes=exp_data.get("node_purposes", {}),
            data_flow_description=exp_data.get("data_flow_description", ""),
            trigger_explanation=exp_data.get("trigger_explanation", ""),
            suggested_tests=exp_data.get("suggested_tests", [
                "Test successful flow end-to-end",
                "Test error handling at each node",
            ]),
        )

    def _describe_data_flow(self, nodes: List[GraphNodeRef], edges: List[GraphEdge]) -> str:
        """Generate a human-readable description of data flow."""
        if not edges:
            return "No connections defined"

        flow_parts = []
        for edge in edges:
            source = next((n for n in nodes if n.node_id == edge.source_node_id), None)
            target = next((n for n in nodes if n.node_id == edge.target_node_id), None)
            if source and target:
                flow_parts.append(f"{source.node_id} → {target.node_id}")

        return " → ".join(flow_parts) if len(flow_parts) <= 3 else "\n".join(flow_parts)


# Singleton instance
_composition_service: Optional[GraphCompositionService] = None


def get_graph_composition_service() -> GraphCompositionService:
    """Get the singleton graph composition service."""
    global _composition_service
    if _composition_service is None:
        _composition_service = GraphCompositionService()
    return _composition_service
