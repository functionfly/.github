import { useState } from 'react';
import { motion } from 'framer-motion';
import { CodeBlock } from '@/components/common/CodeBlock';
import { cn } from '@/lib/utils';
import { useTranslation } from 'react-i18next';
import { usePlaygroundStore } from '../store/playgroundStore';
import {
  generateSnippet,
  SNIPPET_LANGUAGES,
  SnippetLanguage,
} from '../utils/codeGenerators';

interface CodeSnippetGeneratorProps {
  className?: string;
}

export function CodeSnippetGenerator({ className }: CodeSnippetGeneratorProps) {
  const { t } = useTranslation();
  const [activeLanguage, setActiveLanguage] = useState<SnippetLanguage>('curl');
  const { functionInfo, inputValue } = usePlaygroundStore();

  if (!functionInfo) {
    return (
      <div className={cn('text-xs text-text-muted text-center py-4', className)}>
        {t('playground.noFunctionLoaded')}
      </div>
    );
  }

  const snippet = generateSnippet(activeLanguage, {
    author: functionInfo.author,
    name: functionInfo.name,
    input: inputValue,
  });

  const activeLang = SNIPPET_LANGUAGES.find((l) => l.id === activeLanguage);

  return (
    <div className={cn('space-y-3', className)}>
      {/* Language tabs */}
      <div className="flex flex-wrap gap-1">
        {SNIPPET_LANGUAGES.map((lang) => (
          <button
            key={lang.id}
            onClick={() => setActiveLanguage(lang.id)}
            className={cn(
              'px-2.5 py-1 text-xs rounded-md transition-colors',
              activeLanguage === lang.id
                ? 'bg-indigo-600 text-white'
                : 'bg-bg-tertiary text-text-secondary hover:text-text-primary hover:bg-bg-secondary'
            )}
          >
            {lang.label}
          </button>
        ))}
      </div>

      {/* Code block */}
      <motion.div
        key={activeLanguage}
        initial={{ opacity: 0, y: 4 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.15 }}
      >
        <CodeBlock
          code={snippet}
          language={activeLang?.syntaxLang || 'bash'}
          showLineNumbers={false}
          maxHeight="300px"
        />
      </motion.div>
    </div>
  );
}
