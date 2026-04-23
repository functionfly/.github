import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Globe, Lock, Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { InfoTip, SectionCard } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function VisibilitySection({ editor }: Props) {
  const { t } = useTranslation();
  const {
    visibility,
    setVisibility,
    tags,
    newTag,
    setNewTag,
    addTag,
    removeTag,
    markDirty,
    tagInputRef,
  } = editor;

  return (
    <SectionCard icon={<Globe className="w-4 h-4" />} title={t('funcEditor.visibilityTags')} step={7}>
      {/* Public/Private toggle */}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-text-primary flex items-center gap-2">
            {visibility === 'public' ? (
              <>
                <Globe className="w-4 h-4 text-emerald-400" />
                {t('funcEditor.public')}
              </>
            ) : (
              <>
                <Lock className="w-4 h-4 text-text-muted" />
                {t('funcEditor.private')}
              </>
            )}
          </p>
          <p className="text-xs text-text-muted mt-0.5">
            {visibility === 'public'
              ? t('funcEditor.publicDescription')
              : t('funcEditor.privateDescription')}
          </p>
        </div>
        <Switch
          checked={visibility === 'public'}
          onCheckedChange={(c) => {
            setVisibility(c ? 'public' : 'private');
            markDirty();
          }}
          aria-label="Toggle visibility"
        />
      </div>

      <Separator className="opacity-30" />

      {/* Tags */}
      <div>
        <Label className="text-xs text-text-secondary mb-2 block">
          {t('funcEditor.tags')}
          <InfoTip content={t('funcEditor.tagsInfoTip')} />
        </Label>
        {tags.length > 0 && (
          <div className="flex flex-wrap gap-1.5 mb-2">
            {tags.map((t) => (
              <Badge
                key={t}
                variant="secondary"
                className="gap-1 text-xs cursor-pointer hover:bg-red-500/20 hover:text-red-400 transition-colors"
                onClick={() => removeTag(t)}
                role="button"
                aria-label={`Remove tag ${t}`}
              >
                {t}
                <X className="w-2.5 h-2.5" />
              </Badge>
            ))}
          </div>
        )}
        <div className="flex gap-2">
          <Input
            ref={tagInputRef}
            placeholder={t('funcEditor.addTagPlaceholder')}
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ',') {
                e.preventDefault();
                addTag();
              }
            }}
            className="input text-sm"
          />
          <Button size="sm" variant="outline" onClick={addTag} disabled={!newTag.trim()}>
            <Plus className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>
    </SectionCard>
  );
}
