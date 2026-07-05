/**
 * FunctionEditorPage — Create / Edit Function page.
 *
 * Handles both /functions/new (create) and /functions/:id/edit (edit).
 * Layout: sticky ActionBar + two-column grid (sections left, code+summary right).
 * Mobile: Collapsible sections with summary preview always visible.
 */

import '@/styles/components.css';
import { useTranslation } from 'react-i18next';
import { ActionBar } from './ActionBar';
import { ConfigSummary } from './ConfigSummary';
import { FunctionEditorVaultDialogs } from './components/FunctionEditorVaultDialogs';
import { MobilePreviewToggle } from './components/MobilePreviewToggle';
import { SectionCollapsible } from './components/SectionCollapsible';
import { AdvancedSection } from './sections/AdvancedSection';
import { BasicInfoSection } from './sections/BasicInfoSection';
import { CodeEditorSection } from './sections/CodeEditorSection';
import { CodeEditorSectionMobile } from './sections/CodeEditorSectionMobile';
import { DeployTargetSection } from './sections/DeployTargetSection';
import { EnvVarsSection } from './sections/EnvVarsSection';
import { ResourceLimitsSection } from './sections/ResourceLimitsSection';
import { RuntimeSection } from './sections/RuntimeSection';
import { TemplateGallery } from './sections/TemplateGallery';
import { TriggersSection } from './sections/TriggersSection';
import { VisibilitySection } from './sections/VisibilitySection';
import { useFunctionEditor } from './useFunctionEditor';

export function FunctionEditorPage() {
  const { t } = useTranslation();
  const editor = useFunctionEditor();
  const { isEditing } = editor;

  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--bg)' }}>
      <ActionBar editor={editor} />

      {/* Mobile Preview Toggle - only visible on small screens */}
      <MobilePreviewToggle editor={editor} />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 pb-24 pt-6">
        <div className="grid grid-cols-1 lg:grid-cols-[1fr_380px] xl:grid-cols-[1fr_420px] gap-6 items-start">
          {/* Left column: form sections */}
          <div className="space-y-4 sm:space-y-5">
            {/* Template Gallery - only shown for new functions */}
            {!isEditing && <TemplateGallery editor={editor} />}

            {/* Core sections - always expanded */}
            <BasicInfoSection editor={editor} />
            <RuntimeSection editor={editor} />
            <CodeEditorSectionMobile editor={editor} />

            {/* Collapsible sections for progressive disclosure */}
            <SectionCollapsible title={t('funcEditor.environmentResources')} defaultOpen={false}>
              <div className="space-y-4">
                <EnvVarsSection editor={editor} />
                <ResourceLimitsSection editor={editor} />
              </div>
            </SectionCollapsible>

            <SectionCollapsible title={t('funcEditor.triggersAccess')} defaultOpen={false}>
              <div className="space-y-4">
                <TriggersSection editor={editor} />
                <VisibilitySection editor={editor} />
              </div>
            </SectionCollapsible>

            <SectionCollapsible title={t('funcEditor.deploymentAdvanced')} defaultOpen={false}>
              <div className="space-y-4">
                <DeployTargetSection editor={editor} />
                <AdvancedSection editor={editor} />
              </div>
            </SectionCollapsible>
          </div>

          {/* Right column: code editor + summary - hidden on mobile, sticky on desktop */}
          <div className="hidden lg:block space-y-5 xl:sticky xl:top-[72px]">
            <CodeEditorSection editor={editor} />
            <ConfigSummary editor={editor} />
          </div>
        </div>
      </div>

      <FunctionEditorVaultDialogs editor={editor} />
    </div>
  );
}
