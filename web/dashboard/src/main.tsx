import '@/styles/index.css';
import { enableMapSet } from 'immer';
import { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.tsx';
import './i18n';
import { initSentry } from './sentry';

enableMapSet();

// Use bundled monaco-editor (same origin) instead of CDN loader.js — matches CSP script-src 'self'.
// @ts-expect-error - monaco-editor and @monaco-editor/react have versioned type peers
loader.config({ monaco });

if (typeof self !== 'undefined') {
  self.MonacoEnvironment = {
    getWorker(moduleId: string, label: string) {
      if (label === 'json') {
        return new Worker(
          new URL('monaco-editor/esm/vs/language/json/json.worker.js', import.meta.url),
          { type: 'module' }
        );
      }
      if (label === 'typescript' || label === 'javascript') {
        return new Worker(
          new URL('monaco-editor/esm/vs/language/typescript/ts.worker.js', import.meta.url),
          { type: 'module' }
        );
      }
      return new Worker(
        new URL('monaco-editor/esm/vs/editor/editor.worker.js', import.meta.url),
        { type: 'module' }
      );
    },
  };
}

initSentry();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
