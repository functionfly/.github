'use client';

import { useState, useEffect } from 'react';
import { Settings, Save, RotateCcw } from 'lucide-react';
import {
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
} from '@/components/containment';

interface AtlasConfigPanelProps {
  config: {
    sampling_rate: number;
    trace_errors_only: boolean;
    sample_head_percent: number;
    sample_tail_count: number;
    retention_days: number;
  } | null;
  onUpdate: (updates: any) => Promise<any>;
  loading?: boolean;
}

export default function AtlasConfigPanel({ config, onUpdate, loading }: AtlasConfigPanelProps) {
  const [localConfig, setLocalConfig] = useState(config);
  const [saving, setSaving] = useState(false);
  const [showPanel, setShowPanel] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  useEffect(() => {
    if (config) {
      setLocalConfig(config);
    }
  }, [config]);

  const handleSave = async () => {
    if (!localConfig) return;
    setSaving(true);
    setSaveError(null);
    setSaveSuccess(false);
    try {
      await onUpdate(localConfig);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (e: any) {
      setSaveError(e?.message || 'Failed to save configuration');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    setLocalConfig(config);
    setSaveError(null);
    setSaveSuccess(false);
  };

  if (!showPanel) {
    return (
      <FrameButton size="sm" onClick={() => setShowPanel(true)} iconLeft={<Settings className="atlas-icon-sm" />}>
        Configure Atlas
      </FrameButton>
    );
  }

  return (
    <Chamber className="atlas-config">
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div className="atlas-config__header">
        <div>
          <h2 className="atlas-config__title"><Settings className="atlas-icon-sm" /> Atlas Configuration</h2>
          <p className="atlas-config__desc">Configure observability sampling and retention settings</p>
        </div>
        <button className="atlas-config__close" onClick={() => setShowPanel(false)}>Close</button>
      </div>

      {loading ? (
        <p className="atlas-config__loading">Loading configuration...</p>
      ) : localConfig ? (
        <div className="atlas-config__body">
          {/* Sampling Rate */}
          <div className="atlas-config__row">
            <div>
              <label className="atlas-config__label">Sampling Rate</label>
              <p className="atlas-config__hint">Percentage of runs to sample (0–100)</p>
            </div>
            <div className="atlas-config__control">
              <input type="range" min={0} max={100} step={1} value={Math.round(localConfig.sampling_rate * 100)}
                onChange={(e) => setLocalConfig({ ...localConfig, sampling_rate: Number(e.target.value) / 100 })} className="atlas-range" />
              <span className="atlas-config__value">{Math.round(localConfig.sampling_rate * 100)}%</span>
            </div>
          </div>

          {/* Trace Errors Only */}
          <div className="atlas-config__row">
            <div>
              <label className="atlas-config__label">Trace Errors Only</label>
              <p className="atlas-config__hint">Only trace runs with errors</p>
            </div>
            <button className={`atlas-switch ${localConfig.trace_errors_only ? 'atlas-switch--on' : ''}`}
              onClick={() => setLocalConfig({ ...localConfig, trace_errors_only: !localConfig.trace_errors_only })}>
              <span className="atlas-switch__thumb" />
            </button>
          </div>

          {/* Sample Head Percent */}
          <div className="atlas-config__row">
            <div>
              <label className="atlas-config__label">Sample Head Percent</label>
              <p className="atlas-config__hint">Sample first N% of tokens per run</p>
            </div>
            <div className="atlas-config__control">
              <input type="range" min={0} max={100} step={1} value={localConfig.sample_head_percent}
                onChange={(e) => setLocalConfig({ ...localConfig, sample_head_percent: Number(e.target.value) })} className="atlas-range" />
              <span className="atlas-config__value">{localConfig.sample_head_percent}%</span>
            </div>
          </div>

          {/* Sample Tail Count */}
          <div className="atlas-config__row">
            <div>
              <label className="atlas-config__label">Sample Tail Count</label>
              <p className="atlas-config__hint">Number of final events to always capture</p>
            </div>
            <input type="number" min={0} value={localConfig.sample_tail_count}
              onChange={(e) => setLocalConfig({ ...localConfig, sample_tail_count: parseInt(e.target.value) || 0 })}
              className="atlas-input atlas-input--sm" />
          </div>

          {/* Retention Days */}
          <div className="atlas-config__row">
            <div>
              <label className="atlas-config__label">Retention Days</label>
              <p className="atlas-config__hint">How long to keep observability data</p>
            </div>
            <div className="atlas-config__control">
              <input type="number" min={1} value={localConfig.retention_days}
                onChange={(e) => setLocalConfig({ ...localConfig, retention_days: parseInt(e.target.value) || 0 })}
                className="atlas-input atlas-input--sm" />
              <span className="atlas-config__unit">days</span>
            </div>
          </div>

          {/* Actions */}
          <div className="atlas-config__actions">
            <SealedButton onClick={handleSave} disabled={saving} loading={saving} iconLeft={<Save className="atlas-icon-sm" />}>
              Save Changes
            </SealedButton>
            <FrameButton onClick={handleReset} iconLeft={<RotateCcw className="atlas-icon-sm" />}>
              Reset
            </FrameButton>
          </div>
          {saveError && (
            <p className="atlas-config__error">{saveError}</p>
          )}
          {saveSuccess && (
            <p className="atlas-config__success">Configuration saved successfully</p>
          )}
        </div>
      ) : (
        <p className="atlas-config__loading">No configuration available</p>
      )}
    </Chamber>
  );
}
