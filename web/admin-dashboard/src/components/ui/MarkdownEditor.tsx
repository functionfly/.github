/**
 * Markdown Editor Component
 *
 * Replaces the TipTap-based RichTextEditor with a simpler Markdown editor.
 * Markdown is easier to store, version-control, and render anywhere.
 *
 * Features:
 * - Live preview (side-by-side or tabbed)
 * - Toolbar with common formatting actions
 * - Dark mode support
 * - Better error handling (graceful fallback to plain text)
 */

import { logger } from '@/lib/monitoring/logger';
import MDEditor from '@uiw/react-md-editor';
import { Eye, Pencil, AlertCircle } from 'lucide-react';
import { useState } from 'react';

interface MarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  minHeight?: number;
}

export function MarkdownEditor({
  value,
  onChange,
  placeholder = 'Start writing your post in Markdown...',
  minHeight = 400,
}: MarkdownEditorProps) {
  const [mode, setMode] = useState<'edit' | 'preview' | 'split'>('edit');
  const [error, setError] = useState<string | null>(null);

  // Force light theme colors to avoid dark mode issues
  const colorMode: 'light' | 'dark' = 'light';

  const handleChange = (val?: string) => {
    try {
      setError(null);
      onChange(val ?? '');
    } catch (e: any) {
      logger.error('Markdown editor change failed', { error: e?.message });
      setError('Failed to update content');
    }
  };

  // Fallback content if value is empty
  const content = value || '';

  if (error) {
    return (
      <div className="border border-red-300 dark:border-red-700 rounded-lg p-4 bg-red-50 dark:bg-red-900/20">
        <div className="flex items-center gap-2 text-red-800 dark:text-red-200">
          <AlertCircle className="w-4 h-4" />
          <p className="text-sm font-medium">Editor error</p>
        </div>
        <p className="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
        <textarea
          value={content}
          onChange={(e) => handleChange(e.target.value)}
          placeholder={placeholder}
          className="w-full mt-3 p-3 border border-red-200 dark:border-red-800 rounded-md text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          style={{ minHeight: `${minHeight}px` }}
        />
      </div>
    );
  }

  // Single mode: edit only (or preview only)
  if (mode === 'edit' || mode === 'preview') {
    return (
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden bg-white dark:bg-gray-800">
        <div className="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-3 py-2">
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setMode('edit')}
              className={`flex items-center gap-1 px-2 py-1 text-sm rounded ${
                mode === 'edit'
                  ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
                  : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
              }`}
            >
              <Pencil className="w-3.5 h-3.5" />
              Edit
            </button>
            <button
              type="button"
              onClick={() => setMode('preview')}
              className={`flex items-center gap-1 px-2 py-1 text-sm rounded ${
                mode === 'preview'
                  ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300'
                  : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'
              }`}
            >
              <Eye className="w-3.5 h-3.5" />
              Preview
            </button>
          </div>
          <span className="text-xs text-gray-500 dark:text-gray-400">Markdown</span>
        </div>
        <MDEditor
          value={content}
          onChange={handleChange}
          height={minHeight}
          preview={mode === 'preview' ? 'preview' : 'edit'}
          visibleDragbar={false}
          data-color-mode={colorMode}
          textareaProps={{
            placeholder,
            className: '!bg-white dark:!bg-gray-800 !text-gray-900 dark:!text-gray-100',
          }}
        />
      </div>
    );
  }

  // Split view: edit + preview side by side
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden bg-white dark:bg-gray-800">
      <div className="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-3 py-2">
        <span className="text-xs text-gray-500 dark:text-gray-400">Markdown — Edit & Preview</span>
      </div>
      <MDEditor
        value={content}
        onChange={handleChange}
        height={minHeight}
        preview="live"
        visibleDragbar={true}
        data-color-mode={colorMode}
        textareaProps={{
          placeholder,
          className: '!bg-white dark:!bg-gray-800 !text-gray-900 dark:!text-gray-100',
        }}
      />
    </div>
  );
}
