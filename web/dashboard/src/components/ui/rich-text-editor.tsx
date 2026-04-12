'use client';

import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import DOMPurify from 'dompurify';
import { Bold, Italic, List, ListOrdered, Link, Unlink, Heading1, Heading2, Quote, Code, Undo, Redo } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './button';

const richTextEditorVariants = cva(
  'relative overflow-hidden rounded-lg border bg-bg-primary',
  {
    variants: {
      variant: {
        default: 'border-border-default',
        ghost: 'border-transparent bg-transparent',
        card: 'border-border-default shadow-sm',
      },
      size: {
        default: 'min-h-[200px]',
        sm: 'min-h-[120px]',
        lg: 'min-h-[400px]',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

export interface RichTextEditorProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'onChange'>,
    VariantProps<typeof richTextEditorVariants> {
  value?: string;
  onChange?: (value: string) => void;
  onBlur?: () => void;
  placeholder?: string;
  disabled?: boolean;
  readonly?: boolean;
  maxLength?: number;
  showToolbar?: boolean;
  showCharacterCount?: boolean;
  sanitize?: boolean;
  allowedTags?: string[];
}

interface EditorCommand {
  icon: React.ReactNode;
  command: string;
  label: string;
  shortcut?: string;
}

const defaultCommands: EditorCommand[] = [
  { icon: <Bold className="h-4 w-4" />, command: 'bold', label: 'Bold', shortcut: 'Ctrl+B' },
  { icon: <Italic className="h-4 w-4" />, command: 'italic', label: 'Italic', shortcut: 'Ctrl+I' },
  { icon: <Heading1 className="h-4 w-4" />, command: 'formatBlock', label: 'Heading 1' },
  { icon: <Heading2 className="h-4 w-4" />, command: 'formatBlock', label: 'Heading 2' },
  { icon: <List className="h-4 w-4" />, command: 'insertUnorderedList', label: 'Bullet List' },
  { icon: <ListOrdered className="h-4 w-4" />, command: 'insertOrderedList', label: 'Numbered List' },
  { icon: <Quote className="h-4 w-4" />, command: 'formatBlock', label: 'Quote' },
  { icon: <Code className="h-4 w-4" />, command: 'formatBlock', label: 'Code Block' },
  { icon: <Link className="h-4 w-4" />, command: 'createLink', label: 'Link' },
];

const historyCommands: EditorCommand[] = [
  { icon: <Undo className="h-4 w-4" />, command: 'undo', label: 'Undo', shortcut: 'Ctrl+Z' },
  { icon: <Redo className="h-4 w-4" />, command: 'redo', label: 'Redo', shortcut: 'Ctrl+Y' },
];

const RichTextEditor = React.forwardRef<HTMLDivElement, RichTextEditorProps>(
  (
    {
      className,
      variant,
      size,
      value,
      onChange,
      onBlur,
      placeholder = 'Start typing...',
      disabled = false,
      readonly = false,
      maxLength,
      showToolbar = true,
      showCharacterCount = false,
      sanitize = true,
      allowedTags = ['p', 'br', 'strong', 'b', 'em', 'i', 'u', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'ul', 'ol', 'li', 'blockquote', 'pre', 'code', 'a'],
      ...props
    },
    ref
  ) => {
    const editorRef = React.useRef<HTMLDivElement>(null);
    const [internalValue, setInternalValue] = React.useState(value || '');
    const [activeCommands, setActiveCommands] = React.useState<Set<string>>(new Set());
    const [showLinkDialog, setShowLinkDialog] = React.useState(false);
    const [linkUrl, setLinkUrl] = React.useState('');

    const currentValue = value !== undefined ? value : internalValue;

    const sanitizeHtml = React.useCallback(
      (html: string) => {
        if (!sanitize) return html;
        return DOMPurify.sanitize(html, {
          ALLOWED_TAGS: allowedTags,
          ALLOWED_ATTR: ['href', 'target', 'rel', 'class'],
          ALLOW_DATA_ATTR: false,
        });
      },
      [sanitize, allowedTags]
    );

    const updateValue = React.useCallback(
      (html: string) => {
        const cleanHtml = sanitizeHtml(html);
        if (maxLength && cleanHtml.length > maxLength) {
          return;
        }
        setInternalValue(cleanHtml);
        onChange?.(cleanHtml);
      },
      [sanitizeHtml, maxLength, onChange]
    );

    const handleInput = React.useCallback(() => {
      if (editorRef.current) {
        updateValue(editorRef.current.innerHTML);
      }
    }, [updateValue]);

    const execCommand = React.useCallback(
      (command: string, value?: string) => {
        if (disabled || readonly) return;
        document.execCommand(command, false, value);
        handleInput();
        editorRef.current?.focus();
      },
      [disabled, readonly, handleInput]
    );

    const handleToolbarAction = React.useCallback(
      (command: string, label: string) => {
        if (command === 'formatBlock') {
          const tag = label === 'Heading 1' ? 'H1' : label === 'Heading 2' ? 'H2' : label === 'Quote' ? 'BLOCKQUOTE' : label === 'Code Block' ? 'PRE' : 'P';
          execCommand(command, tag);
        } else if (command === 'createLink') {
          setShowLinkDialog(true);
        } else {
          execCommand(command);
        }
      },
      [execCommand]
    );

    const createLink = React.useCallback(() => {
      if (linkUrl) {
        execCommand('createLink', linkUrl);
        setLinkUrl('');
        setShowLinkDialog(false);
      }
    }, [linkUrl, execCommand]);

    const unlink = React.useCallback(() => {
      execCommand('unlink');
    }, [execCommand]);

    const updateActiveCommands = React.useCallback(() => {
      const commands = new Set<string>();
      if (document.queryCommandState('bold')) commands.add('bold');
      if (document.queryCommandState('italic')) commands.add('italic');
      if (document.queryCommandState('underline')) commands.add('underline');
      if (document.queryCommandState('insertUnorderedList')) commands.add('insertUnorderedList');
      if (document.queryCommandState('insertOrderedList')) commands.add('insertOrderedList');
      setActiveCommands(commands);
    }, []);

    React.useEffect(() => {
      const handleSelectionChange = () => {
        updateActiveCommands();
      };
      document.addEventListener('selectionchange', handleSelectionChange);
      return () => document.removeEventListener('selectionchange', handleSelectionChange);
    }, [updateActiveCommands]);

    React.useEffect(() => {
      if (editorRef.current && value !== undefined && value !== editorRef.current.innerHTML) {
        editorRef.current.innerHTML = value;
      }
    }, [value]);

    return (
      <div
        ref={ref}
        className={cn(richTextEditorVariants({ variant, size, className }))}
        {...props}
      >
        {showToolbar && !readonly && (
          <div className="flex flex-wrap items-center gap-1 border-b border-border-default bg-bg-secondary/50 p-2">
            {defaultCommands.map((cmd) => (
              <Button
                key={cmd.label}
                variant="ghost"
                size="icon"
                className={cn(
                  'h-8 w-8',
                  activeCommands.has(cmd.command) && 'bg-brand-500/10 text-brand-500'
                )}
                onClick={() => handleToolbarAction(cmd.command, cmd.label)}
                disabled={disabled}
                title={`${cmd.label}${cmd.shortcut ? ` (${cmd.shortcut})` : ''}`}
              >
                {cmd.icon}
              </Button>
            ))}
            <div className="mx-2 h-4 w-px bg-border-default" />
            {historyCommands.map((cmd) => (
              <Button
                key={cmd.label}
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => execCommand(cmd.command)}
                disabled={disabled}
                title={`${cmd.label} (${cmd.shortcut})`}
              >
                {cmd.icon}
              </Button>
            ))}
            <div className="mx-2 h-4 w-px bg-border-default" />
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={unlink}
              disabled={disabled}
              title="Remove Link"
            >
              <Unlink className="h-4 w-4" />
            </Button>
          </div>
        )}

        {showLinkDialog && (
          <div className="absolute inset-0 z-50 flex items-center justify-center bg-bg-primary/80 backdrop-blur-sm">
            <div className="rounded-lg border border-border-default bg-card p-4 shadow-lg">
              <input
                type="url"
                value={linkUrl}
                onChange={(e) => setLinkUrl(e.target.value)}
                placeholder="Enter URL..."
                className="mb-3 w-64 rounded-md border border-border-default bg-bg-primary px-3 py-2 text-sm text-text-primary"
                autoFocus
              />
              <div className="flex justify-end gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setShowLinkDialog(false);
                    setLinkUrl('');
                  }}
                >
                  Cancel
                </Button>
                <Button size="sm" onClick={createLink} disabled={!linkUrl}>
                  Insert
                </Button>
              </div>
            </div>
          </div>
        )}

        <div
          ref={editorRef}
          className={cn(
            'min-h-[inherit] p-4 text-text-primary outline-none',
            disabled && 'cursor-not-allowed opacity-50',
            readonly && 'cursor-default'
          )}
          contentEditable={!disabled && !readonly}
          onInput={handleInput}
          onBlur={onBlur}
          data-placeholder={placeholder}
          suppressContentEditableWarning
          style={{
            minHeight: 'inherit',
          }}
        />

        {showCharacterCount && (
          <div className="flex items-center justify-end border-t border-border-default bg-bg-secondary/50 px-3 py-1.5">
            <span
              className={cn(
                'text-xs',
                maxLength && currentValue.length > maxLength * 0.9
                  ? 'text-error'
                  : 'text-text-muted'
              )}
            >
              {currentValue.length}
              {maxLength && `/${maxLength}`} characters
            </span>
          </div>
        )}
      </div>
    );
  }
);
RichTextEditor.displayName = 'RichTextEditor';

/**
 * RichTextViewer - Read-only display of sanitized HTML content
 * Use this for displaying rich text without editing capabilities
 */
export interface RichTextViewerProps extends React.HTMLAttributes<HTMLDivElement> {
  content: string;
  sanitize?: boolean;
  allowedTags?: string[];
}

const RichTextViewer = React.forwardRef<HTMLDivElement, RichTextViewerProps>(
  ({ content, sanitize = true, allowedTags, className, ...props }, ref) => {
    const sanitizedContent = React.useMemo(() => {
      if (!sanitize) return content;
      return DOMPurify.sanitize(content, {
        ALLOWED_TAGS: allowedTags || ['p', 'br', 'strong', 'b', 'em', 'i', 'u', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'ul', 'ol', 'li', 'blockquote', 'pre', 'code', 'a'],
        ALLOWED_ATTR: ['href', 'target', 'rel', 'class'],
        ALLOW_DATA_ATTR: false,
      });
    }, [content, sanitize, allowedTags]);

    return (
      <div
        ref={ref}
        className={cn('prose prose-sm max-w-none dark:prose-invert', className)}
        dangerouslySetInnerHTML={{ __html: sanitizedContent }}
        {...props}
      />
    );
  }
);
RichTextViewer.displayName = 'RichTextViewer';

export { RichTextEditor, RichTextViewer, richTextEditorVariants };
