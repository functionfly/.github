import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Hash, Zap } from 'lucide-react';
import { FieldError, InfoTip, SectionCard } from '../components/editor-ui';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

const MAX_DESCRIPTION = 500;

export function BasicInfoSection({ editor }: Props) {
  const {
    functionName,
    slug,
    description,
    setDescription,
    errors,
    handleNameChange,
    handleSlugChange,
    markDirty,
  } = editor;

  return (
    <SectionCard icon={<Zap className="w-4 h-4" />} title="Function Basics" step={1}>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <Label htmlFor="fn-name" className="text-xs font-medium text-text-secondary mb-1.5 block">
            Name <span className="text-red-400">*</span>
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
          <FieldError message={errors.name} />
        </div>
        <div>
          <Label htmlFor="fn-slug" className="text-xs font-medium text-text-secondary mb-1.5 block">
            Slug / Identifier
            <InfoTip content="URL-safe identifier used in API calls. Auto-generated from name." />
          </Label>
          <div className="relative">
            <Hash className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted" />
            <Input
              id="fn-slug"
              placeholder="my-awesome-function"
              value={slug}
              onChange={(e) => handleSlugChange(e.target.value)}
              className="input pl-8 font-mono text-sm"
              autoComplete="off"
            />
          </div>
          <FieldError message={errors.slug} />
        </div>
      </div>
      <div>
        <div className="flex items-center justify-between mb-1.5">
          <Label htmlFor="fn-desc" className="text-xs font-medium text-text-secondary block">
            Description
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
          placeholder="What does this function do?"
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
