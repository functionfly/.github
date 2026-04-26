#!/bin/bash
# Seed initial unfair advantage opportunities
# Usage: ./scripts/seed_unfair_advantage.sh <admin-api-key>
set -e

API_URL="${API_URL:-http://localhost:8080}"
ADMIN_KEY="${1:-}"

if [ -z "$ADMIN_KEY" ]; then
    echo "Usage: $0 <admin-api-key>"
    echo "Or set ADMIN_KEY environment variable"
    exit 1
fi

echo "Seeding unfair advantage opportunities..."

# Seed Real-time Collaboration Engine
curl -s -X POST "$API_URL/admin/unfair-advantage/opportunities/seed" \
    -H "Authorization: Bearer $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Real-time Collaboration Engine",
        "description": "Multi-user real-time code editing with operational transformation, similar to Figma/Google Docs for code. Includes presence awareness, cursor tracking, conflict resolution, and session persistence.",
        "category": "revenue_boost",
        "tags": ["collaboration", "real-time", "websocket", "operational-transform"],
        "estimated_value": 75000,
        "priority": "critical",
        "seeded_by": "functionfly-team",
        "business_impact": "Can become a flagship feature that drives new business",
        "competitive_edge": "No competitor offers real-time collaborative function editing with OT",
        "time_to_market": "4-6 months"
    }' | jq .

# Seed AI-Powered Function Recommender
curl -s -X POST "$API_URL/admin/unfair-advantage/opportunities/seed" \
    -H "Authorization: Bearer $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "AI-Powered Function Recommender",
        "description": "Based on user code context and natural language queries, recommend relevant functions from the marketplace. Uses embedding similarity + LLM reranking.",
        "category": "revenue_boost",
        "tags": ["ai", "recommendation", "embeddings", "llm"],
        "estimated_value": 50000,
        "priority": "high",
        "seeded_by": "functionfly-team",
        "business_impact": "Increases marketplace conversion rate and time-to-value",
        "competitive_edge": "Personalized recommendations based on usage patterns",
        "time_to_market": "2-3 months"
    }' | jq .

# Seed Edge Analytics Dashboard
curl -s -X POST "$API_URL/admin/unfair-advantage/opportunities/seed" \
    -H "Authorization: Bearer $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Edge Analytics Dashboard",
        "description": "Real-time geographic visualization of function execution patterns, latency heatmaps, error rates by region, and CDN cache hit ratios.",
        "category": "speed_gain",
        "tags": ["analytics", "visualization", "edge", "real-time", "geo"],
        "estimated_value": 30000,
        "priority": "high",
        "seeded_by": "functionfly-team",
        "business_impact": "Reduces debugging time and helps customers optimize",
        "competitive_edge": "Native integration with FunctionFly edge network",
        "time_to_market": "1-2 months"
    }' | jq .

# Seed Automated Cost Anomaly Detection
curl -s -X POST "$API_URL/admin/unfair-advantage/opportunities/seed" \
    -H "Authorization: Bearer $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Automated Cost Anomaly Detection",
        "description": "ML-based detection of unusual usage patterns that might indicate issues or opportunities. Automatically diagnoses root cause and suggests remediation.",
        "category": "cost_savings",
        "tags": ["ml", "anomaly-detection", "cost-optimization", "automation"],
        "estimated_value": 40000,
        "priority": "high",
        "seeded_by": "functionfly-team",
        "business_impact": "Reduces support tickets and prevents bill shock",
        "competitive_edge": "Deep integration with billing and usage data",
        "time_to_market": "2 months"
    }' | jq .

# Seed Universal WASM Pipeline
curl -s -X POST "$API_URL/admin/unfair-advantage/opportunities/seed" \
    -H "Authorization: Bearer $ADMIN_KEY" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Universal WASM Pipeline",
        "description": "Polyglot serverless runtime accepting Rust, Go, C, Ruby, Kotlin, Swift compiled to WASM. Enables near-native performance with full portability.",
        "category": "competitive_moat",
        "tags": ["wasm", "runtimes", "polyglot", "edge", "performance"],
        "estimated_value": 100000,
        "priority": "critical",
        "seeded_by": "functionfly-team",
        "business_impact": "Differentiates from AWS Lambda, Cloud Functions, Vercel, Netlify",
        "competitive_edge": "Only platform with universal WASM compilation pipeline",
        "time_to_market": "6-9 months"
    }' | jq .

echo ""
echo "Done! Check dashboard at: $API_URL/admin/unfair-advantage/dashboard"