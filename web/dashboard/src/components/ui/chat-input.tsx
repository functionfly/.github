import * as React from 'react';
import DOMPurify from 'dompurify';
import { cn } from '@/lib/utils';
import {
  Paperclip,
  Send,
  Loader2,
  X,
  FileText,
  Image as ImageIcon,
  SmilePlus,
  Eye,
  EyeOff,
  Bold,
  Italic,
  Code2,
  AlertTriangle,
} from 'lucide-react';
import { Button } from './button';
import EmojiPicker, { Theme as EmojiTheme } from 'emoji-picker-react';
import { useThemeStore } from '@/stores/themeStore';

export interface ChatInputProps {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  onTyping?: (typing: boolean) => void;
  onFilesChange?: (files: File[]) => void;
  pending?: boolean;
  disabled?: boolean;
  placeholder?: string;
  maxFiles?: number;
  maxSize?: number;
  maxLength?: number;
  showMarkdownPreview?: boolean;
  rateLimitRemaining?: number;
  rateLimitResetAt?: string;
  className?: string;
}

function formatFileSize(bytes: number) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function sanitizeAndRenderMarkdown(text: string): string {
  const escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/_(.+?)_/g, '<em>$1</em>')
    .replace(/`(.+?)`/g, '<code class="px-1 py-0.5 rounded bg-muted text-xs font-mono">$1</code>')
    .replace(/\n/g, '<br />');

  return DOMPurify.sanitize(escaped, {
    ALLOWED_TAGS: ['strong', 'em', 'code', 'br'],
    ALLOWED_ATTR: ['class'],
  });
}

export function ChatInput({
  value,
  onChange,
  onSend,
  onTyping,
  onFilesChange,
  pending = false,
  disabled = false,
  placeholder = 'Type a message\u2026',
  maxFiles = 5,
  maxSize = 25 * 1024 * 1024,
  maxLength,
  showMarkdownPreview: enableMarkdownPreview = false,
  rateLimitRemaining,
  rateLimitResetAt,
  className,
}: ChatInputProps) {
  const textareaRef = React.useRef<HTMLTextAreaElement>(null);
  const fileInputRef = React.useRef<HTMLInputElement>(null);
  const typingTimeoutRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);
  const emojiPickerRef = React.useRef<HTMLDivElement>(null);
  const isTypingRef = React.useRef(false);
  const [attachedFiles, setAttachedFiles] = React.useState<File[]>([]);
  const [isDragging, setIsDragging] = React.useState(false);
  const [showEmojiPicker, setShowEmojiPicker] = React.useState(false);
  const [showPreview, setShowPreview] = React.useState(false);

  const resolvedTheme = useThemeStore((s) => s.resolvedTheme);
  const emojiTheme = resolvedTheme === 'dark' ? EmojiTheme.DARK : EmojiTheme.LIGHT;

  const showRateLimitWarning = rateLimitRemaining !== undefined && rateLimitRemaining < 3;

  // Auto-grow textarea
  React.useEffect(() => {
    const el = textareaRef.current;
    if (el) {
      el.style.height = 'auto';
      el.style.height = `${Math.min(el.scrollHeight, 160)}px`;
    }
  }, [value]);

  React.useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    };
  }, []);

  // Close emoji picker on outside click
  React.useEffect(() => {
    if (!showEmojiPicker) return;
    const handleClick = (e: MouseEvent) => {
      if (emojiPickerRef.current && !emojiPickerRef.current.contains(e.target as Node)) {
        setShowEmojiPicker(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [showEmojiPicker]);

  const emitTyping = (nextValue: string) => {
    if (!onTyping) return;
    if (nextValue.length > 0 && !isTypingRef.current) {
      isTypingRef.current = true;
      onTyping(true);
    } else if (nextValue.length === 0 && isTypingRef.current) {
      isTypingRef.current = false;
      onTyping(false);
    }
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    if (nextValue.length > 0) {
      typingTimeoutRef.current = setTimeout(() => {
        isTypingRef.current = false;
        onTyping(false);
      }, 3000);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const next = e.target.value;
    if (maxLength && next.length > maxLength) return;
    onChange(next);
    emitTyping(next);
  };

  const stopTyping = () => {
    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    isTypingRef.current = false;
    onTyping?.(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      stopTyping();
      onSend();
    }
  };

  const addFiles = (incoming: FileList | File[]) => {
    const remaining = maxFiles - attachedFiles.length;
    if (remaining <= 0) return;
    const valid = Array.from(incoming).slice(0, remaining).filter((f) => f.size <= maxSize);
    if (valid.length === 0) return;
    const next = [...attachedFiles, ...valid];
    setAttachedFiles(next);
    onFilesChange?.(next);
  };

  const removeFile = (index: number) => {
    const next = attachedFiles.filter((_, i) => i !== index);
    setAttachedFiles(next);
    onFilesChange?.(next);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    if (!disabled) setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (!disabled && e.dataTransfer.files) {
      addFiles(e.dataTransfer.files);
    }
  };

  const handleEmojiClick = (emojiData: { emoji: string }) => {
    const el = textareaRef.current;
    if (!el) return;
    const start = el.selectionStart;
    const end = el.selectionEnd;
    const next = value.slice(0, start) + emojiData.emoji + value.slice(end);
    if (maxLength && next.length > maxLength) return;
    onChange(next);
    // Restore cursor position after emoji insertion
    requestAnimationFrame(() => {
      el.selectionStart = el.selectionEnd = start + emojiData.emoji.length;
      el.focus();
    });
    setShowEmojiPicker(false);
  };

  const insertMarkdown = (before: string, after: string) => {
    const el = textareaRef.current;
    if (!el) return;
    const start = el.selectionStart;
    const end = el.selectionEnd;
    const selected = value.slice(start, end);
    const next = value.slice(0, start) + before + selected + after + value.slice(end);
    if (maxLength && next.length > maxLength) return;
    onChange(next);
    requestAnimationFrame(() => {
      el.selectionStart = start + before.length;
      el.selectionEnd = start + before.length + selected.length;
      el.focus();
    });
  };

  const fileIcon = (type: string) => {
    if (type.startsWith('image/')) return <ImageIcon className="h-3.5 w-3.5" />;
    return <FileText className="h-3.5 w-3.5" />;
  };

  return (
    <div
      className={cn('border-t border-border', className)}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      {isDragging && (
        <div className="flex items-center justify-center gap-2 p-3 bg-brand-500/10 border-b border-brand-500/30 text-sm text-brand-foreground">
          <Paperclip className="h-4 w-4" />
          Drop files here to attach
        </div>
      )}

      {attachedFiles.length > 0 && (
        <div className="flex flex-wrap gap-2 px-3 pt-3">
          {attachedFiles.map((file, i) => (
            <div
              key={`${file.name}-${i}`}
              className="flex items-center gap-1.5 rounded-md border border-border bg-muted/50 px-2 py-1 text-xs"
            >
              {fileIcon(file.type)}
              <span className="truncate max-w-[120px]">{file.name}</span>
              <span className="text-muted-foreground">{formatFileSize(file.size)}</span>
              <button
                onClick={() => removeFile(i)}
                className="text-muted-foreground hover:text-foreground ml-0.5"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
          ))}
        </div>
      )}

      {showPreview && value.trim() && (
        <div className="mx-3 mt-2 rounded-md border border-border bg-muted/30 p-3 text-sm prose prose-sm dark:prose-invert max-w-none">
          <div
            className="whitespace-pre-wrap"
            dangerouslySetInnerHTML={{ __html: sanitizeAndRenderMarkdown(value) }}
          />
        </div>
      )}

      <div className="flex items-end gap-1.5 p-3">
        <input
          ref={fileInputRef}
          type="file"
          multiple
          onChange={(e) => e.target.files && addFiles(e.target.files)}
          className="hidden"
        />
        <Button
          variant="ghost"
          size="icon"
          className="shrink-0 h-8 w-8"
          onClick={() => fileInputRef.current?.click()}
          disabled={disabled || attachedFiles.length >= maxFiles}
          title="Attach files"
        >
          <Paperclip className="h-4 w-4" />
        </Button>

        <div className="flex-1 relative">
          <textarea
            ref={textareaRef}
            value={value}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            className={cn(
              'w-full resize-none bg-transparent py-2.5 text-sm outline-none',
              'placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50',
              'min-h-[40px] max-h-[160px]',
            )}
          />
        </div>

        {enableMarkdownPreview && (
          <Button
            variant="ghost"
            size="icon"
            className={cn('shrink-0 h-8 w-8', showPreview && 'text-brand-500')}
            onClick={() => setShowPreview(!showPreview)}
            disabled={disabled}
            title={showPreview ? 'Hide preview' : 'Markdown preview'}
          >
            {showPreview ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </Button>
        )}

        <div className="flex items-center gap-0.5 shrink-0">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => insertMarkdown('**', '**')}
            disabled={disabled}
            title="Bold"
          >
            <Bold className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => insertMarkdown('_', '_')}
            disabled={disabled}
            title="Italic"
          >
            <Italic className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => insertMarkdown('`', '`')}
            disabled={disabled}
            title="Inline code"
          >
            <Code2 className="h-3.5 w-3.5" />
          </Button>
        </div>

        <div className="relative" ref={emojiPickerRef}>
          <Button
            variant="ghost"
            size="icon"
            className={cn('shrink-0 h-8 w-8', showEmojiPicker && 'text-brand-500')}
            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
            disabled={disabled}
            title="Emoji"
          >
            <SmilePlus className="h-4 w-4" />
          </Button>
          {showEmojiPicker && (
            <div className="absolute bottom-full right-0 mb-2 z-50 shadow-xl rounded-lg overflow-hidden">
              <EmojiPicker
                onEmojiClick={handleEmojiClick}
                theme={emojiTheme}
                width={320}
                height={380}
                lazyLoadEmojis
              />
            </div>
          )}
        </div>

        <Button
          size="icon"
          className="shrink-0 h-8 w-8"
          onClick={() => {
            stopTyping();
            onSend();
          }}
          disabled={disabled || (!value.trim() && attachedFiles.length === 0) || pending}
        >
          {pending ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Send className="h-4 w-4" />
          )}
        </Button>
      </div>

      {maxLength && (
        <div className="px-3 pb-1.5 flex justify-end">
          <span
            className={cn(
              'text-[10px] tabular-nums',
              value.length > maxLength * 0.9
                ? value.length >= maxLength
                  ? 'text-red-500 font-medium'
                  : 'text-amber-500'
                : 'text-muted-foreground',
            )}
          >
            {value.length}/{maxLength}
          </span>
        </div>
      )}

      {showRateLimitWarning && (
        <div className="px-3 pb-2 flex items-center gap-1.5 text-amber-500">
          <AlertTriangle className="h-3 w-3" />
          <span className="text-xs">
            Rate limit low ({rateLimitRemaining} remaining)
            {rateLimitResetAt && ` - resets ${new Date(rateLimitResetAt).toLocaleTimeString()}`}
          </span>
        </div>
      )}
    </div>
  );
}
