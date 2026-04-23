import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { Play, BookOpen } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { usePlaygroundStore } from '../store/playgroundStore';

interface ExampleSelectorProps {
  className?: string;
}

export function ExampleSelector({ className }: ExampleSelectorProps) {
  const { t } = useTranslation();
  const { functionInfo, setInputValue, setInputJson } = usePlaygroundStore();

  const examples = functionInfo?.manifest?.examples || [];

  // Also check for a single example in the input spec
  const inputExample = functionInfo?.manifest?.input?.example;
  const allExamples =
    examples.length > 0
      ? examples
      : inputExample !== undefined
      ? [{ name: t('playground.defaultExample'), input: inputExample, description: t('playground.defaultExampleDescription') }]
      : [];

  if (allExamples.length === 0) {
    return (
      <div className={cn('flex flex-col items-center justify-center py-12 text-center', className)}>
        <BookOpen className="w-10 h-10 text-text-muted mb-3" />
        <p className="text-sm text-text-muted">{t('playground.noExamplesAvailable')}</p>
        <p className="text-xs text-text-muted mt-1">
          {t('playground.addExamplesHint')}
        </p>
      </div>
    );
  }

  const handleLoadExample = (input: unknown) => {
    setInputValue(input);
    setInputJson(JSON.stringify(input, null, 2));
  };

  return (
    <div className={cn('space-y-3', className)}>
      <p className="text-xs text-text-muted">
        {t('playground.clickExampleToLoad')}
      </p>

      <div className="grid gap-3">
        {allExamples.map((example, i) => (
          <motion.div
            key={i}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2, delay: i * 0.05 }}
            className="group border border-border-subtle rounded-lg p-3 hover:border-indigo-500/50 hover:bg-indigo-500/5 transition-all cursor-pointer"
            onClick={() => handleLoadExample(example.input)}
          >
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium text-text-primary">{example.name}</p>
                {example.description && (
                  <p className="text-xs text-text-muted mt-0.5">{example.description}</p>
                )}
                <pre className="text-xs font-mono text-text-secondary mt-2 bg-bg-tertiary rounded p-2 overflow-auto max-h-24">
                  {JSON.stringify(example.input, null, 2)}
                </pre>
              </div>
              <Button
                size="sm"
                variant="ghost"
                className="shrink-0 h-7 gap-1.5 text-xs opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={(e) => {
                  e.stopPropagation();
                  handleLoadExample(example.input);
                }}
              >
                <Play className="w-3 h-3" />
                {t('playground.load')}
              </Button>
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
