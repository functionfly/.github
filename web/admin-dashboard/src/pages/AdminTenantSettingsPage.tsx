/**
 * Admin Tenant Settings Page
 * Manage tenant security settings including SAML, MFA, and Session policies
 */

import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import {
  Shield,
  Key,
  Clock,
  Save,
  RefreshCw,
  Copy,
  Eye,
  EyeOff,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Download,
  ExternalLink
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { MFA_POLICY_OPTIONS, SESSION_POLICY_DEFAULTS } from '@/lib/constants';
import type {
  Tenant,
  SAMLConfig,
  SAMLMetadata,
  MFAPolicy,
  SessionPolicy,
  ActiveSession
} from '@/types';

type SettingsTab = 'saml' | 'mfa' | 'sessions';

export function AdminTenantSettingsPage() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const [activeTab, setActiveTab] = useState<SettingsTab>('saml');
  const [showCertificates, setShowCertificates] = useState(false);
  const queryClient = useQueryClient();

  // SAML form state
  const [samlConfig, setSamlConfig] = useState<SAMLConfig>({
    enabled: false,
    entity_id: '',
    sso_url: '',
    slo_url: '',
    certificate: '',
  });

  // MFA form state
  const [mfaPolicy, setMfaPolicy] = useState<MFAPolicy>('optional');

  // Session policy form state
  const [sessionPolicy, setSessionPolicy] = useState<SessionPolicy>({
    max_duration_minutes: SESSION_POLICY_DEFAULTS.max_duration_minutes,
    idle_timeout_minutes: SESSION_POLICY_DEFAULTS.idle_timeout_minutes,
    max_concurrent_sessions: SESSION_POLICY_DEFAULTS.max_concurrent_sessions,
  });

  // Fetch tenant info
  const { data: tenantResponse, isLoading: tenantLoading } = useQuery({
    queryKey: ['admin-tenant', tenantId],
    queryFn: async () => {
      if (!tenantId) return null;
      return await adminApiClient.get<Tenant>(`/tenants/${tenantId}`);
    },
    enabled: !!tenantId,
  });

  // API may return { data: tenant } (new) or the tenant object directly (legacy)
  const raw = tenantResponse as { data?: Tenant } | Tenant | undefined;
  const tenant =
    raw && typeof raw === 'object' && 'data' in raw && raw.data != null
      ? raw.data
      : raw && typeof raw === 'object' && 'id' in raw && 'name' in raw
        ? (raw as Tenant)
        : undefined;

  // Fetch SAML config
  const { data: samlResponse, isLoading: samlLoading } = useQuery({
    queryKey: ['admin-saml-config', tenantId],
    queryFn: async () => {
      if (!tenantId) return null;
      return await adminApiClient.get<SAMLConfig>(`/tenants/${tenantId}/saml/config`);
    },
    enabled: !!tenantId && activeTab === 'saml',
  });

  // Fetch SP metadata
  const { data: metadataResponse } = useQuery({
    queryKey: ['admin-saml-metadata', tenantId],
    queryFn: async () => {
      if (!tenantId) return null;
      return await adminApiClient.get<SAMLMetadata>(`/tenants/${tenantId}/saml/metadata`);
    },
    enabled: !!tenantId && activeTab === 'saml',
  });

  // Fetch MFA policy
  const { data: mfaResponse, isLoading: mfaLoading } = useQuery({
    queryKey: ['admin-mfa-policy', tenantId],
    queryFn: async () => {
      if (!tenantId) return null;
      return await adminApiClient.get<{ policy: MFAPolicy }>(`/tenants/${tenantId}/mfa-policy`);
    },
    enabled: !!tenantId && activeTab === 'mfa',
  });

  // Fetch session policy
  const { data: sessionResponse, isLoading: sessionLoading } = useQuery({
    queryKey: ['admin-session-policy', tenantId],
    queryFn: async () => {
      if (!tenantId) return null;
      return await adminApiClient.get<SessionPolicy>(`/tenants/${tenantId}/session-policy`);
    },
    enabled: !!tenantId && activeTab === 'sessions',
  });

  // Fetch active sessions
  const { data: sessionsResponse } = useQuery({
    queryKey: ['admin-active-sessions', tenantId],
    queryFn: async () => {
      if (!tenantId) return null;
      return await adminApiClient.get<ActiveSession[]>(`/tenants/${tenantId}/active-sessions`);
    },
    enabled: !!tenantId && activeTab === 'sessions',
  });

  // Update state when data loads
  useEffect(() => {
    if (samlResponse?.data) {
      setSamlConfig(samlResponse.data);
    }
  }, [samlResponse]);

  useEffect(() => {
    if (mfaResponse?.data) {
      setMfaPolicy(mfaResponse.data.policy);
    }
  }, [mfaResponse]);

  useEffect(() => {
    if (sessionResponse?.data) {
      setSessionPolicy(sessionResponse.data);
    }
  }, [sessionResponse]);

  // Save SAML config mutation
  const saveSamlMutation = useMutation({
    mutationFn: (config: SAMLConfig) =>
      adminApiClient.put(`/tenants/${tenantId}/saml/config`, config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-saml-config', tenantId] });
    },
  });

  // Save MFA policy mutation
  const saveMfaMutation = useMutation({
    mutationFn: (policy: MFAPolicy) =>
      adminApiClient.put(`/tenants/${tenantId}/mfa-policy`, { policy }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-mfa-policy', tenantId] });
    },
  });

  // Save session policy mutation
  const saveSessionMutation = useMutation({
    mutationFn: (policy: SessionPolicy) =>
      adminApiClient.put(`/tenants/${tenantId}/session-policy`, policy),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-session-policy', tenantId] });
    },
  });

  // Revoke session mutation
  const revokeSessionMutation = useMutation({
    mutationFn: (sessionId: string) =>
      adminApiClient.delete(`/sessions/${sessionId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-active-sessions', tenantId] });
    },
  });

  const handleSaveSaml = () => {
    saveSamlMutation.mutate(samlConfig);
  };

  const handleSaveMfa = () => {
    saveMfaMutation.mutate(mfaPolicy);
  };

  const handleSaveSession = () => {
    saveSessionMutation.mutate(sessionPolicy);
  };

  const handleRevokeSession = (sessionId: string) => {
    if (confirm('Are you sure you want to revoke this session?')) {
      revokeSessionMutation.mutate(sessionId);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const downloadMetadata = () => {
    if (metadataResponse?.data?.metadata_xml) {
      const blob = new Blob([metadataResponse.data.metadata_xml], { type: 'application/xml' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `sp-metadata-${tenantId}.xml`;
      a.click();
    }
  };

  if (tenantLoading) {
    return <LoadingScreen />;
  }

  if (!tenant) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg dark:bg-red-900/20 dark:border-red-800">
        <h3 className="font-semibold text-red-900 dark:text-red-200">Tenant not found</h3>
      </div>
    );
  }

  const tabs: { id: SettingsTab; label: string; icon: React.ReactNode }[] = [
    { id: 'saml', label: 'SAML SSO', icon: <Key className="w-4 h-4" /> },
    { id: 'mfa', label: 'MFA Policy', icon: <Shield className="w-4 h-4" /> },
    { id: 'sessions', label: 'Sessions', icon: <Clock className="w-4 h-4" /> },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            {tenant.name} - Security Settings
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Configure authentication and security policies
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="flex gap-4">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-3 border-b-2 font-medium text-sm transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* SAML Tab */}
      {activeTab === 'saml' && (
        <div className="space-y-6">
          {samlLoading ? (
            <LoadingScreen />
          ) : (
            <>
              {/* SAML Status */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-white">SAML Configuration</h2>
                  <div className="flex items-center gap-2">
                    {samlConfig.enabled ? (
                      <span className="inline-flex items-center gap-1 px-3 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded-full text-sm">
                        <CheckCircle className="w-4 h-4" />
                        Enabled
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 px-3 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full text-sm">
                        <XCircle className="w-4 h-4" />
                        Disabled
                      </span>
                    )}
                  </div>
                </div>

                <div className="space-y-4">
                  {/* Enable/Disable Toggle */}
                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="saml_enabled"
                      checked={samlConfig.enabled}
                      onChange={(e) => setSamlConfig({ ...samlConfig, enabled: e.target.checked })}
                      className="rounded border-gray-300 dark:border-gray-600"
                    />
                    <label htmlFor="saml_enabled" className="text-sm text-gray-700 dark:text-gray-300">
                      Enable SAML Single Sign-On for this tenant
                    </label>
                  </div>

                  {/* IdP Configuration */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        IdP Entity ID
                      </label>
                      <input
                        type="text"
                        value={samlConfig.entity_id || ''}
                        onChange={(e) => setSamlConfig({ ...samlConfig, entity_id: e.target.value })}
                        placeholder="https://idp.example.com/entity"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        SSO URL
                      </label>
                      <input
                        type="url"
                        value={samlConfig.sso_url || ''}
                        onChange={(e) => setSamlConfig({ ...samlConfig, sso_url: e.target.value })}
                        placeholder="https://idp.example.com/sso"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        SLO URL (Optional)
                      </label>
                      <input
                        type="url"
                        value={samlConfig.slo_url || ''}
                        onChange={(e) => setSamlConfig({ ...samlConfig, slo_url: e.target.value })}
                        placeholder="https://idp.example.com/slo"
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                      />
                    </div>
                  </div>

                  {/* Certificate */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      IdP Certificate (X.509)
                    </label>
                    <div className="relative">
                      <textarea
                        value={samlConfig.certificate || ''}
                        onChange={(e) => setSamlConfig({ ...samlConfig, certificate: e.target.value })}
                        placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                        rows={6}
                        className={`w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white font-mono text-xs ${!showCertificates ? 'filter blur-sm' : ''}`}
                      />
                      <button
                        type="button"
                        onClick={() => setShowCertificates(!showCertificates)}
                        className="absolute top-2 right-2 p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                      >
                        {showCertificates ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>

                  <div className="flex justify-end">
                    <button
                      onClick={handleSaveSaml}
                      disabled={saveSamlMutation.isPending}
                      className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                      {saveSamlMutation.isPending ? (
                        <RefreshCw className="w-4 h-4 animate-spin" />
                      ) : (
                        <Save className="w-4 h-4" />
                      )}
                      Save Configuration
                    </button>
                  </div>
                </div>
              </div>

              {/* SP Metadata */}
              {metadataResponse?.data && (
                <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
                  <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Service Provider Metadata</h2>
                    <button
                      onClick={downloadMetadata}
                      className="flex items-center gap-2 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 text-sm"
                    >
                      <Download className="w-4 h-4" />
                      Download XML
                    </button>
                  </div>

                  <div className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <div className="p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">SP Entity ID</p>
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-mono text-gray-900 dark:text-white truncate">
                            {metadataResponse.data.entity_id}
                          </p>
                          <button
                            onClick={() => copyToClipboard(metadataResponse.data.entity_id)}
                            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                          >
                            <Copy className="w-3 h-3" />
                          </button>
                        </div>
                      </div>

                      <div className="p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">ACS URL</p>
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-mono text-gray-900 dark:text-white truncate">
                            {metadataResponse.data.acs_url}
                          </p>
                          <button
                            onClick={() => copyToClipboard(metadataResponse.data.acs_url)}
                            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                          >
                            <Copy className="w-3 h-3" />
                          </button>
                        </div>
                      </div>

                      <div className="p-3 bg-gray-50 dark:bg-gray-900 rounded-lg">
                        <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">SLO URL</p>
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-mono text-gray-900 dark:text-white truncate">
                            {metadataResponse.data.slo_url || '—'}
                          </p>
                          {metadataResponse.data.slo_url && (
                            <button
                              onClick={() => copyToClipboard(metadataResponse.data.slo_url)}
                              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                            >
                              <Copy className="w-3 h-3" />
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* MFA Tab */}
      {activeTab === 'mfa' && (
        <div className="space-y-6">
          {mfaLoading ? (
            <LoadingScreen />
          ) : (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">MFA Policy</h2>
              <p className="text-sm text-gray-600 dark:text-gray-400 mb-6">
                Configure the multi-factor authentication policy for all users in this tenant.
              </p>

              <div className="space-y-3">
                {MFA_POLICY_OPTIONS.map((option) => (
                  <label
                    key={option.value}
                    className={`flex items-start gap-3 p-4 border rounded-lg cursor-pointer transition-colors ${
                      mfaPolicy === option.value
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
                    }`}
                  >
                    <input
                      type="radio"
                      name="mfa_policy"
                      value={option.value}
                      checked={mfaPolicy === option.value}
                      onChange={(e) => setMfaPolicy(e.target.value as MFAPolicy)}
                      className="mt-1"
                    />
                    <div>
                      <span className="font-medium text-gray-900 dark:text-white">{option.label}</span>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{option.description}</p>
                    </div>
                  </label>
                ))}
              </div>

              <div className="flex justify-end mt-6">
                <button
                  onClick={handleSaveMfa}
                  disabled={saveMfaMutation.isPending}
                  className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                >
                  {saveMfaMutation.isPending ? (
                    <RefreshCw className="w-4 h-4 animate-spin" />
                  ) : (
                    <Save className="w-4 h-4" />
                  )}
                  Save Policy
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Sessions Tab */}
      {activeTab === 'sessions' && (
        <div className="space-y-6">
          {sessionLoading ? (
            <LoadingScreen />
          ) : (
            <>
              {/* Session Policy */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Session Policy</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-6">
                  Configure session duration and concurrency limits.
                </p>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Max Session Duration (minutes)
                    </label>
                    <input
                      type="number"
                      value={sessionPolicy.max_duration_minutes}
                      onChange={(e) => setSessionPolicy({
                        ...sessionPolicy,
                        max_duration_minutes: parseInt(e.target.value) || SESSION_POLICY_DEFAULTS.max_duration_minutes
                      })}
                      min={5}
                      max={525600} // 1 year
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                    <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                      Default: {SESSION_POLICY_DEFAULTS.max_duration_minutes} min (30 days)
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Idle Timeout (minutes)
                    </label>
                    <input
                      type="number"
                      value={sessionPolicy.idle_timeout_minutes}
                      onChange={(e) => setSessionPolicy({
                        ...sessionPolicy,
                        idle_timeout_minutes: parseInt(e.target.value) || SESSION_POLICY_DEFAULTS.idle_timeout_minutes
                      })}
                      min={5}
                      max={10080} // 1 week
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                    <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                      Default: {SESSION_POLICY_DEFAULTS.idle_timeout_minutes} min (1 hour)
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Max Concurrent Sessions
                    </label>
                    <input
                      type="number"
                      value={sessionPolicy.max_concurrent_sessions}
                      onChange={(e) => setSessionPolicy({
                        ...sessionPolicy,
                        max_concurrent_sessions: parseInt(e.target.value) || SESSION_POLICY_DEFAULTS.max_concurrent_sessions
                      })}
                      min={1}
                      max={100}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
                    />
                    <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                      Default: {SESSION_POLICY_DEFAULTS.max_concurrent_sessions}
                    </p>
                  </div>
                </div>

                <div className="flex justify-end mt-6">
                  <button
                    onClick={handleSaveSession}
                    disabled={saveSessionMutation.isPending}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {saveSessionMutation.isPending ? (
                      <RefreshCw className="w-4 h-4 animate-spin" />
                    ) : (
                      <Save className="w-4 h-4" />
                    )}
                    Save Policy
                  </button>
                </div>
              </div>

              {/* Active Sessions */}
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Active Sessions</h2>
                  <span className="text-sm text-gray-500 dark:text-gray-400">
                    {sessionsResponse?.data?.length || 0} active
                  </span>
                </div>

                {sessionsResponse?.data && sessionsResponse.data.length > 0 ? (
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-gray-200 dark:border-gray-700">
                          <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">User</th>
                          <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">IP Address</th>
                          <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Created</th>
                          <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Last Activity</th>
                          <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Expires</th>
                          <th className="px-4 py-2 text-left text-sm font-medium text-gray-700 dark:text-gray-300">Actions</th>
                        </tr>
                      </thead>
                      <tbody>
                        {sessionsResponse.data.map((session) => (
                          <tr key={session.id} className="border-b border-gray-100 dark:border-gray-700">
                            <td className="px-4 py-3 text-sm text-gray-900 dark:text-white">
                              {session.user_email}
                            </td>
                            <td className="px-4 py-3 text-sm font-mono text-gray-600 dark:text-gray-300">
                              {session.ip_address}
                            </td>
                            <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300">
                              {new Date(session.created_at).toLocaleString()}
                            </td>
                            <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300">
                              {new Date(session.last_activity_at).toLocaleString()}
                            </td>
                            <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300">
                              {new Date(session.expires_at).toLocaleString()}
                            </td>
                            <td className="px-4 py-3">
                              <button
                                onClick={() => handleRevokeSession(session.id)}
                                disabled={revokeSessionMutation.isPending}
                                className="text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 text-sm"
                              >
                                Revoke
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p className="text-sm text-gray-500 dark:text-gray-400 text-center py-8">
                    No active sessions
                  </p>
                )}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}

export default AdminTenantSettingsPage;
