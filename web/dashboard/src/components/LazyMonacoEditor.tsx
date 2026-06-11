import { lazy, Suspense } from 'react';
import type { OnMount, OnChange } from '@monaco-editor/react';

const MonacoEditor = lazy(() => import('@monaco-editor/react').then(m => ({ default: m.default })));

interface LazyMonacoEditorProps {
  height?: string | number;
  language?: string;
  value?: string;
  onChange?: OnChange;
  onMount?: OnMount;
  theme?: string;
  options?: Record<string, unknown>;
  loading?: React.ReactNode;
}

function LoadingFallback({ height = '400px' }: { height?: string | number }) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: typeof height === 'number' ? `${height}px` : height,
        background: '#1e1e1e',
        color: '#888',
        borderRadius: '4px',
      }}
    >
      Loading editor...
    </div>
  );
}

export function LazyMonacoEditor({
  height = '400px',
  language,
  value,
  onChange,
  onMount,
  theme = 'vs-dark',
  options,
  loading,
}: LazyMonacoEditorProps) {
  return (
    <Suspense fallback={loading || <LoadingFallback height={height} />}>
      <MonacoEditor
        height={height}
        language={language}
        value={value}
        onChange={onChange}
        onMount={onMount}
        theme={theme}
        options={options}
        loading={loading || <LoadingFallback height={height} />}
      />
    </Suspense>
  );
}

export type { OnMount, OnChange };