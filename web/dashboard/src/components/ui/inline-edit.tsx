import { useState, useCallback, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Check, X, Pencil, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface InlineEditProps {
  value: string;
  onSave: (value: string) => Promise<void> | void;
  onCancel?: () => void;
  placeholder?: string;
  multiline?: boolean;
  disabled?: boolean;
  validate?: (value: string) => string | null;
  renderDisplay?: (value: string) => React.ReactNode;
  className?: string;
  inputClassName?: string;
  saveOnBlur?: boolean;
  autoFocus?: boolean;
  maxLength?: number;
  loading?: boolean;
  showEditIcon?: boolean;
}

export function InlineEdit({
  value: initialValue,
  onSave,
  onCancel,
  placeholder = 'Enter value...',
  multiline = false,
  disabled = false,
  validate,
  renderDisplay,
  className,
  inputClassName,
  saveOnBlur = true,
  autoFocus = false,
  maxLength,
  loading = false,
  showEditIcon = true,
}: InlineEditProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [value, setValue] = useState(initialValue);
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement | HTMLTextAreaElement>(null);

  useEffect(() => {
    if (isEditing && autoFocus && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing, autoFocus]);

  const handleEdit = useCallback(() => {
    if (disabled) return;
    setIsEditing(true);
    setValue(initialValue);
    setError(null);
  }, [disabled, initialValue]);

  const handleCancel = useCallback(() => {
    setIsEditing(false);
    setValue(initialValue);
    setError(null);
    onCancel?.();
  }, [initialValue, onCancel]);

  const validateAndSave = useCallback(async () => {
    if (validate) {
      const validationError = validate(value);
      if (validationError) {
        setError(validationError);
        return;
      }
    }

    if (value === initialValue) {
      setIsEditing(false);
      return;
    }

    try {
      setIsSaving(true);
      await onSave(value);
      setIsEditing(false);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setIsSaving(false);
    }
  }, [value, initialValue, onSave, validate]);

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    switch (e.key) {
      case 'Enter':
        if (!multiline) {
          e.preventDefault();
          validateAndSave();
        }
        break;
      case 'Escape':
        e.preventDefault();
        handleCancel();
        break;
    }
  }, [multiline, validateAndSave, handleCancel]);

  const handleBlur = useCallback(() => {
    if (saveOnBlur && !isSaving) {
      // Small delay to allow button clicks to process
      setTimeout(() => {
        validateAndSave();
      }, 150);
    }
  }, [saveOnBlur, isSaving, validateAndSave]);

  if (isEditing) {
    const EditComponent = multiline ? Textarea : Input;
    
    return (
      <div className={cn('flex items-start gap-2', className)}>
        <div className="flex-1 min-w-0">
          <EditComponent
            ref={inputRef as any}
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
              if (error) setError(null);
            }}
            onKeyDown={handleKeyDown}
            onBlur={handleBlur}
            placeholder={placeholder}
            disabled={disabled || isSaving}
            maxLength={maxLength}
            className={cn(
              error && 'border-red-500 focus-visible:ring-red-500',
              inputClassName
            )}
            rows={multiline ? 3 : undefined}
          />
          {error && (
            <p className="text-xs text-red-500 mt-1">{error}</p>
          )}
        </div>
        <div className="flex gap-1">
          <Button
            size="icon"
            variant="ghost"
            onClick={handleCancel}
            disabled={isSaving}
            className="h-8 w-8 shrink-0"
            title="Cancel (Esc)"
          >
            <X className="h-4 w-4" />
          </Button>
          <Button
            size="icon"
            variant="ghost"
            onClick={validateAndSave}
            disabled={disabled || isSaving}
            className="h-8 w-8 shrink-0"
            title="Save (Enter)"
          >
            {isSaving ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Check className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>
    );
  }

  const displayValue = initialValue || placeholder;

  return (
    <div
      onClick={handleEdit}
      className={cn(
        'group flex items-center gap-2 cursor-pointer hover:bg-muted/50 rounded px-2 py-1 -mx-2 -my-1 transition-colors',
        disabled && 'cursor-not-allowed opacity-50',
        loading && 'animate-pulse',
        className
      )}
      role="button"
      tabIndex={disabled ? -1 : 0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleEdit();
        }
      }}
    >
      <div className="flex-1 min-w-0 overflow-hidden text-ellipsis whitespace-nowrap">
        {renderDisplay ? renderDisplay(initialValue) : (
          <span className={cn(!initialValue && 'text-muted-foreground italic')}>
            {displayValue}
          </span>
        )}
      </div>
      {showEditIcon && !disabled && (
        <Button
          size="icon"
          variant="ghost"
          className="h-6 w-6 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
          onClick={(e) => {
            e.stopPropagation();
            handleEdit();
          }}
        >
          <Pencil className="h-3.5 w-3.5" />
        </Button>
      )}
    </div>
  );
}

// Inline edit with confirmation dialog for critical changes
export interface InlineEditWithConfirmProps extends Omit<InlineEditProps, 'onSave'> {
  onSave: (value: string) => Promise<void> | void;
  confirmMessage?: string;
  confirmTitle?: string;
  requireConfirm?: boolean;
}

export function InlineEditWithConfirm({
  onSave,
  confirmMessage,
  confirmTitle = 'Confirm Changes',
  requireConfirm = false,
  ...props
}: InlineEditWithConfirmProps) {
  const [showConfirm, setShowConfirm] = useState(false);
  const [pendingValue, setPendingValue] = useState('');

  const handleSave = useCallback(async (value: string) => {
    if (requireConfirm) {
      setPendingValue(value);
      setShowConfirm(true);
      return;
    }
    await onSave(value);
  }, [onSave, requireConfirm]);

  const handleConfirm = useCallback(async () => {
    await onSave(pendingValue);
    setShowConfirm(false);
  }, [onSave, pendingValue]);

  return (
    <>
      <InlineEdit {...props} onSave={handleSave} />
      {/* Confirmation dialog would be imported and used here */}
    </>
  );
}

// Hook for managing multiple inline edits
interface UseInlineEditGroupOptions {
  onBatchSave?: (changes: Record<string, string>) => Promise<void> | void;
}

export function useInlineEditGroup(options: UseInlineEditGroupOptions = {}) {
  const [pendingChanges, setPendingChanges] = useState<Record<string, string>>({});
  const [hasChanges, setHasChanges] = useState(false);

  const registerChange = useCallback((field: string, value: string) => {
    setPendingChanges(prev => ({
      ...prev,
      [field]: value,
    }));
    setHasChanges(true);
  }, []);

  const clearChanges = useCallback(() => {
    setPendingChanges({});
    setHasChanges(false);
  }, []);

  const saveChanges = useCallback(async () => {
    if (options.onBatchSave) {
      await options.onBatchSave(pendingChanges);
    }
    clearChanges();
  }, [pendingChanges, options, clearChanges]);

  return {
    pendingChanges,
    hasChanges,
    registerChange,
    clearChanges,
    saveChanges,
  };
}
