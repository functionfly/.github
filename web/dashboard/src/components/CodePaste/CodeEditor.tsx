import { useRef, useEffect } from 'react';
import { LazyMonacoEditor, type OnMount, type OnChange } from '@/components/LazyMonacoEditor';
import type { editor } from 'monaco-editor';
import { SUPPORTED_LANGUAGES, SupportedLanguage } from '../../types/codePaste';

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string | null;
  readOnly?: boolean;
  height?: string | number;
  placeholder?: string;
}

const languageToMonaco: Record<string, string> = {
  python: 'python',
  javascript: 'javascript',
  typescript: 'typescript',
  go: 'go',
  rust: 'rust',
  ruby: 'ruby',
  java: 'java',
  kotlin: 'kotlin',
  swift: 'swift',
  cpp: 'cpp',
  c: 'c',
};

export function CodeEditor({
  value,
  onChange,
  language,
  readOnly = false,
  height = '400px',
  placeholder = 'Paste or type your code here...',
}: CodeEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);

  const handleEditorDidMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;

    editor.updateOptions({
      wordWrap: 'on',
      minimap: { enabled: true },
      scrollBeyondLastLine: false,
      fontSize: 14,
      lineNumbers: 'on',
      renderWhitespace: 'selection' as any,
      tabSize: 4,
      insertSpaces: true,
      autoClosingBrackets: 'always' as any,
      autoClosingQuotes: 'always' as any,
      formatOnPaste: true,
      formatOnType: false,
    });
  };

  const handleChange: OnChange = (newValue) => {
    onChange(newValue || '');
  };

  const getLanguage = (): string => {
    if (!language) return 'plaintext';
    if (language === 'auto' || language === 'unknown') return 'plaintext';
    return languageToMonaco[language] || 'plaintext';
  };

  const getPlaceholder = (): string => {
    if (value && value.length > 0) return '';
    return placeholder;
  };

  return (
    <div className="code-editor-wrapper" style={{ height, position: 'relative' }}>
      <LazyMonacoEditor
        height={height}
        language={getLanguage()}
        value={value}
        onChange={handleChange}
        onMount={handleEditorDidMount}
        theme="vs-dark"
        options={{
          readOnly,
          wordWrap: 'on',
          minimap: { enabled: true },
          scrollBeyondLastLine: false,
          fontSize: 14,
          lineNumbers: 'on',
          renderWhitespace: 'selection' as any,
          tabSize: 4,
          insertSpaces: true,
          autoClosingBrackets: 'always' as any,
          autoClosingQuotes: 'always' as any,
          placeholder: getPlaceholder(),
        }}
        loading={
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100%',
            background: '#1e1e1e',
            color: '#888',
          }}>
            Loading editor...
          </div>
        }
      />
      {!value && (
        <div
          className="code-placeholder"
          style={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            color: '#666',
            fontSize: '16px',
            pointerEvents: 'none',
            zIndex: 1,
          }}
        >
          {placeholder}
        </div>
      )}
    </div>
  );
}