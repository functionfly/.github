import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ArrowRight, CheckCircle2, Circle, FileText, Loader2, Rocket, Server } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { FieldError } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

function WorkflowStep({
  step,
  current,
  completed,
  icon,
  label,
  description,
}: {
  step: number;
  current: boolean;
  completed: boolean;
  icon: React.ReactNode;
  label: string;
  description?: string;
}) {
  return (
    <div className={`flex items-start gap-3 p-3 rounded-[var(--radius)] transition-colors ${
      current ? 'border' : ''
    }`} style={{
      background: current ? 'rgba(143, 255, 208, 0.03)' : 'var(--panel-raised)',
      borderColor: current ? 'rgba(143, 255, 208, 0.15)' : 'transparent',
    }}>
      <div className={`flex items-center justify-center w-6 h-6 rounded-full shrink-0 ${
        completed
          ? ''
          : current
          ? ''
          : ''
      }`} style={{
        background: completed ? 'rgba(143, 255, 208, 0.1)' : current ? 'var(--status-ok)' : 'var(--panel)',
        color: completed ? 'var(--status-ok)' : current ? 'var(--bg)' : 'var(--text-faint)',
        border: !completed && !current ? '1px solid var(--panel-edge)' : 'none',
      }}>
        {completed ? <CheckCircle2 className="w-3.5 h-3.5" /> : <span className="text-xs font-medium">{step}</span>}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium" style={{
            color: current ? 'var(--status-ok)' : completed ? 'var(--status-ok)' : 'var(--text-faint)',
          }}>
            {label}
          </span>
          {icon}
        </div>
        {description && <p className="text-xs text-[var(--text-faint)] mt-0.5">{description}</p>}
      </div>
    </div>
  );
}

export function DeployTargetSection({ editor }: Props) {
  const { t } = useTranslation();
  const {
    isEditing,
    isDirty,
    linkedAppId,
    deployBackendsLoading,
    filteredDeployBackends,
    selectedDeployBackendId,
    setDeployBackendId,
    errors,
  } = editor;

  // Determine workflow step
  const isDraftStep = !isEditing;
  const isBackendStep = isEditing && !selectedDeployBackendId && filteredDeployBackends.length > 0;
  const isDeployStep = isEditing && selectedDeployBackendId;

  if (!isEditing) {
    return (
      <Card className="overflow-hidden" style={{ background: 'var(--panel)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-chamber)' }}>
        <CardHeader className="pb-2 pt-4 px-5">
            <CardTitle className="text-sm font-semibold text-[var(--text)] flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
            <Server className="w-4 h-4 text-[var(--status-ok)]" />
            {t('funcEditor.deployTarget')}
          </CardTitle>
        </CardHeader>
        <CardContent className="px-5 pb-4 space-y-4">
          {/* Workflow visualizer */}
          <div className="space-y-2">
            <WorkflowStep
              step={1}
              current={isDraftStep}
              completed={false}
              icon={<FileText className="w-3.5 h-3.5 text-[var(--text-faint)]" />}
              label={t('funcEditor.draft')}
              description={t('funcEditor.draftDescription')}
            />
            <div className="flex items-center justify-center py-0.5">
              <ArrowRight className="w-4 h-4 text-[var(--text-faint)]/30" />
            </div>
            <WorkflowStep
              step={2}
              current={false}
              completed={false}
              icon={<Server className="w-3.5 h-3.5 text-[var(--text-faint)]" />}
              label={t('funcEditor.selectBackend')}
              description={t('funcEditor.selectBackendDescription')}
            />
            <div className="flex items-center justify-center py-0.5">
              <ArrowRight className="w-4 h-4 text-[var(--text-faint)]/30" />
            </div>
            <WorkflowStep
              step={3}
              current={false}
              completed={false}
              icon={<Rocket className="w-3.5 h-3.5 text-[var(--text-faint)]" />}
              label={t('funcEditor.deploy')}
              description={t('funcEditor.deployDescription')}
            />
          </div>

          <div className="p-3 rounded-[var(--radius)]" style={{ background: 'rgba(143, 255, 208, 0.03)', border: '1px solid rgba(143, 255, 208, 0.15)' }}>
            <p className="text-xs text-[var(--text-dim)] leading-relaxed">
              <span className="text-[var(--status-ok)] font-medium">{t('funcEditor.step1Message')}</span>{' '}
              {t('funcEditor.saveDraftFirst')}
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="overflow-hidden" style={{ background: 'var(--panel)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)', boxShadow: 'var(--shadow-chamber)' }}>
      <CardHeader className="pb-2 pt-4 px-5">
        <CardTitle className="text-sm font-semibold text-[var(--text)] flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
          <Server className="w-4 h-4 text-[var(--status-ok)]" />
          {t('funcEditor.deployTarget')}
        </CardTitle>
        {linkedAppId ? (
          <p className="text-xs text-[var(--text-faint)] mt-1 font-normal leading-relaxed">
            {t('funcEditor.linkedToApp')}
          </p>
        ) : null}
      </CardHeader>
      <CardContent className="px-5 pb-4 space-y-4">
        {/* Mini workflow indicator */}
        <div className="flex items-center gap-2 text-xs">
          <span className="flex items-center gap-1.5 text-[var(--status-ok)]">
            <CheckCircle2 className="w-3.5 h-3.5" />
            {t('funcEditor.draftSaved')}
          </span>
          <ArrowRight className="w-3 h-3 text-[var(--text-faint)]" />
          <span className={`flex items-center gap-1.5 ${isBackendStep ? 'text-[var(--accent)]' : 'text-[var(--status-ok)]'}`}>
            {isBackendStep ? <Circle className="w-3.5 h-3.5 fill-current" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
            {t('funcEditor.selectBackendStatus')}
          </span>
          <ArrowRight className="w-3 h-3 text-[var(--text-faint)]" />
          <span className={`flex items-center gap-1.5 ${isDeployStep ? 'text-[var(--accent)]' : 'text-[var(--text-faint)]'}`}>
            {isDeployStep ? <Rocket className="w-3.5 h-3.5" /> : <Rocket className="w-3.5 h-3.5 opacity-30" />}
            {t('funcEditor.deploy')}
          </span>
        </div>

        {deployBackendsLoading ? (
          <div className="flex items-center gap-2 text-xs text-[var(--text-faint)]">
            <Loader2 className="w-3.5 h-3.5 animate-spin shrink-0" />
            {t('funcEditor.loadingBackends')}
          </div>
        ) : filteredDeployBackends.length === 0 ? (
          <div className="p-3 rounded-[var(--radius)]" style={{ background: 'rgba(232, 196, 104, 0.05)', border: '1px solid rgba(232, 196, 104, 0.2)' }}>
            <p className="text-xs text-[var(--text-faint)] leading-relaxed">
              {t('funcEditor.noBackendsFound')}
            </p>
            <ol className="text-xs text-[var(--text-dim)] mt-2 ml-4 list-decimal space-y-1">
              <li>{t('funcEditor.goToApps')} <Link to="/apps" className="text-[var(--accent)] hover:underline">{t('funcEditor.apps')}</Link></li>
              <li>{t('funcEditor.createNewAppStep')}</li>
              <li>{t('funcEditor.addDeploymentBackend')}</li>
              <li>{t('funcEditor.returnHereToDeploy')}</li>
            </ol>
          </div>
        ) : (
          <>
            <Label htmlFor="deploy-backend" className="text-xs text-[var(--text-faint)]">
              {t('funcEditor.backend')} {isDeployStep && <span className="text-[var(--status-ok)] ml-1">{t('funcEditor.readyToDeploy')}</span>}
            </Label>
            <Select value={selectedDeployBackendId} onValueChange={setDeployBackendId}>
              <SelectTrigger id="deploy-backend" className="h-9 text-xs">
                <SelectValue placeholder={t('funcEditor.selectBackendPlaceholder')} />
              </SelectTrigger>
              <SelectContent>
                {filteredDeployBackends.map((b) => (
                  <SelectItem key={b.id} value={b.id} className="text-xs">
                    {b.appName} · {b.provider} / {b.region}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FieldError message={errors.deployBackend} />

            {isDeployStep && !isDirty && (
              <p className="text-xs text-[var(--status-ok)] flex items-center gap-1.5">
                <CheckCircle2 className="w-3.5 h-3.5" />
                {t('funcEditor.readyToDeployMessage')}
              </p>
            )}
            {isDeployStep && isDirty && (
              <p className="text-xs text-[var(--status-pending)] flex items-center gap-1.5">
                <Circle className="w-3.5 h-3.5" />
                {t('funcEditor.saveBeforeDeploy')}
              </p>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
