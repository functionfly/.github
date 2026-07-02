import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle, Circle, Loader2, AlertCircle, Shield, CreditCard, Mail, BarChart3, Database, Store, MessageSquare, Bell, Brain, Cpu, Sparkles, MessageCircle } from 'lucide-react';
import { apiClient } from '@/api/client';
import {
  Chamber,
  CornerBrace,
  StatusPill,
  SealedButton,
  FrameButton,
  AnnotationTag,
} from '@/components/containment';
import '@/styles/sc-provisioning.css';

// ─── Types ───────────────────────────────────────────────────────────────────

interface ComponentState {
  status: 'pending' | 'provisioning' | 'active' | 'failed' | 'rolled_back';
  timestamp: string;
  error?: string;
  resource_id?: string;
}

interface ProvisioningResult {
  tenant_id: string;
  bundle_slug: string;
  status: 'pending' | 'provisioning' | 'active' | 'failed';
  components: Record<string, ComponentState>;
  started_at: string;
  finished_at: string;
  duration_ms: number;
  error_log?: string[];
}

interface ProvisioningStatusCardProps {
  bundleSlug: string;
  onComplete?: () => void;
}

// ─── Component definitions ───────────────────────────────────────────────────

const COMPONENT_META: Record<string, { icon: React.ComponentType<React.SVGAttributes<SVGElement>>; label: string; description: string }> = {
  user_db: { icon: Database, label: 'User Database', description: 'Dedicated PostgreSQL database with isolated schema' },
  auth: { icon: Shield, label: 'Authentication', description: 'JWT keys, OAuth configs, sessions, MFA' },
  payments: { icon: CreditCard, label: 'Payments', description: 'Stripe integration, products, invoices, webhooks' },
  email_workflows: { icon: Mail, label: 'Email Workflows', description: 'Templates, automated sequences, dunning flows' },
  analytics: { icon: BarChart3, label: 'Analytics', description: 'Event tracking, funnels, cohorts, dashboards' },
  marketplace: { icon: Store, label: 'Marketplace', description: 'Listings, categories, reviews, seller accounts' },
  messaging: { icon: MessageSquare, label: 'Messaging', description: 'Buyer-seller conversations, offers, attachments' },
  notifications: { icon: Bell, label: 'Notifications', description: 'In-app, email, push notification templates and delivery' },
  ai_app: { icon: Brain, label: 'AI Infrastructure', description: 'Vector DB, embeddings, RAG collections, memory system' },
  vector_db: { icon: Cpu, label: 'Vector Database', description: 'pgvector collections with HNSW indexing' },
  chat_workflows: { icon: MessageCircle, label: 'Chat Workflows', description: 'AI assistants, conversations, RAG pipelines, tool calling' },
  memory_system: { icon: Sparkles, label: 'Memory System', description: 'Long-term memory, user profiles, semantic recall' },
};

const BUNDLE_COMPONENTS: Record<string, string[]> = {
  'saas-starter': ['user_db', 'auth', 'payments', 'email_workflows', 'analytics'],
  'marketplace': ['user_db', 'auth', 'payments', 'email_workflows', 'analytics', 'marketplace'],
  'ai-app': ['user_db', 'auth', 'payments', 'email_workflows', 'analytics', 'ai_app'],
};

// ─── Animated status indicator ──────────────────────────────────────────────

function StatusIndicator({ status }: { status: string }) {
  return (
    <div className="sc-prov__status-indicator">
      <AnimatePresence mode="wait">
        {status === 'active' && (
          <motion.div
            key="active"
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 500, damping: 25 }}
            className="sc-prov__status-ring sc-prov__status-ring--ok"
          >
            <CheckCircle className="sc-prov__status-icon sc-prov__status-icon--ok" />
          </motion.div>
        )}

        {status === 'provisioning' && (
          <motion.div
            key="provisioning"
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0, opacity: 0 }}
            transition={{ duration: 0.15 }}
            className="sc-prov__status-ring sc-prov__status-ring--spin"
          >
            <Loader2 className="sc-prov__status-icon sc-prov__status-icon--spin" />
          </motion.div>
        )}

        {status === 'failed' && (
          <motion.div
            key="failed"
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            exit={{ scale: 0, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 500, damping: 25 }}
            className="sc-prov__status-ring sc-prov__status-ring--fail"
          >
            <AlertCircle className="sc-prov__status-icon sc-prov__status-icon--fail" />
          </motion.div>
        )}

        {(status === 'pending' || !status) && (
          <motion.div
            key="pending"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.15 }}
            className="sc-prov__status-ring"
          >
            <Circle className="sc-prov__status-icon sc-prov__status-icon--pending" />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

// ─── Progress strip ─────────────────────────────────────────────────────────

function ProgressStrip({ componentOrder, components }: { componentOrder: string[]; components: Record<string, ComponentState> }) {
  const completed = componentOrder.filter(k => components[k]?.status === 'active').length;
  const total = componentOrder.length;
  const percent = total > 0 ? (completed / total) * 100 : 0;

  return (
    <div className="sc-prov__progress">
      <div className="sc-prov__progress-track">
        <motion.div
          className="sc-prov__progress-fill"
          initial={{ width: 0 }}
          animate={{ width: `${percent}%` }}
          transition={{ duration: 0.4, ease: [0.23, 1, 0.32, 1] }}
        />
      </div>
      <span className="sc-prov__progress-label">{completed}/{total}</span>
    </div>
  );
}

// ─── Main component ──────────────────────────────────────────────────────────

export function ProvisioningStatusCard({ bundleSlug, onComplete }: ProvisioningStatusCardProps) {
  const [result, setResult] = useState<ProvisioningResult | null>(null);
  const [isProvisioning, setIsProvisioning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const res = await apiClient.get<ProvisioningResult | { status: 'not_provisioned' }>('/v1/provisioning/status');
      if (res && 'status' in res && res.status === 'not_provisioned') {
        setResult(null);
        return null;
      }
      setResult(res as ProvisioningResult);
      return res as ProvisioningResult;
    } catch {
      return null;
    }
  }, []);

  useEffect(() => {
    const shouldPoll = isProvisioning || result?.status === 'provisioning';
    if (!shouldPoll) return;
    const interval = setInterval(async () => {
      const latest = await fetchStatus();
      if (latest?.status === 'active') {
        setIsProvisioning(false);
        onComplete?.();
      } else if (latest?.status === 'failed') {
        setIsProvisioning(false);
      }
    }, 1500);
    return () => clearInterval(interval);
  }, [isProvisioning, result?.status, fetchStatus, onComplete]);

  useEffect(() => {
    fetchStatus().then((data) => {
      if (data?.status === 'provisioning') setIsProvisioning(true);
      else if (data?.status === 'active') onComplete?.();
    });
  }, []);

  const handleProvision = async () => {
    setIsProvisioning(true);
    setError(null);
    try { await apiClient.post('/v1/provisioning/bundle', { bundle_slug: bundleSlug }); }
    catch (err: any) { setError(err?.response?.data?.error || 'Provisioning failed'); setIsProvisioning(false); }
  };

  const handleRetry = async () => {
    setIsProvisioning(true);
    setError(null);
    try { await apiClient.post('/v1/provisioning/retry'); }
    catch (err: any) { setError(err?.response?.data?.error || 'Retry failed'); setIsProvisioning(false); }
  };

  const allActive = result?.status === 'active';
  const hasFailures = result?.status === 'failed' || Object.values(result?.components || {}).some(c => c.status === 'failed');
  const componentOrder = BUNDLE_COMPONENTS[bundleSlug] || BUNDLE_COMPONENTS['saas-starter'];
  const components = result?.components || {};
  const bundleNames: Record<string, string> = { 'saas-starter': 'SaaS Starter', 'marketplace': 'Marketplace', 'ai-app': 'AI App' };

  return (
    <Chamber className="sc-prov">
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <AnnotationTag primary="MODULE PROV-01" secondary="Provisioning" position="top-right" />

      {/* Header */}
      <div className="sc-prov__header">
        <div>
          <h3 className="sc-prov__title">{bundleNames[bundleSlug] || bundleSlug}</h3>
          <p className="sc-prov__subtitle">
            {bundleSlug === 'ai-app' ? 'AI infrastructure — Vector DB, embeddings, chat, and memory' : 'Fully isolated production stack — zero platform leakage'}
          </p>
        </div>
        <AnimatePresence>
          {allActive && (
            <motion.div initial={{ scale: 0.8, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} transition={{ type: 'spring', stiffness: 400, damping: 20 }}>
              <StatusPill status="live" label="All Active" />
            </motion.div>
          )}
          {isProvisioning && !allActive && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
              <StatusPill status="pending" label="Provisioning" />
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Progress */}
      {Object.keys(components).length > 0 && (
        <ProgressStrip componentOrder={componentOrder} components={components} />
      )}

      {/* Component rows */}
      <div className="sc-prov__list">
        {componentOrder.map((key, index) => {
          const meta = COMPONENT_META[key];
          const state = components[key];
          const Icon = meta.icon;
          const status = state?.status || 'pending';

          return (
            <motion.div
              key={key}
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.06, duration: 0.3 }}
              className={`sc-prov__row ${status === 'provisioning' ? 'sc-prov__row--active' : ''} ${status === 'active' ? 'sc-prov__row--done' : ''} ${status === 'failed' ? 'sc-prov__row--fail' : ''}`}
            >
              <div className={`sc-prov__row-icon ${status === 'active' ? 'sc-prov__row-icon--ok' : status === 'provisioning' ? 'sc-prov__row-icon--spin' : status === 'failed' ? 'sc-prov__row-icon--fail' : ''}`}>
                <Icon className="sc-prov__row-icon-svg" />
              </div>

              <div className="sc-prov__row-body">
                <span className="sc-prov__row-label">{meta.label}</span>
                <span className="sc-prov__row-desc">
                  {status === 'provisioning' ? 'Configuring...' : status === 'active' ? 'Ready' : meta.description}
                </span>
                <AnimatePresence>
                  {state?.error && (
                    <motion.span initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="sc-prov__row-error">
                      {state.error}
                    </motion.span>
                  )}
                </AnimatePresence>
              </div>

              <StatusIndicator status={status} />
            </motion.div>
          );
        })}
      </div>

      {/* Error */}
      <AnimatePresence>
        {error && (
          <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="sc-prov__error">
            {error}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Actions */}
      <div className="sc-prov__actions">
        {!result && !isProvisioning && (
          <SealedButton onClick={handleProvision}>Provision {bundleNames[bundleSlug] || 'Everything'}</SealedButton>
        )}
        {isProvisioning && (
          <div className="sc-prov__provisioning-msg">
            <Loader2 className="sc-prov__provisioning-spinner" />
            <span>Provisioning isolated infrastructure...</span>
          </div>
        )}
        {allActive && (
          <motion.div initial={{ opacity: 0, y: 4 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="sc-prov__done-msg">
            Provisioned in {((result?.duration_ms || 0) < 1000 ? `${result?.duration_ms || 0}ms` : `${((result?.duration_ms || 0) / 1000).toFixed(1)}s`)}
          </motion.div>
        )}
        {hasFailures && !isProvisioning && (
          <FrameButton onClick={handleRetry}>Retry Failed Components</FrameButton>
        )}
      </div>
    </Chamber>
  );
}
