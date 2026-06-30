package plans

// Feature categories
const (
	CategoryCore      = "core"
	CategorySecurity  = "security"
	CategoryAnalytics = "analytics"
	CategorySupport   = "support"
)

// Feature types
type FeatureType string

const (
	FeatureTypeBoolean FeatureType = "boolean"
	FeatureTypeNumeric FeatureType = "numeric"
	FeatureTypeList    FeatureType = "list"
)

// Feature represents a feature with its metadata
type Feature struct {
	Key         string      `json:"key"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	Type        FeatureType `json:"type"`
	Default     interface{} `json:"default"`
}

// BYOK (Bring Your Own Key) feature
const FeatureBYOK = "byok"

// Enterprise-only features
const (
	FeatureMicroVMs           = "micro_vms"
	FeatureMicroVMsEnterprise = "micro_vms_enterprise" // Enhanced MicroVM capabilities
	FeatureDedicatedPool      = "dedicated_pool"
	FeatureCustomLimits       = "custom_limits"
	FeatureAdvancedSecurity   = "advanced_security"
	FeatureSLA                = "sla"
	FeatureSLAPremium         = "sla_premium"        // Enterprise SLA: enhanced SLA with 99.999% uptime
	FeatureSLACredits         = "sla_credits"        // SLA credits for downtime violations
	FeaturePrioritySupport    = "priority_support"
	FeatureCustomDomains      = "custom_domains"
	FeatureSSOSAML            = "sso_saml"
	FeatureSCIM               = "scim"
	FeatureAuditLogs          = "audit_logs"
	FeatureDataResidency      = "data_residency"
	FeatureAPIRateLimits      = "api_rate_limits"
	FeatureWebhookSigning     = "webhook_signing"
	FeatureAdvancedAnalytics  = "advanced_analytics"
	FeatureTeamRBAC           = "team_rbac"
	FeatureSecretRotation     = "secret_rotation"
	// Playground features
	FeatureCollaborativeSessions = "collaborative_sessions"
)

// Function Consciousness features
const (
	FeatureConsciousnessBasic      = "consciousness_basic"      // Pro+: system awareness score, daily digest, basic insights
	FeatureConsciousnessAdvanced   = "consciousness_advanced"   // Enterprise+: real-time, predictive, auto-fix proposals
	FeatureConsciousnessAutonomous = "consciousness_autonomous" // Agent Enterprise: fully autonomous actions
)

// Time Machine features
const (
	FeatureTimeMachineBasic      = "time_machine_basic"      // Free+: 24h window, basic diff
	FeatureTimeMachineExtended   = "time_machine_extended"   // Starter+: 72h window
	FeatureTimeMachinePro        = "time_machine_pro"        // Pro+: 30d window, full diffs, dry-run reconciliation
	FeatureTimeMachineEnterprise = "time_machine_enterprise" // Enterprise+: 90d window, live reconciliation, audit certs
	FeatureTimeMachineUnlimited  = "time_machine_unlimited"  // Agent Enterprise: unlimited everything
	FeatureTimeMachineInsurance  = "time_machine_insurance"  // Agent Enterprise: dedicated incident engineer
)

// State Fabric Add-on features (premium stackable add-ons)
const (
	FeatureSFHotCache           = "sf_hot_cache"           // Hot cache tier for faster replay
	FeatureSFMultiRegion        = "sf_multi_region"        // Multi-region replication
	FeatureSFAIRecall           = "sf_ai_recall"           // Vector embeddings + AI memory
	FeatureSFAdvancedInsights   = "sf_advanced_insights"   // Cost forecasting, anomaly detection
	FeatureSFAdvancedSecurity   = "sf_advanced_security"   // SOC2 logs, key rotation, audit streams
)

// Pro-only features
const (
	FeatureExtendedProviders = "extended_providers"
	FeatureHigherRequests    = "higher_requests"
	FeatureAgentScaleTier    = "agent_scale_tier"
	FeatureBasicAnalytics    = "basic_analytics"
	FeatureWebhookRetries    = "webhook_retries"
	FeatureCustomHeaders     = "custom_headers"
	FeatureLongTimeout       = "long_timeout"
	FeatureBulkOperations    = "bulk_operations"
	FeaturePremiumSupport    = "premium_support"
)

// Trust & Attestation features
const (
	FeatureAttestationRead       = "attestation_read"       // Free+: view attestations and verify chains
	FeatureAttestationCreate     = "attestation_create"     // Pro+: create attestations
	FeatureAttestationRevoke     = "attestation_revoke"     // Enterprise+: revoke attestations
	FeatureAttestationManage     = "attestation_manage"     // Pro+: manage attestation lifecycle (create + view)
	FeatureTrustVerification     = "trust_verification"     // Starter+: submit verification requests
	FeatureTrustReports          = "trust_reports"          // Starter+: submit trust reports
	FeatureTrustPolicyEvaluate   = "trust_policy_evaluate"  // Pro+: evaluate functions against trust policies
	FeatureTrustPolicyManage     = "trust_policy_manage"    // Pro+: create/update/delete trust policies
)

// Starter features (included by default)
const (
	FeatureBasicProviders   = "basic_providers"
	FeatureBaseRequests     = "base_requests"
	FeatureAgentStarter     = "agent_starter"
	FeatureAgents           = "agents" // Bundled agent capability (replaces agent_starter as standalone)
	FeatureCommunitySupport = "community_support"
	FeatureBasicLogging     = "basic_logging"
	FeatureStandardSLA      = "standard_sla"
	FeaturePublishFunctions = "publish_functions"
)

// Feature definitions - all available features in the system
var featureDefinitions = map[string]Feature{
	// Enterprise features
	FeatureMicroVMs: {
		Key:         FeatureMicroVMs,
		Name:        "MicroVMs",
		Description: "Python MicroVM runtime (Firecracker)",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureDedicatedPool: {
		Key:         FeatureDedicatedPool,
		Name:        "Dedicated Pool",
		Description: "Unlimited agent concurrency",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureCustomLimits: {
		Key:         FeatureCustomLimits,
		Name:        "Custom Limits",
		Description: "Configurable request/limits",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAdvancedSecurity: {
		Key:         FeatureAdvancedSecurity,
		Name:        "Advanced Security",
		Description: "Enhanced security middleware",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSLA: {
		Key:         FeatureSLA,
		Name:        "SLA",
		Description: "Service Level Agreement",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSLAPremium: {
		Key:         FeatureSLAPremium,
		Name:        "Premium SLA",
		Description: "Enhanced SLA with 99.999% uptime guarantee, 15-min response time, and priority incident handling",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSLACredits: {
		Key:         FeatureSLACredits,
		Name:        "SLA Credits",
		Description: "Automatic credits for SLA violations (up to 50% of monthly fee)",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeaturePrioritySupport: {
		Key:         FeaturePrioritySupport,
		Name:        "Priority Support",
		Description: "24/7 priority support",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureCustomDomains: {
		Key:         FeatureCustomDomains,
		Name:        "Custom Domains",
		Description: "Branded custom domains",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSSOSAML: {
		Key:         FeatureSSOSAML,
		Name:        "SSO/SAML",
		Description: "Single Sign-On integration",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSCIM: {
		Key:         FeatureSCIM,
		Name:        "SCIM Provisioning",
		Description: "System for Cross-domain Identity Management",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAuditLogs: {
		Key:         FeatureAuditLogs,
		Name:        "Audit Logs",
		Description: "Extended audit logging",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureDataResidency: {
		Key:         FeatureDataResidency,
		Name:        "Data Residency",
		Description: "Region-specific data storage",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAPIRateLimits: {
		Key:         FeatureAPIRateLimits,
		Name:        "API Rate Limits",
		Description: "Custom rate limiting",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureWebhookSigning: {
		Key:         FeatureWebhookSigning,
		Name:        "Webhook Signing",
		Description: "Webhook signature verification",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAdvancedAnalytics: {
		Key:         FeatureAdvancedAnalytics,
		Name:        "Advanced Analytics",
		Description: "Extended analytics",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTeamRBAC: {
		Key:         FeatureTeamRBAC,
		Name:        "Team RBAC",
		Description: "Role-based access control",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSecretRotation: {
		Key:         FeatureSecretRotation,
		Name:        "Secret Rotation",
		Description: "Automatic secret rotation",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureCollaborativeSessions: {
		Key:         FeatureCollaborativeSessions,
		Name:        "Collaborative Playground Sessions",
		Description: "Real-time collaborative editing in function playground",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},

	// Function Consciousness features
	FeatureConsciousnessBasic: {
		Key:         FeatureConsciousnessBasic,
		Name:        "Function Consciousness",
		Description: "System awareness score, daily insight digest, basic cost and health insights",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureConsciousnessAdvanced: {
		Key:         FeatureConsciousnessAdvanced,
		Name:        "Advanced Consciousness",
		Description: "Real-time insights, predictive alerts, marketplace recommendations, auto-fix proposals",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureConsciousnessAutonomous: {
		Key:         FeatureConsciousnessAutonomous,
		Name:        "Autonomous Consciousness",
		Description: "Autonomous fix deployment, unlimited lookback, priority insight queue",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},

	// Trust & Attestation features
	FeatureAttestationRead: {
		Key:         FeatureAttestationRead,
		Name:        "Attestation Read",
		Description: "View function attestations and verify cryptographic chains",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     true, // Available to all plans
	},
	FeatureAttestationCreate: {
		Key:         FeatureAttestationCreate,
		Name:        "Attestation Create",
		Description: "Create new cryptographic attestations for function properties",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAttestationRevoke: {
		Key:         FeatureAttestationRevoke,
		Name:        "Attestation Revoke",
		Description: "Revoke existing attestations (admin/security action)",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAttestationManage: {
		Key:         FeatureAttestationManage,
		Name:        "Attestation Management",
		Description: "Full attestation lifecycle management including creation and policy evaluation",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTrustVerification: {
		Key:         FeatureTrustVerification,
		Name:        "Trust Verification",
		Description: "Submit function verification requests to establish trust",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTrustReports: {
		Key:         FeatureTrustReports,
		Name:        "Trust Reports",
		Description: "Submit trust issue reports for functions",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTrustPolicyEvaluate: {
		Key:         FeatureTrustPolicyEvaluate,
		Name:        "Trust Policy Evaluation",
		Description: "Evaluate functions against trust policies",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTrustPolicyManage: {
		Key:         FeatureTrustPolicyManage,
		Name:        "Trust Policy Management",
		Description: "Create, update, and delete trust policies",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},

	// Pro features
	FeatureExtendedProviders: {
		Key:         FeatureExtendedProviders,
		Name:        "Extended Providers",
		Description: "3 providers per app (vs 2)",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureHigherRequests: {
		Key:         FeatureHigherRequests,
		Name:        "Higher Requests",
		Description: "2.5M requests/month (vs 250K)",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureAgentScaleTier: {
		Key:         FeatureAgentScaleTier,
		Name:        "Agent Scale Tier",
		Description: "Access to agent_scale plan",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureBasicAnalytics: {
		Key:         FeatureBasicAnalytics,
		Name:        "Basic Analytics",
		Description: "Usage analytics dashboard",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureWebhookRetries: {
		Key:         FeatureWebhookRetries,
		Name:        "Webhook Retries",
		Description: "Automatic webhook retries",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureCustomHeaders: {
		Key:         FeatureCustomHeaders,
		Name:        "Custom Headers",
		Description: "Custom HTTP headers",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureLongTimeout: {
		Key:         FeatureLongTimeout,
		Name:        "Long Timeout",
		Description: "Extended function timeout",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureBulkOperations: {
		Key:         FeatureBulkOperations,
		Name:        "Bulk Operations",
		Description: "Bulk function operations",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},

	// Starter features
	FeatureBasicProviders: {
		Key:         FeatureBasicProviders,
		Name:        "Basic Providers",
		Description: "2 providers per app",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureBaseRequests: {
		Key:         FeatureBaseRequests,
		Name:        "Base Requests",
		Description: "100K requests/month",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureAgentStarter: {
		Key:         FeatureAgentStarter,
		Name:        "Agent Starter",
		Description: "Basic agent access (100K calls/mo, 10 concurrency)",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureAgents: {
		Key:         FeatureAgents,
		Name:        "AI Agents",
		Description: "Bundled agent capability (limits vary by plan)",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureCommunitySupport: {
		Key:         FeatureCommunitySupport,
		Name:        "Community Support",
		Description: "Community support",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureBasicLogging: {
		Key:         FeatureBasicLogging,
		Name:        "Basic Logging",
		Description: "7-day log retention",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureStandardSLA: {
		Key:         FeatureStandardSLA,
		Name:        "Standard SLA",
		Description: "Basic SLA (99.5%)",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeaturePremiumSupport: {
		Key:         FeaturePremiumSupport,
		Name:        "Premium Support",
		Description: "24/7 priority email and chat support with 4-hour response time",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeaturePublishFunctions: {
		Key:         FeaturePublishFunctions,
		Name:        "Publish Functions",
		Description: "Publish functions to registry",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureBYOK: {
		Key:         FeatureBYOK,
		Name:        "Bring Your Own Key",
		Description: "Connect your own AI provider API keys to avoid platform AI costs",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},

	// Time Machine features
	FeatureTimeMachineBasic: {
		Key:         FeatureTimeMachineBasic,
		Name:        "Time Machine Basic",
		Description: "24-hour replay window, basic text diff reports",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     true,
	},
	FeatureTimeMachineExtended: {
		Key:         FeatureTimeMachineExtended,
		Name:        "Time Machine Extended",
		Description: "72-hour replay window, up to 1,000 executions per replay",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTimeMachinePro: {
		Key:         FeatureTimeMachinePro,
		Name:        "Time Machine Pro",
		Description: "30-day replay window, full structured diff reports, dry-run reconciliation",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTimeMachineEnterprise: {
		Key:         FeatureTimeMachineEnterprise,
		Name:        "Time Machine Enterprise",
		Description: "90-day replay window, live reconciliation, SOC2/HIPAA audit certificates",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTimeMachineUnlimited: {
		Key:         FeatureTimeMachineUnlimited,
		Name:        "Time Machine Unlimited",
		Description: "Unlimited replay history, custom reconciliation rules, legal-grade audit certificates",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureTimeMachineInsurance: {
		Key:         FeatureTimeMachineInsurance,
		Name:        "Incident Insurance",
		Description: "Dedicated engineer support during critical production incidents",
		Category:    CategorySupport,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},

	// State Fabric Add-on features
	FeatureSFHotCache: {
		Key:         FeatureSFHotCache,
		Name:        "SF Hot Cache",
		Description: "Hot cache tier for faster replay and reduced read costs",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSFMultiRegion: {
		Key:         FeatureSFMultiRegion,
		Name:        "SF Multi-Region",
		Description: "Active-active multi-region replication for HA and global latency",
		Category:    CategoryCore,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSFAIRecall: {
		Key:         FeatureSFAIRecall,
		Name:        "SF AI Memory",
		Description: "Vector embeddings, AI memory storage, and fast recall engine",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSFAdvancedInsights: {
		Key:         FeatureSFAdvancedInsights,
		Name:        "SF Advanced Insights",
		Description: "Cost forecasting, anomaly detection, hot path alerts",
		Category:    CategoryAnalytics,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
	FeatureSFAdvancedSecurity: {
		Key:         FeatureSFAdvancedSecurity,
		Name:        "SF Advanced Security",
		Description: "SOC2-friendly logs, key rotation, audit streams",
		Category:    CategorySecurity,
		Type:        FeatureTypeBoolean,
		Default:     false,
	},
}

// Feature sets per plan - 2026 unified structure
// Agents are now BUNDLED into main plans, no separate agent plan hierarchy
var (
	enterpriseFeatures = []string{
		FeatureMicroVMs,
		FeatureDedicatedPool,
		FeatureCustomLimits,
		FeatureAdvancedSecurity,
		FeatureSLA,
		FeaturePrioritySupport,
		FeaturePremiumSupport,
		FeatureCustomDomains,
		FeatureSSOSAML,
		FeatureAuditLogs,
		FeatureDataResidency,
		FeatureAPIRateLimits,
		FeatureWebhookSigning,
		FeatureAdvancedAnalytics,
		FeatureTeamRBAC,
		FeatureSecretRotation,
		FeatureExtendedProviders,
		FeatureHigherRequests,
		FeatureAgentScaleTier,
		FeatureBasicAnalytics,
		FeatureWebhookRetries,
		FeatureCustomHeaders,
		FeatureLongTimeout,
		FeatureBulkOperations,
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureAgentStarter,
		FeatureCommunitySupport,
		FeatureBasicLogging,
		FeatureStandardSLA,
		FeaturePublishFunctions,
		FeatureAgents, // Bundled agent capability
		FeatureTimeMachineBasic,
		FeatureTimeMachineExtended,
		FeatureTimeMachinePro,
		FeatureTimeMachineEnterprise,
		FeatureConsciousnessBasic,
		FeatureConsciousnessAdvanced,
		FeatureCollaborativeSessions,
		// State Fabric add-ons (all available as add-ons)
		FeatureSFHotCache,
		FeatureSFMultiRegion,
		FeatureSFAIRecall,
		FeatureSFAdvancedInsights,
		FeatureSFAdvancedSecurity,
		// Trust & Attestation (full access)
		FeatureAttestationRead,
		FeatureAttestationCreate,
		FeatureAttestationRevoke,
		FeatureAttestationManage,
		FeatureTrustVerification,
		FeatureTrustReports,
		FeatureTrustPolicyEvaluate,
		FeatureTrustPolicyManage,
		FeatureBYOK,
	}

	// Enterprise SLA features - enhanced SLA with 99.999% uptime guarantee
	enterpriseSLAFeatures = []string{
		FeatureMicroVMs,
		FeatureDedicatedPool,
		FeatureCustomLimits,
		FeatureAdvancedSecurity,
		FeatureSLA,
		FeatureSLAPremium,     // Enhanced SLA with 99.999% uptime
		FeatureSLACredits,     // SLA credits for violations
		FeaturePrioritySupport,
		FeaturePremiumSupport,
		FeatureCustomDomains,
		FeatureSSOSAML,
		FeatureAuditLogs,
		FeatureDataResidency,
		FeatureAPIRateLimits,
		FeatureWebhookSigning,
		FeatureAdvancedAnalytics,
		FeatureTeamRBAC,
		FeatureSecretRotation,
		FeatureExtendedProviders,
		FeatureHigherRequests,
		FeatureAgentScaleTier,
		FeatureBasicAnalytics,
		FeatureWebhookRetries,
		FeatureCustomHeaders,
		FeatureLongTimeout,
		FeatureBulkOperations,
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureAgentStarter,
		FeatureCommunitySupport,
		FeatureBasicLogging,
		FeatureStandardSLA,
		FeaturePublishFunctions,
		FeatureAgents,
		FeatureTimeMachineBasic,
		FeatureTimeMachineExtended,
		FeatureTimeMachinePro,
		FeatureTimeMachineEnterprise,
		FeatureConsciousnessBasic,
		FeatureConsciousnessAdvanced,
		FeatureCollaborativeSessions,
		// All State Fabric add-ons included
		FeatureSFHotCache,
		FeatureSFMultiRegion,
		FeatureSFAIRecall,
		FeatureSFAdvancedInsights,
		FeatureSFAdvancedSecurity,
		// Trust & Attestation (full access)
		FeatureAttestationRead,
		FeatureAttestationCreate,
		FeatureAttestationRevoke,
		FeatureAttestationManage,
		FeatureTrustVerification,
		FeatureTrustReports,
		FeatureTrustPolicyEvaluate,
		FeatureTrustPolicyManage,
	}

	proFeatures = []string{
		FeatureExtendedProviders,
		FeatureHigherRequests,
		FeatureAgentScaleTier,
		FeatureBasicAnalytics,
		FeatureWebhookRetries,
		FeatureCustomHeaders,
		FeatureLongTimeout,
		FeatureBulkOperations,
		FeaturePremiumSupport,
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureAgentStarter,
		FeatureCommunitySupport,
		FeatureBasicLogging,
		FeatureStandardSLA,
		FeaturePublishFunctions,
		FeatureAgents, // Bundled agent capability (Agent Scale level)
		FeatureTimeMachineBasic,
		FeatureTimeMachineExtended,
		FeatureTimeMachinePro,
		FeatureConsciousnessBasic,
		// State Fabric add-ons (Hot Cache available as add-on)
		FeatureSFHotCache,
		// Trust & Attestation
		FeatureAttestationRead,
		FeatureAttestationCreate,
		FeatureAttestationManage,
		FeatureTrustVerification,
		FeatureTrustReports,
		FeatureTrustPolicyEvaluate,
		FeatureTrustPolicyManage,
		FeatureBYOK,
	}

	starterFeatures = []string{
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureAgentStarter, // Bundled agent capability (Agent Starter level)
		FeatureCommunitySupport,
		FeatureBasicLogging,
		FeatureStandardSLA,
		FeaturePublishFunctions,
		FeatureTimeMachineBasic,
		FeatureTimeMachineExtended,
		FeatureConsciousnessBasic,
		FeatureAttestationRead,
		FeatureTrustVerification,
		FeatureTrustReports,
		FeatureBYOK,
	}

	freeFeatures = []string{
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureCommunitySupport,
		FeatureBasicLogging,
		FeatureStandardSLA,
		FeaturePublishFunctions,
		FeatureTimeMachineBasic,
		FeatureAttestationRead,
		FeatureBYOK,
	}

	agentEnterpriseFeatures = []string{
		FeatureDedicatedPool,
		FeatureAdvancedSecurity,
		FeaturePrioritySupport,
		FeaturePremiumSupport,
		FeatureAuditLogs,
		FeatureTeamRBAC,
		FeatureSecretRotation,
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureTimeMachineBasic,
		FeatureTimeMachineExtended,
		FeatureTimeMachinePro,
		FeatureTimeMachineEnterprise,
		FeatureTimeMachineUnlimited,
		FeatureTimeMachineInsurance,
		FeatureConsciousnessBasic,
		FeatureConsciousnessAdvanced,
		FeatureConsciousnessAutonomous,
		FeatureBYOK,
	}

	agentProFeatures = []string{
		FeatureBasicAnalytics,
		FeatureCustomHeaders,
		FeatureBulkOperations,
		FeatureBasicProviders,
		FeatureBaseRequests,
	}

	agentScaleFeatures = []string{
		FeatureBasicProviders,
		FeatureBaseRequests,
	}

	agentStarterFeatures = []string{
		FeatureBasicProviders,
		FeatureBaseRequests,
	}

	// MicroVM Enterprise features - enhanced MicroVM capabilities
	microVMEnterpriseFeatures = []string{
		FeatureMicroVMs,
		FeatureMicroVMsEnterprise, // Enhanced MicroVM capabilities (500 VMs, 8GB memory, 16 vCPU)
		FeatureDedicatedPool,
		FeatureCustomLimits,
		FeatureAdvancedSecurity,
		FeatureSLA,
		FeaturePrioritySupport,
		FeaturePremiumSupport,
		FeatureCustomDomains,
		FeatureSSOSAML,
		FeatureSCIM,
		FeatureAuditLogs,
		FeatureDataResidency,
		FeatureAPIRateLimits,
		FeatureWebhookSigning,
		FeatureAdvancedAnalytics,
		FeatureTeamRBAC,
		FeatureSecretRotation,
		FeatureExtendedProviders,
		FeatureHigherRequests,
		FeatureAgentScaleTier,
		FeatureBasicAnalytics,
		FeatureWebhookRetries,
		FeatureCustomHeaders,
		FeatureLongTimeout,
		FeatureBulkOperations,
		FeatureBasicProviders,
		FeatureBaseRequests,
		FeatureAgentStarter,
		FeatureCommunitySupport,
		FeatureBasicLogging,
		FeatureStandardSLA,
		FeaturePublishFunctions,
		FeatureAgents,
		FeatureTimeMachineBasic,
		FeatureTimeMachineExtended,
		FeatureTimeMachinePro,
		FeatureTimeMachineEnterprise,
		FeatureConsciousnessBasic,
		FeatureConsciousnessAdvanced,
		FeatureCollaborativeSessions,
		// State Fabric add-ons (all available)
		FeatureSFHotCache,
		FeatureSFMultiRegion,
		FeatureSFAIRecall,
		FeatureSFAdvancedInsights,
		FeatureSFAdvancedSecurity,
	}
)

// GetAllFeatures returns all feature definitions
func GetAllFeatures() []Feature {
	features := make([]Feature, 0, len(featureDefinitions))
	for _, f := range featureDefinitions {
		features = append(features, f)
	}
	return features
}

// GetFeatureDefinition returns a feature definition by key
func GetFeatureDefinition(key string) (Feature, bool) {
	f, ok := featureDefinitions[key]
	return f, ok
}

// GetFeaturesForPlan returns all features available for a plan
// Supports both main plans (free/starter/pro/enterprise) and legacy agent tiers
func GetFeaturesForPlan(plan string) []string {
	switch plan {
	case PlanFree:
		return freeFeatures
	case PlanEnterpriseSLA:
		return enterpriseSLAFeatures
	case PlanEnterprise, PlanAgentPro:
		return enterpriseFeatures
	case PlanMicroVMEnterprise:
		return microVMEnterpriseFeatures
	case PlanPro, PlanAgentScale:
		return proFeatures
	case PlanStarter, PlanAgentStarter:
		return starterFeatures
	case PlanAgentEnterprise:
		return agentEnterpriseFeatures
	default:
		return starterFeatures
	}
}

// HasFeature checks if a plan has a specific feature
func HasFeature(plan string, feature string) bool {
	features := GetFeaturesForPlan(plan)
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}

// IsEnterpriseOnly checks if a feature is only available on enterprise
func IsEnterpriseOnly(feature string) bool {
	for _, f := range enterpriseFeatures {
		if f == feature {
			return true
		}
	}
	// Check if it's NOT in pro features - if so, it's enterprise only
	for _, f := range proFeatures {
		if f == feature {
			return false
		}
	}
	return true
}

// IsProOnly checks if a feature is only available on pro and enterprise
func IsProOnly(feature string) bool {
	// It's pro only if it's in proFeatures but not in starterFeatures
	for _, f := range proFeatures {
		if f == feature {
			// Now check if it's in starter features
			for _, sf := range starterFeatures {
				if sf == feature {
					return false // It's available in starter too
				}
			}
			return true
		}
	}
	return false
}

// GetFeaturesByCategory returns features filtered by category
func GetFeaturesByCategory(category string) []Feature {
	var features []Feature
	for _, f := range featureDefinitions {
		if f.Category == category {
			features = append(features, f)
		}
	}
	return features
}

// PlanInfo represents plan information with features
type PlanInfo struct {
	Plan     string   `json:"plan"`
	Features []string `json:"features"`
}

// GetAllPlanInfo returns information about all plans
func GetAllPlanInfo() []PlanInfo {
	return []PlanInfo{
		{Plan: PlanFree, Features: freeFeatures},
		{Plan: PlanStarter, Features: starterFeatures},
		{Plan: PlanPro, Features: proFeatures},
		{Plan: PlanEnterprise, Features: enterpriseFeatures},
		{Plan: PlanEnterpriseSLA, Features: enterpriseSLAFeatures},
		{Plan: PlanMicroVMEnterprise, Features: microVMEnterpriseFeatures},
		{Plan: PlanAgentStarter, Features: agentStarterFeatures},
		{Plan: PlanAgentScale, Features: agentScaleFeatures},
		{Plan: PlanAgentPro, Features: agentProFeatures},
		{Plan: PlanAgentEnterprise, Features: agentEnterpriseFeatures},
	}
}
