/**
 * FunctionEditorPage — Create / Edit Function page.
 *
 * Handles both /functions/new (create) and /functions/:id/edit (edit).
 * Layout: sticky ActionBar + two-column grid (sections left, code+summary right).
 */

import '@/styles/components.css';
import { ActionBar } from './ActionBar';
import { ConfigSummary } from './ConfigSummary';
import { FunctionEditorVaultDialogs } from './components/FunctionEditorVaultDialogs';
import { AdvancedSection } from './sections/AdvancedSection';
import { BasicInfoSection } from './sections/BasicInfoSection';
import { CodeEditorSection } from './sections/CodeEditorSection';
import { EnvVarsSection } from './sections/EnvVarsSection';
import { ResourceLimitsSection } from './sections/ResourceLimitsSection';
import { RuntimeSection } from './sections/RuntimeSection';
import { TriggersSection } from './sections/TriggersSection';
import { VisibilitySection } from './sections/VisibilitySection';
import { useFunctionEditor } from './useFunctionEditor';

export function FunctionEditorPage() {
  const editor = useFunctionEditor();

  return (
    <div className="min-h-screen" style={{ background: 'var(--bg-primary, #0a0a0f)' }}>
      <ActionBar editor={editor} />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 pb-24 pt-6">
        <div className="grid grid-cols-1 xl:grid-cols-[1fr_380px] gap-6">
          {/* Left column: form sections */}
          <div className="space-y-5">
            <BasicInfoSection editor={editor} />
            <RuntimeSection editor={editor} />
            <EnvVarsSection editor={editor} />
            <ResourceLimitsSection editor={editor} />
            <TriggersSection editor={editor} />
            <VisibilitySection editor={editor} />
            <AdvancedSection editor={editor} />
          </div>

          {/* Right column: code editor + summary */}
          <div className="space-y-5">
            <CodeEditorSection editor={editor} />
            <ConfigSummary editor={editor} />
          </div>
        </div>
      </div>

      <FunctionEditorVaultDialogs editor={editor} />
    </div>
  );
}
