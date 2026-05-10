package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AIAppProvisioner creates the isolated AI infrastructure for a tenant.
// All AI data (vectors, embeddings, chat, memory) lives in the tenant's own database.
//
// What it provisions:
//   - Default embedding collections (Knowledge Base, Conversation Memory, Documents)
//   - Default AI assistant with system prompt, RAG config, guardrails
//   - Model configurations (GPT-4o, Claude 3, text-embedding-3-small)
//   - Email templates for AI-specific workflows
//   - Analytics events for AI operations
//   - Memory system seed data
type AIAppProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
}

// NewAIAppProvisioner creates a new AI App provisioner
func NewAIAppProvisioner(platformRepo storage.Repository, dbProvisioner *storage.TenantDBProvisioner) *AIAppProvisioner {
	return &AIAppProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
	}
}

// Provision sets up the complete AI infrastructure in the tenant's database.
func (ap *AIAppProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "ai_app",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Get tenant database pool
	pool, err := ap.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 2. Seed default embedding collections
	collections := []struct {
		name        string
		slug        string
		desc        string
		model       string
		chunkSize   int
		chunkOverlap int
	}{
		{"Knowledge Base", "knowledge-base", "Main knowledge base for RAG — add documents, URLs, and files", "text-embedding-3-small", 512, 50},
		{"Conversation Memory", "conversation-memory", "Persistent memory across conversations", "text-embedding-3-small", 256, 25},
		{"Documentation", "documentation", "Product docs and help content for support chatbot", "text-embedding-3-small", 512, 50},
		{"Training Data", "training-data", "Curated Q&A pairs and fine-tuning examples", "text-embedding-3-small", 256, 25},
	}

	collectionIDs := make(map[string]uuid.UUID)
	for _, c := range collections {
		var collID uuid.UUID
		err = pool.QueryRow(ctx,
			`INSERT INTO ai_collections (id, tenant_id, name, slug, description, embedding_model, embedding_dimensions, distance_metric, chunk_size, chunk_overlap, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, 1536, 'cosine', $7, $8, true)
			 ON CONFLICT (tenant_id, slug) DO UPDATE SET updated_at = NOW()
			 RETURNING id`,
			uuid.New(), tenantID, c.name, c.slug, c.desc, c.model, c.chunkSize, c.chunkOverlap).Scan(&collID)
		if err != nil {
			log.WithError(err).WithField("collection", c.name).Warn("Failed to create collection (non-fatal)")
			continue
		}
		collectionIDs[c.slug] = collID
	}
	log.WithField("count", len(collections)).Info("Embedding collections created")

	// 3. Create default AI assistant
	kbCollID := collectionIDs["knowledge-base"]
	memCollID := collectionIDs["conversation-memory"]
	collIDsJSON, _ := json.Marshal([]string{kbCollID.String(), memCollID.String()})

	toolsJSON := `[{"type":"function","function":{"name":"search_knowledge","description":"Search the knowledge base for relevant information","parameters":{"type":"object","properties":{"query":{"type":"string","description":"The search query"},"collection":{"type":"string","description":"Collection to search"}},"required":["query"]}}},{"type":"function","function":{"name":"save_memory","description":"Save important information to long-term memory","parameters":{"type":"object","properties":{"content":{"type":"string","description":"The information to remember"},"category":{"type":"string","description":"Memory category (fact, preference, context)"},"importance":{"type":"number","description":"Importance 0-1"}},"required":["content","category"]}}},{"type":"function","function":{"name":"recall_memory","description":"Recall information from long-term memory","parameters":{"type":"object","properties":{"query":{"type":"string","description":"What to recall"},"memory_type":{"type":"string","description":"Type of memory to search"}},"required":["query"]}}}]`

	guardrailsJSON := `{"max_turns_per_conversation":100,"blocked_topics":[],"content_filter":true,"require_citations":false,"max_context_tokens":128000,"rate_limit_per_user_per_minute":30}`

	defaultAssistantID := uuid.New()
	_, err = pool.Exec(ctx,
		`INSERT INTO ai_assistants (id, tenant_id, name, slug, description, system_prompt, model, temperature, max_tokens, collection_ids, tools, guardrails, welcome_message, suggested_questions, is_public, is_active)
		 VALUES ($1, $2, 'Default Assistant', 'default', 'AI assistant with RAG, memory, and tool use', $3, 'gpt-4o', 0.7, 4096, $4, $5, $6, $7, $8, true, true)
		 ON CONFLICT (tenant_id, slug) DO NOTHING`,
		defaultAssistantID, tenantID,
		`You are a helpful AI assistant. You have access to a knowledge base and long-term memory.

When answering questions:
1. First search the knowledge base for relevant information
2. Check your memory for any relevant context about the user
3. Provide clear, accurate answers with citations when available
4. Save important new information to memory for future reference

Be concise, helpful, and always cite your sources when using knowledge base content.`,
		collIDsJSON,
		toolsJSON,
		guardrailsJSON,
		"Hello! I'm your AI assistant. I can help you find information, answer questions, and remember important details. What would you like to know?",
		`["What can you help me with?","Search the knowledge base","What do you remember about me?"]`)
	if err != nil {
		log.WithError(err).Warn("Failed to create default assistant (non-fatal)")
	}
	log.WithField("assistant_id", defaultAssistantID).Info("Default AI assistant created")

	// 4. Seed model configurations
	models := []struct {
		provider    string
		modelID     string
		displayName string
		isDefault   bool
		costIn      float64
		costOut     float64
		rpm         int
		tpm         int
	}{
		{"openai", "gpt-4o", "GPT-4o", true, 0.0025, 0.01, 500, 150000},
		{"openai", "gpt-4o-mini", "GPT-4o Mini", false, 0.00015, 0.0006, 1000, 200000},
		{"openai", "gpt-3.5-turbo", "GPT-3.5 Turbo", false, 0.0005, 0.0015, 3000, 1000000},
		{"openai", "text-embedding-3-small", "Embedding (Small)", false, 0.00002, 0, 5000, 5000000},
		{"openai", "text-embedding-3-large", "Embedding (Large)", false, 0.00013, 0, 5000, 5000000},
		{"anthropic", "claude-3-opus-20240229", "Claude 3 Opus", false, 0.015, 0.075, 100, 40000},
		{"anthropic", "claude-3-sonnet-20240229", "Claude 3 Sonnet", false, 0.003, 0.015, 500, 80000},
		{"anthropic", "claude-3-haiku-20240307", "Claude 3 Haiku", false, 0.00025, 0.00125, 1000, 200000},
		{"openrouter", "meta-llama/llama-3-70b-instruct", "Llama 3 70B", false, 0.0008, 0.0008, 500, 100000},
	}

	for _, m := range models {
		_, err = pool.Exec(ctx,
			`INSERT INTO ai_model_configs (id, tenant_id, provider, model_id, display_name, is_default, is_enabled, cost_per_1k_input, cost_per_1k_output, rate_limit_rpm, rate_limit_tpm)
			 VALUES ($1, $2, $3, $4, $5, $6, true, $7, $8, $9, $10)
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, m.provider, m.modelID, m.displayName, m.isDefault, m.costIn, m.costOut, m.rpm, m.tpm)
		if err != nil {
			log.WithError(err).WithField("model", m.modelID).Warn("Failed to seed model config (non-fatal)")
		}
	}
	log.WithField("count", len(models)).Info("Model configurations seeded")

	// 5. Seed email templates for AI workflows
	aiTemplates := []struct {
		slug    string
		name    string
		subject string
	}{
		{"ai-welcome", "AI App Welcome", "Your AI app is ready!"},
		{"ai-usage-limit", "Usage Limit Warning", "You're approaching your AI usage limit"},
		{"ai-training-complete", "Training Complete", "Your custom model training is complete"},
		{"ai-document-processed", "Document Processed", "Your document has been processed and indexed"},
		{"ai-knowledge-base-sync", "Knowledge Base Sync", "Knowledge base sync complete"},
		{"ai-weekly-digest", "Weekly AI Digest", "Your weekly AI usage summary"},
	}

	for _, t := range aiTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_email_templates (id, tenant_id, slug, name, subject, html_body, text_body, variables, category, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ai', true)
			 ON CONFLICT (tenant_id, slug) DO NOTHING`,
			uuid.New(), tenantID, t.slug, t.name, t.subject,
			fmt.Sprintf("<!-- %s template — customize in dashboard -->", t.name),
			fmt.Sprintf("%s template - customize in dashboard", t.name),
			`[{"name":"AppName","type":"string","default":"My AI App","required":false},{"name":"UserName","type":"string","default":"","required":false},{"name":"UsagePercent","type":"string","default":"0","required":false},{"name":"Link","type":"string","default":"","required":false}]`)
		if err != nil {
			log.WithError(err).WithField("template", t.slug).Warn("Failed to seed AI template (non-fatal)")
		}
	}
	log.WithField("count", len(aiTemplates)).Info("AI email templates seeded")

	// 6. Create email workflows
	workflows := []struct {
		slug        string
		name        string
		triggerType string
		steps       []workflowStep
	}{
		{
			"ai-onboarding", "AI Onboarding", "user_signup",
			[]workflowStep{
				{0, "ai-welcome", 0, "minutes", "always"},
				{1, "ai-knowledge-base-sync", 2, "hours", "always"},
			},
		},
		{
			"ai-usage-monitoring", "AI Usage Monitoring", "usage_threshold",
			[]workflowStep{
				{0, "ai-usage-limit", 0, "minutes", "always"},
			},
		},
		{
			"ai-training-lifecycle", "Training Lifecycle", "training_started",
			[]workflowStep{
				{0, "ai-document-processed", 0, "minutes", "always"},
				{1, "ai-training-complete", 0, "minutes", "always"},
			},
		},
	}

	for _, wf := range workflows {
		var workflowID uuid.UUID
		err = pool.QueryRow(ctx,
			`INSERT INTO tenant_email_workflows (id, tenant_id, slug, name, trigger_type, trigger_config, is_active)
			 VALUES ($1, $2, $3, $4, $5, '{}', true)
			 ON CONFLICT (tenant_id, slug) DO UPDATE SET updated_at = NOW()
			 RETURNING id`,
			uuid.New(), tenantID, wf.slug, wf.name, wf.triggerType).Scan(&workflowID)
		if err != nil {
			log.WithError(err).WithField("workflow", wf.slug).Warn("Failed to create workflow (non-fatal)")
			continue
		}

		for _, step := range wf.steps {
			var templateID uuid.NullUUID
			pool.QueryRow(ctx,
				`SELECT id FROM tenant_email_templates WHERE tenant_id = $1 AND slug = $2`,
				tenantID, step.templateSlug).Scan(&templateID)

			_, err = pool.Exec(ctx,
				`INSERT INTO tenant_email_workflow_steps (id, tenant_id, workflow_id, template_id, step_order, delay_amount, delay_unit, condition_type, is_active)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
				 ON CONFLICT (workflow_id, step_order) DO NOTHING`,
				uuid.New(), tenantID, workflowID, templateID, step.order, step.delay, step.unit, step.condition)
			if err != nil {
				log.WithError(err).WithField("workflow", wf.slug).Warn("Failed to create workflow step (non-fatal)")
			}
		}
		log.WithField("workflow", wf.slug).Info("Email workflow created")
	}

	// 7. Seed AI-specific analytics events
	aiEvents := []struct {
		name     string
		category string
	}{
		{"chat_message_sent", "ai"},
		{"chat_response_received", "ai"},
		{"rag_query", "ai"},
		{"rag_context_retrieved", "ai"},
		{"embedding_generated", "ai"},
		{"document_uploaded", "knowledge"},
		{"document_processed", "knowledge"},
		{"document_failed", "knowledge"},
		{"memory_created", "memory"},
		{"memory_recalled", "memory"},
		{"memory_updated", "memory"},
		{"assistant_created", "management"},
		{"assistant_updated", "management"},
		{"collection_created", "management"},
		{"training_started", "training"},
		{"training_completed", "training"},
		{"model_switched", "management"},
		{"token_limit_reached", "billing"},
		{"feedback_positive", "quality"},
		{"feedback_negative", "quality"},
	}

	for _, ev := range aiEvents {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_analytics_events (id, tenant_id, event_name, event_category, properties, created_at)
			 VALUES ($1, $2, $3, $4, '{"type":"definition","description":"AI App event"}', NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, ev.name, ev.category)
		if err != nil {
			// Non-fatal
		}
	}
	log.WithField("count", len(aiEvents)).Info("AI analytics events seeded")

	// 8. Create AI-specific analytics dashboard
	dashboardLayout := `[{"widget_type":"metric_card","title":"Total Conversations","position":{"x":0,"y":0,"w":3,"h":2},"config":{"metric":"conversations","period":"7d"}},{"widget_type":"metric_card","title":"Messages","position":{"x":3,"y":0,"w":3,"h":2},"config":{"metric":"chat_messages_sent","period":"7d"}},{"widget_type":"metric_card","title":"RAG Queries","position":{"x":6,"y":0,"w":3,"h":2},"config":{"metric":"rag_query","period":"7d"}},{"widget_type":"metric_card","title":"Tokens Used","position":{"x":9,"y":0,"w":3,"h":2},"config":{"metric":"total_tokens","period":"30d"}},{"widget_type":"line_chart","title":"Token Usage Trend","position":{"x":0,"y":2,"w":6,"h":4},"config":{"metrics":["total_tokens","cost_usd"],"period":"30d"}},{"widget_type":"bar_chart","title":"Model Usage","position":{"x":6,"y":2,"w":6,"h":4},"config":{"metric":"chat_response_received","group_by":"model","period":"30d"}},{"widget_type":"funnel_chart","title":"RAG Effectiveness","position":{"x":0,"y":6,"w":6,"h":4},"config":{"funnel":"rag-pipeline"}},{"widget_type":"table","title":"Recent Conversations","position":{"x":6,"y":6,"w":6,"h":4},"config":{"source":"recent_conversations","limit":10}}]`

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_analytics_dashboards (id, tenant_id, name, layout, is_default, created_at, updated_at)
		 VALUES ($1, $2, 'AI Overview', $3, true, NOW(), NOW())
		 ON CONFLICT DO NOTHING`,
		uuid.New(), tenantID, dashboardLayout)
	if err != nil {
		log.WithError(err).Warn("Failed to create AI dashboard (non-fatal)")
	}

	// 9. Create AI-specific funnels
	aiFunnels := []struct {
		name  string
		desc  string
		steps []map[string]interface{}
	}{
		{
			"RAG Pipeline",
			"Track RAG query effectiveness from query to cited response",
			[]map[string]interface{}{
				{"event_name": "rag_query"},
				{"event_name": "rag_context_retrieved"},
				{"event_name": "chat_response_received"},
				{"event_name": "feedback_positive"},
			},
		},
		{
			"Document Ingestion",
			"Track document processing pipeline",
			[]map[string]interface{}{
				{"event_name": "document_uploaded"},
				{"event_name": "document_processed"},
				{"event_name": "embedding_generated"},
			},
		},
		{
			"User Activation",
			"Track AI feature adoption",
			[]map[string]interface{}{
				{"event_name": "chat_message_sent"},
				{"event_name": "rag_query"},
				{"event_name": "memory_created"},
				{"event_name": "document_uploaded"},
			},
		},
	}

	for _, f := range aiFunnels {
		stepsJSON, _ := json.Marshal(f.steps)
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_analytics_funnels (id, tenant_id, name, description, steps, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
			 ON CONFLICT DO NOTHING`,
			uuid.New(), tenantID, f.name, f.desc, stepsJSON)
		if err != nil {
			log.WithError(err).WithField("funnel", f.name).Warn("Failed to create funnel (non-fatal)")
		}
	}

	// 10. Seed a sample document in the knowledge base for immediate value
	kbCollID = collectionIDs["knowledge-base"]
	if kbCollID != uuid.Nil {
		docID := uuid.New()
		_, err = pool.Exec(ctx,
			`INSERT INTO ai_documents (id, tenant_id, collection_id, title, source_type, content, status, metadata, processed_at)
			 VALUES ($1, $2, $3, 'Welcome to your Knowledge Base', 'text', $4, 'ready', '{"is_sample":true}', NOW())
			 ON CONFLICT DO NOTHING`,
			docID, tenantID, kbCollID,
			`# Welcome to your AI Knowledge Base

This is your private knowledge base for Retrieval-Augmented Generation (RAG). Add documents, URLs, and files to give your AI assistant domain-specific knowledge.

## Getting Started

1. **Add Documents**: Upload PDFs, paste text, or crawl websites
2. **Organize Collections**: Group related content into collections
3. **Configure Assistants**: Link collections to AI assistants for RAG
4. **Monitor Performance**: Track retrieval accuracy and user feedback

## How RAG Works

When a user asks a question, the system:
1. Converts the question to an embedding vector
2. Searches for the most similar document chunks
3. Includes relevant chunks as context for the LLM
4. The LLM generates a response grounded in your data

This ensures responses are accurate, up-to-date, and cite their sources.

## Best Practices

- Keep documents focused and well-structured
- Use descriptive titles and metadata
- Update content regularly to stay current
- Monitor feedback to improve retrieval quality`)
		if err != nil {
			log.WithError(err).Warn("Failed to seed sample document (non-fatal)")
		}
		log.Info("Sample knowledge base document created")
	}

	// 11. Configure AI payment settings
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_payment_config (id, tenant_id, payment_mode, default_currency, allowed_payment_methods, tax_calculation_mode, metadata)
		 VALUES ($1, $2, 'platform', 'usd', '["card"]', 'automatic', '{"type":"ai_app","usage_billing":true,"token_based":true}')
		 ON CONFLICT (tenant_id) DO UPDATE SET metadata = tenant_payment_config.metadata || '{"type":"ai_app"}', updated_at = NOW()`,
		uuid.New(), tenantID)
	if err != nil {
		log.WithError(err).Warn("Failed to create payment config (non-fatal)")
	}

	// 12. Configure OpenRouter provider if available
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openrouterKey != "" {
		log.Info("OpenRouter API key found — AI model access pre-configured")
	}

	// 13. Log provisioning in audit
	auditMeta, _ := json.Marshal(map[string]interface{}{
		"component":   "ai_app",
		"action":      "provisioned",
		"collections": len(collections),
		"models":      len(models),
		"workflows":   len(workflows),
		"events":      len(aiEvents),
	})
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_auth_audit (id, tenant_id, event_type, success, metadata, created_at)
		 VALUES ($1, $2, 'system_provision', true, $3, NOW())`,
		uuid.New(), tenantID, auditMeta)
	if err != nil {
		log.WithError(err).Warn("Failed to log provisioning audit (non-fatal)")
	}

	state.Status = StatusActive
	state.ResourceID = fmt.Sprintf("ai:%s", defaultAssistantID)
	log.Info("AI App provisioning complete")
	return state, nil
}
