import '@/styles/index.css';
import { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.tsx';
import './i18n';
import { initSentry } from './sentry';

// Use bundled monaco-editor (same origin) instead of CDN loader.js — matches CSP script-src 'self'.
loader.config({ monaco });

initSentry();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
