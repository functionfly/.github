import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Sparkles, Hash, Wand2, Zap } from 'lucide-react';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { FieldError, InfoTip, SectionCard } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

const MAX_DESCRIPTION = 500;

// Simple naming assistant - suggests names based on code patterns
function analyzeCodeForSuggestions(code: string): string[] {
  const suggestions: string[] = [];

  // Check for common patterns
  if (code.includes('webhook') || code.includes('Webhook')) {
    suggestions.push('webhook-handler', 'process-webhook');
  }
  if (code.includes('proxy') || code.includes('Proxy') || code.includes('forward')) {
    suggestions.push('api-proxy', 'request-proxy');
  }
  if (code.includes('auth') || code.includes('jwt') || code.includes('JWT')) {
    suggestions.push('auth-middleware', 'verify-jwt');
  }
  if (code.includes('email') || code.includes('mail') || code.includes('send')) {
    suggestions.push('send-email', 'email-notifier');
  }
  if (code.includes('scheduled') || code.includes('cron') || code.includes('Cron')) {
    suggestions.push('scheduled-task', 'cron-job');
  }
  if (code.includes('slack') || code.includes('bot') || code.includes('discord')) {
    suggestions.push('slack-bot', 'chat-notifier');
  }
  if (code.includes('transform') || code.includes('map') || code.includes('filter')) {
    suggestions.push('data-transform', 'json-processor');
  }
  if (code.includes('database') || code.includes('db') || code.includes('sql')) {
    suggestions.push('db-query', 'data-access');
  }
  if (code.includes('cache') || code.includes('redis')) {
    suggestions.push('cache-handler', 'redis-proxy');
  }
  if (code.includes('upload') || code.includes('file') || code.includes('s3')) {
    suggestions.push('file-upload', 'asset-handler');
  }
  if (code.includes('payment') || code.includes('stripe') || code.includes('billing')) {
    suggestions.push('payment-webhook', 'process-payment');
  }

  // Generic fallback suggestions based on HTTP methods
  if (code.includes('GET') && code.includes('POST')) {
    suggestions.push('crud-api', 'rest-handler');
  } else if (suggestions.length === 0) {
    suggestions.push('http-handler', 'api-endpoint', 'request-processor');
  }

  return [...new Set(suggestions)].slice(0, 3);
}

export function BasicInfoSection({ editor }: Props) {
  const { t } = useTranslation();
  const {
    functionName,
    slug,
    description,
    setDescription,
    code,
    errors,
    handleNameChange,
    handleSlugChange,
    markDirty,
    setFunctionName,
  } = editor;

  const suggestions = useMemo(() => analyzeCodeForSuggestions(code), [code]);
  const showSuggestions = !functionName && suggestions.length > 0;

  const handleSuggestionClick = (suggestion: string) => {
    setFunctionName(suggestion);
    handleNameChange(suggestion);
  };

  return (
    <SectionCard icon={<Zap className="w-4 h-4" />} title={t('funcEditor.functionBasics')} step={1}>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <Label htmlFor="fn-name" className="text-xs font-medium text-text-secondary mb-1.5 block">
            {t('funcEditor.name')} <span className="text-red-400">*</span>
          </Label>
          <Input
            id="fn-name"
            placeholder="my-awesome-function"
            value={functionName}
            onChange={(e) => handleNameChange(e.target.value)}
            className="input"
            aria-describedby={errors.name ? 'fn-name-error' : undefined}
            autoComplete="off"
          />
          {/* Naming Assistant */}
          {showSuggestions && (
            <div className="mt-2">
              <div className="flex items-center gap-1.5 text-xs text-[#FF6B35] mb-1.5">
                <Wand2 className="w-3 h-3" />
                <span>{t('funcEditor.suggestedNames')}</span>
              </div>
              <div className="flex flex-wrap gap-1.5">
                {suggestions.map((suggestion) => (
                  <button
                    key={suggestion}
                    onClick={() => handleSuggestionClick(suggestion)}
                    className="px-2 py-1 text-xs rounded-md bg-bg-tertiary text-text-secondary hover:text-[#FF6B35] hover:border-[#FF6B35]/30 border border-border-subtle/50 transition-colors"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>
          )}
          <FieldError message={errors.name} />
        </div>
        <div>
          <Label htmlFor="fn-slug" className="text-xs font-medium text-text-secondary mb-1.5 block">
            {t('funcEditor.slugIdentifier')}
            <InfoTip content={t('funcEditor.slugInfoTip')} />
          </Label>
          <div className="input-with-icon">
            <Hash className="icon h-3.5 w-3.5 shrink-0" aria-hidden />
            <Input
              id="fn-slug"
              placeholder="my-awesome-function"
              value={slug}
              onChange={(e) => handleSlugChange(e.target.value)}
              className="input w-full font-mono text-sm"
              autoComplete="off"
            />
          </div>
          <FieldError message={errors.slug} />
        </div>
      </div>
      <div>
        <div className="flex items-center justify-between mb-1.5">
          <Label htmlFor="fn-desc" className="text-xs font-medium text-text-secondary block">
            {t('funcEditor.description')}
          </Label>
          <span
            className={`text-xs tabular-nums ${
              description.length > MAX_DESCRIPTION * 0.9
                ? description.length >= MAX_DESCRIPTION
                  ? 'text-red-400'
                  : 'text-amber-400'
                : 'text-text-muted'
            }`}
          >
            {description.length}/{MAX_DESCRIPTION}
          </span>
        </div>
        <Textarea
          id="fn-desc"
          placeholder={t('funcEditor.descriptionPlaceholder')}
          value={description}
          onChange={(e) => {
            if (e.target.value.length <= MAX_DESCRIPTION) {
              setDescription(e.target.value);
              markDirty();
            }
          }}
          className="input resize-none"
          rows={2}
          maxLength={MAX_DESCRIPTION}
        />
      </div>
    </SectionCard>
  );
}
