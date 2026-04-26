/**
 * Factory Settings Tab Component
 * Factory configuration form with all settings organized in sections
 */

import { Save, Settings } from 'lucide-react';
import type { FactoryConfig } from '@/lib/api/factory';

type FactorySettingsForm = Partial<FactoryConfig>;

interface FactorySettingsTabProps {
  settingsForm: FactorySettingsForm;
  isDirty: boolean;
  isSaving: boolean;
  onSettingChange: <K extends keyof FactorySettingsForm>(key: K, value: FactorySettingsForm[K]) => void;
  onSave: () => void;
}

export function FactorySettingsTab({
  settingsForm,
  isDirty,
  isSaving,
  onSettingChange,
  onSave,
}: FactorySettingsTabProps) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <Settings className="h-5 w-5" />
          Factory configuration
        </h2>
        <button
          onClick={onSave}
          disabled={!isDirty || isSaving}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 dark:bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-700 transition-colors disabled:opacity-50"
        >
          <Save className="h-4 w-4" />
          {isSaving ? 'Saving...' : 'Save changes'}
        </button>
      </div>

      <div className="grid gap-6">
        {/* Discovery */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Discovery</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Discovery batch size</span>
              <input
                type="number"
                min={1}
                max={100}
                value={settingsForm.discovery_batch_size ?? 10}
                onChange={(e) => onSettingChange('discovery_batch_size', parseInt(e.target.value, 10) || 10)}
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Discovery cooldown (minutes)</span>
              <input
                type="number"
                min={0}
                value={settingsForm.discovery_cooldown_minutes ?? 60}
                onChange={(e) => onSettingChange('discovery_cooldown_minutes', parseInt(e.target.value, 10) || 0)}
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block md:col-span-2 lg:col-span-1">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Discovery sources (comma-separated)</span>
              <input
                type="text"
                value={(settingsForm.discovery_sources ?? []).join(', ')}
                onChange={(e) =>
                  onSettingChange(
                    'discovery_sources',
                    e.target.value.split(',').map((s) => s.trim()).filter(Boolean)
                  )
                }
                placeholder="e.g. api, registry"
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
          </div>
        </section>

        {/* Quality & Testing */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Quality & testing</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Minimum quality score</span>
              <input
                type="number"
                min={0}
                max={100}
                step={0.1}
                value={settingsForm.minimum_quality_score ?? 70}
                onChange={(e) =>
                  onSettingChange('minimum_quality_score', parseFloat(e.target.value) || 0)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Minimum test score</span>
              <input
                type="number"
                min={0}
                max={100}
                step={0.1}
                value={settingsForm.minimum_test_score ?? 80}
                onChange={(e) =>
                  onSettingChange('minimum_test_score', parseFloat(e.target.value) || 0)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Approval required above quality</span>
              <input
                type="number"
                min={0}
                value={settingsForm.approval_required_above_quality ?? 0}
                onChange={(e) =>
                  onSettingChange('approval_required_above_quality', parseInt(e.target.value, 10) || 0)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Approval required above test</span>
              <input
                type="number"
                min={0}
                value={settingsForm.approval_required_above_test ?? 0}
                onChange={(e) =>
                  onSettingChange('approval_required_above_test', parseInt(e.target.value, 10) || 0)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settingsForm.require_all_tests_pass ?? true}
                onChange={(e) => onSettingChange('require_all_tests_pass', e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Require all tests to pass</span>
            </label>
          </div>
        </section>

        {/* Publishing */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Publishing</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settingsForm.auto_publish ?? true}
                onChange={(e) => onSettingChange('auto_publish', e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Auto-publish</span>
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Max opportunities per run</span>
              <input
                type="number"
                min={1}
                value={settingsForm.max_opportunities_per_run ?? 3}
                onChange={(e) =>
                  onSettingChange('max_opportunities_per_run', parseInt(e.target.value, 10) || 1)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Max versions per function</span>
              <input
                type="number"
                min={1}
                value={settingsForm.max_versions_per_function ?? 5}
                onChange={(e) =>
                  onSettingChange('max_versions_per_function', parseInt(e.target.value, 10) || 1)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
          </div>
        </section>

        {/* Scheduling */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Scheduling</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settingsForm.schedule_enabled ?? false}
                onChange={(e) => onSettingChange('schedule_enabled', e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Schedule enabled</span>
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Cron expression</span>
              <input
                type="text"
                value={settingsForm.schedule_cron ?? '0 0 * * *'}
                onChange={(e) => onSettingChange('schedule_cron', e.target.value)}
                placeholder="0 0 * * *"
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Timezone</span>
              <input
                type="text"
                value={settingsForm.schedule_timezone ?? 'UTC'}
                onChange={(e) => onSettingChange('schedule_timezone', e.target.value)}
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
          </div>
        </section>

        {/* Retries */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Retries</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Retry attempts</span>
              <input
                type="number"
                min={0}
                value={settingsForm.retry_attempts ?? 1}
                onChange={(e) => onSettingChange('retry_attempts', parseInt(e.target.value, 10) || 0)}
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Retry backoff (ms)</span>
              <input
                type="number"
                min={0}
                value={settingsForm.retry_backoff_ms ?? 500}
                onChange={(e) => onSettingChange('retry_backoff_ms', parseInt(e.target.value, 10) || 0)}
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
          </div>
        </section>

        {/* Notifications */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Notifications</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <label className="block md:col-span-2">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Webhook URL</span>
              <input
                type="url"
                value={settingsForm.notification_webhook_url ?? ''}
                onChange={(e) => onSettingChange('notification_webhook_url', e.target.value)}
                placeholder="https://..."
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settingsForm.notify_on_failure ?? true}
                onChange={(e) => onSettingChange('notify_on_failure', e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Notify on failure</span>
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settingsForm.notify_on_review_required ?? true}
                onChange={(e) => onSettingChange('notify_on_review_required', e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Notify on review required</span>
            </label>
          </div>
        </section>

        {/* Advanced */}
        <section className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-md font-semibold text-gray-900 dark:text-gray-100 mb-4">Advanced</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Rate limit (per hour)</span>
              <input
                type="number"
                min={0}
                value={settingsForm.rate_limit_per_hour ?? 10}
                onChange={(e) =>
                  onSettingChange('rate_limit_per_hour', parseInt(e.target.value, 10) || 0)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Max concurrent runs</span>
              <input
                type="number"
                min={1}
                value={settingsForm.max_concurrent_runs ?? 1}
                onChange={(e) =>
                  onSettingChange('max_concurrent_runs', parseInt(e.target.value, 10) || 1)
                }
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </label>
            <label className="block">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Log level</span>
              <select
                value={settingsForm.log_level ?? 'info'}
                onChange={(e) => onSettingChange('log_level', e.target.value)}
                className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              >
                <option value="debug">debug</option>
                <option value="info">info</option>
                <option value="warn">warn</option>
                <option value="error">error</option>
              </select>
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={settingsForm.dry_run_mode ?? false}
                onChange={(e) => onSettingChange('dry_run_mode', e.target.checked)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Dry run mode</span>
            </label>
          </div>
          <div className="mt-4">
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Feature flags (key: true/false, one per line)</span>
            <textarea
              rows={4}
              value={Object.entries(settingsForm.feature_flags ?? {})
                .map(([k, v]) => `${k}=${v}`)
                .join('\n')}
              onChange={(e) => {
                const flags: Record<string, boolean> = {};
                e.target.value
                  .split('\n')
                  .map((s) => s.trim())
                  .filter(Boolean)
                  .forEach((line) => {
                    const eq = line.indexOf('=');
                    if (eq > 0) {
                      const key = line.slice(0, eq).trim();
                      const val = line.slice(eq + 1).trim().toLowerCase();
                      flags[key] = val === 'true' || val === '1';
                    }
                  });
                onSettingChange('feature_flags', flags);
              }}
              placeholder="feature_a=true\nfeature_b=false"
              className="mt-1 block w-full rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 font-mono text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            />
          </div>
        </section>
      </div>
    </div>
  );
}
