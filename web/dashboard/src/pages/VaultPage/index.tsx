/**
 * VaultPage
 *
 * Unified vault UI surface combining secrets management with enterprise features.
 * Tabs map to plan-gated feature groups:
 *
 *   Secrets        — Namespace tree + full SecretList (CRUD, search, rotation)
 *   Security       — MFA, IP allowlist, expiration, break-glass, escrow
 *   Dynamic Creds  — DB targets + credential templates + leases
 *   Access         — RBAC roles + my assignments
 *   Enterprise     — namespaces, shares, SSO, SIEM webhooks
 *   Activity       — audit log + export + token monitor + dep graph
 */

import "@/styles/professional-dashboard.css";

import { useSearchParams } from "react-router-dom";
import { useCurrentPlan } from "@/hooks/useCurrentPlan";
import { usePageTitle } from "@/hooks";
import { platformToVaultPlan } from "@/lib/vaultPlans";
import type { VaultPlan } from "@/types/vault-enterprise";
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  AnnotationTag,
  StatusPill,
} from "@/components/containment";

import { SecretsTab } from "@/components/VaultPage/tabs/SecretsTab";
import { SecurityTab } from "@/components/VaultPage/tabs/SecurityTab";
import { DynamicCredsTab } from "@/components/VaultPage/tabs/DynamicCredsTab";
import { AccessTab } from "@/components/VaultPage/tabs/AccessTab";
import { EnterpriseTab } from "@/components/VaultPage/tabs/EnterpriseTab";
import { ActivityTab } from "@/components/VaultPage/tabs/ActivityTab";

import "./styles.css";

const TABS: { value: string; label: string; plan: VaultPlan[] }[] = [
  { value: "secrets", label: "Secrets", plan: ["free", "pro", "team", "enterprise"] },
  { value: "security", label: "Security", plan: ["pro", "team", "enterprise"] },
  { value: "dynamic", label: "Dynamic Creds", plan: ["pro", "team", "enterprise"] },
  { value: "access", label: "Access", plan: ["team", "enterprise"] },
  { value: "enterprise", label: "Enterprise", plan: ["team", "enterprise"] },
  { value: "activity", label: "Activity", plan: ["pro", "team", "enterprise"] },
];

interface VaultPageProps {
  deprecationBanner?: React.ReactNode;
}

export function VaultPage({ deprecationBanner }: VaultPageProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = searchParams.get("tab") || "secrets";
  const { plan: platformPlan, isLoading } = useCurrentPlan();
  const plan = platformToVaultPlan(platformPlan) as VaultPlan;

  const handleTabChange = (newTab: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      next.set("tab", newTab);
      return next;
    });
  };

  const currentTab = TABS.find((t) => t.value === tab);
  usePageTitle(currentTab ? `Vault - ${currentTab.label}` : "Vault");

  return (
    <div className="vault-page">
      <PageGrid />

      {deprecationBanner}

      {/* Hero */}
      <Chamber className="vault-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE VT-01" secondary="Secrets Vault" position="top-right" />

        <div className="vault-hero__header">
          <div className="vault-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="vault-hero__title">Secrets Vault</h1>
            <StatusPill status="live" label="Zero-Knowledge" />
          </div>
          <p className="vault-hero__subtitle">
            Zero-knowledge encryption, dynamic credentials, and enterprise controls.
          </p>
        </div>
      </Chamber>

      {/* Tabs */}
      <div className="vault-tabs">
        {TABS.map((t) => {
          const disabled = isLoading || !t.plan.includes(plan);
          return (
            <button
              key={t.value}
              className={`vault-tab ${tab === t.value ? 'vault-tab--active' : ''}`}
              onClick={() => !disabled && handleTabChange(t.value)}
              disabled={disabled}
            >
              {t.label}
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      <div className="vault-tab-content">
        {tab === "secrets" && <SecretsTab plan={plan} />}
        {tab === "security" && <SecurityTab plan={plan} />}
        {tab === "dynamic" && <DynamicCredsTab plan={plan} />}
        {tab === "access" && <AccessTab plan={plan} />}
        {tab === "enterprise" && <EnterpriseTab plan={plan} />}
        {tab === "activity" && <ActivityTab plan={plan} />}
      </div>
    </div>
  );
}

export default VaultPage;
