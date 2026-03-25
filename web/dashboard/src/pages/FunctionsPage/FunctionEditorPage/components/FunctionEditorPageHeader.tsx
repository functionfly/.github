import type { FunctionEditorModel } from '../useFunctionEditor';
import { Sparkles } from 'lucide-react';

type Props = { editor: FunctionEditorModel };

export function FunctionEditorPageHeader({ editor }: Props) {
  const { isEditing, functionName } = editor;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 pt-8 pb-6">
      <div className="flex items-start gap-4">
        <div
          className="w-12 h-12 rounded-xl flex items-center justify-center shrink-0"
          style={{ background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)' }}
        >
          <Sparkles className="w-6 h-6 text-white" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-text-primary">
            {isEditing ? `Edit ${functionName || 'Function'}` : functionName || 'New Function'}
          </h1>
          <p className="text-sm text-text-secondary mt-1">
            {isEditing
              ? 'Edit and redeploy your edge function.'
              : 'Create and deploy a serverless edge function.'}
            <span className="ml-2 text-text-muted text-xs">
              Press{' '}
              <kbd className="px-1 py-0.5 rounded text-xs font-mono bg-bg-tertiary border border-border-subtle">
                ⌘S
              </kbd>{' '}
              to save draft
            </span>
          </p>
        </div>
      </div>
    </div>
  );
}
