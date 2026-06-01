import { useState } from 'react';
import { clsx } from 'clsx';
import type { ImportConfig } from '../../types/codePaste';
import { ImportErrorsList } from './ImportErrorsList';
import './ImportConfigPanel.css';

interface ImportConfigPanelProps {
  selectedCount: number;
  importConfig: ImportConfig;
  onConfigChange: (config: Partial<ImportConfig>) => void;
  onImport: () => void;
  onCancel: () => void;
  isImporting?: boolean;
  disabled?: boolean;
  importErrors?: Array<{ name: string; error: string }>;
}

const REGIONS = [
  { value: 'us-east-1', label: 'US East (N. Virginia)' },
  { value: 'us-west-2', label: 'US West (Oregon)' },
  { value: 'eu-west-1', label: 'EU West (Ireland)' },
  { value: 'eu-central-1', label: 'EU Central (Frankfurt)' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)' },
];

const PROVIDERS = [
  { value: 'cloud', label: '☁️ Cloud' },
  { value: 'edge', label: '⚡ Edge' },
  { value: 'local', label: '💻 Local' },
];

export function ImportConfigPanel({
  selectedCount,
  importConfig,
  onConfigChange,
  onImport,
  onCancel,
  isImporting = false,
  disabled = false,
  importErrors = [],
}: ImportConfigPanelProps) {
  const [showAdvanced, setShowAdvanced] = useState(false);

  const handleVisibilityChange = (visibility: 'private' | 'public') => {
    onConfigChange({ visibility });
    if (visibility === 'public') {
      setShowAdvanced(true);
    }
  };

  const handleProviderToggle = (provider: string) => {
    const current = importConfig.providers;
    const updated = current.includes(provider)
      ? current.filter((p) => p !== provider)
      : [...current, provider];
    onConfigChange({ providers: updated });
  };

  return (
    <div className={clsx('import-config-panel', { 'import-config-panel--disabled': disabled })}>
      <div className="import-config-panel__header">
        <h3>Import Configuration</h3>
        <span className="import-config-panel__count">
          {selectedCount} function{selectedCount !== 1 ? 's' : ''} selected
        </span>
      </div>

      <div className="import-config-panel__content">
        <div className="import-config-panel__section">
          <label className="import-config-panel__label">Visibility</label>
          <div className="import-config-panel__visibility-options">
            <button
              className={clsx('visibility-option', {
                'visibility-option--active': importConfig.visibility === 'private',
              })}
              onClick={() => handleVisibilityChange('private')}
              disabled={disabled}
            >
              <span className="visibility-option__icon">🔒</span>
              <span className="visibility-option__title">Private</span>
              <span className="visibility-option__desc">Only accessible by your tenant</span>
            </button>
            <button
              className={clsx('visibility-option', {
                'visibility-option--active': importConfig.visibility === 'public',
              })}
              onClick={() => handleVisibilityChange('public')}
              disabled={disabled}
            >
              <span className="visibility-option__icon">🌐</span>
              <span className="visibility-option__title">Public</span>
              <span className="visibility-option__desc">Publish to registry</span>
            </button>
          </div>
        </div>

        <div className="import-config-panel__section">
          <label className="import-config-panel__label">Region</label>
          <select
            className="import-config-panel__select"
            value={importConfig.region}
            onChange={(e) => onConfigChange({ region: e.target.value })}
            disabled={disabled}
          >
            {REGIONS.map((region) => (
              <option key={region.value} value={region.value}>
                {region.label}
              </option>
            ))}
          </select>
        </div>

        <div className="import-config-panel__section">
          <label className="import-config-panel__label">Providers</label>
          <div className="import-config-panel__providers">
            {PROVIDERS.map((provider) => (
              <label key={provider.value} className="provider-checkbox">
                <input
                  type="checkbox"
                  checked={importConfig.providers.includes(provider.value)}
                  onChange={() => handleProviderToggle(provider.value)}
                  disabled={disabled}
                />
                <span>{provider.label}</span>
              </label>
            ))}
          </div>
        </div>

        {importConfig.visibility === 'public' && (
          <div className="import-config-panel__section import-config-panel__section--advanced">
            <button
              className="import-config-panel__toggle-advanced"
              onClick={() => setShowAdvanced(!showAdvanced)}
            >
              Public Options {showAdvanced ? '▲' : '▼'}
            </button>

            {showAdvanced && (
              <div className="import-config-panel__advanced-fields">
                <div>
                  <label className="import-config-panel__label">Author</label>
                  <input
                    type="text"
                    className="import-config-panel__input"
                    placeholder="functionfly/author"
                    value={importConfig.author || ''}
                    onChange={(e) => onConfigChange({ author: e.target.value })}
                    disabled={disabled}
                  />
                </div>
                <div>
                  <label className="import-config-panel__label">Changelog</label>
                  <textarea
                    className="import-config-panel__textarea"
                    placeholder="Describe what this import does..."
                    value={importConfig.changelog || ''}
                    onChange={(e) => onConfigChange({ changelog: e.target.value })}
                    disabled={disabled}
                    rows={3}
                  />
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="import-config-panel__footer">
        {importErrors.length > 0 && (
          <ImportErrorsList errors={importErrors} />
        )}
        <button
          className="import-config-panel__cancel-btn"
          onClick={onCancel}
          disabled={disabled || isImporting}
        >
          Cancel
        </button>
        <button
          className={clsx('import-config-panel__import-btn', {
            'import-config-panel__import-btn--loading': isImporting,
          })}
          onClick={onImport}
          disabled={disabled || selectedCount === 0 || isImporting}
        >
          {isImporting ? (
            <>
              <span className="import-config-panel__spinner" />
              Importing...
            </>
          ) : (
            <>
              Import {selectedCount} Function{selectedCount !== 1 ? 's' : ''}
            </>
          )}
        </button>
      </div>
    </div>
  );
}