import { ReactNode } from 'react';
import { FieldError } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { CheckCircle, AlertCircle } from 'lucide-react';

interface FormFieldState {
  hasError: boolean;
  isTouched: boolean;
  isDirty: boolean;
  hasValue: boolean;
  isValid: boolean;
  showError: boolean;
}

interface BaseFormFieldProps {
  id: string;
  label: string;
  required?: boolean;
  fieldState: FormFieldState;
  error?: FieldError;
  children: ReactNode;
}

export function BaseFormField({ id, label, required, fieldState, error, children }: BaseFormFieldProps) {
  return (
    <div className="space-y-3 animate-fade-in">
      <Label htmlFor={id} className="flex items-center gap-2 text-text-primary font-medium">
        {label}
        {required && <span className="text-error">*</span>}
        {fieldState.isValid && (
          <CheckCircle className="h-4 w-4 text-success animate-pulse-glow" />
        )}
      </Label>
      <div className="relative group">
        {children}
        {fieldState.hasValue && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 transition-all duration-200">
            {fieldState.isValid ? (
              <CheckCircle className="h-4 w-4 text-success animate-bounce-subtle glow-success" />
            ) : fieldState.showError ? (
              <AlertCircle className="h-4 w-4 text-error animate-shake glow-error" />
            ) : null}
          </div>
        )}
      </div>
      {fieldState.showError && error?.message && (
        <p className="text-sm text-error flex items-center gap-2 animate-slide-up bg-error/10 px-3 py-2 rounded-md border border-error/20">
          <AlertCircle className="h-3 w-3 animate-pulse" />
          {error.message}
        </p>
      )}
    </div>
  );
}

interface TextFormFieldProps {
  id: string;
  label: string;
  required?: boolean;
  placeholder: string;
  fieldState: FormFieldState;
  error?: FieldError;
  register: any;
}

export function TextFormField({ id, label, required, placeholder, fieldState, error, register }: TextFormFieldProps) {
  return (
    <BaseFormField id={id} label={label} required={required} fieldState={fieldState} error={error}>
      <Input
        id={id}
        type="text"
        {...register}
        placeholder={placeholder}
        className={`backdrop-blur-sm border-2 transition-all duration-300 hover:border-brand-500 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 focus:glow-sm text-(--input-foreground) ${
          fieldState.showError
            ? 'border-error focus:border-error focus:ring-error/20 glow-error'
            : fieldState.isValid
            ? 'border-success focus:border-success focus:ring-success/20 glow-success'
            : ''
        }`}
      />
    </BaseFormField>
  );
}

interface EmailFormFieldProps {
  id: string;
  label: string;
  required?: boolean;
  placeholder: string;
  fieldState: FormFieldState;
  error?: FieldError;
  register: any;
}

export function EmailFormField({ id, label, required, placeholder, fieldState, error, register }: EmailFormFieldProps) {
  return (
    <BaseFormField id={id} label={label} required={required} fieldState={fieldState} error={error}>
      <Input
        id={id}
        type="email"
        {...register}
        placeholder={placeholder}
        className={`backdrop-blur-sm border-2 transition-all duration-300 hover:border-brand-500 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 focus:glow-sm text-(--input-foreground) ${
          fieldState.showError
            ? 'border-error focus:border-error focus:ring-error/20 glow-error'
            : fieldState.isValid
            ? 'border-success focus:border-success focus:ring-success/20 glow-success'
            : ''
        }`}
      />
    </BaseFormField>
  );
}

interface TextareaFormFieldProps {
  id: string;
  label: string;
  required?: boolean;
  placeholder: string;
  fieldState: FormFieldState;
  error?: FieldError;
  register: any;
  characterCount?: { current: number; max: number };
}

export function TextareaFormField({
  id,
  label,
  required,
  placeholder,
  fieldState,
  error,
  register,
  characterCount
}: TextareaFormFieldProps) {
  return (
    <div className="space-y-3 animate-fade-in">
      <Label htmlFor={id} className="flex items-center justify-between text-text-primary font-medium">
        <span className="flex items-center gap-2">
          {label}
          {required && <span className="text-error">*</span>}
          {fieldState.isValid && (
            <CheckCircle className="h-4 w-4 text-success animate-pulse-glow" />
          )}
        </span>
        {characterCount && (
          <span className={`text-xs transition-colors duration-200 ${
            characterCount.current > characterCount.max * 0.9
              ? 'text-warning animate-pulse'
              : 'text-text-muted'
          }`}>
            {characterCount.current}/{characterCount.max}
          </span>
        )}
      </Label>
      <div className="relative group">
        <Textarea
          id={id}
          {...register}
          placeholder={placeholder}
          rows={6}
          className={`backdrop-blur-sm border-2 transition-all duration-300 hover:border-brand-500 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 focus:glow-sm resize-none text-(--input-foreground) ${
            fieldState.showError
              ? 'border-error focus:border-error focus:ring-error/20 glow-error'
              : fieldState.isValid
              ? 'border-success focus:border-success focus:ring-success/20 glow-success'
              : ''
          }`}
        />
        {fieldState.hasValue && (
          <div className="absolute right-3 bottom-3 transition-all duration-200">
            {fieldState.isValid ? (
              <CheckCircle className="h-4 w-4 text-success animate-bounce-subtle glow-success" />
            ) : fieldState.showError ? (
              <AlertCircle className="h-4 w-4 text-error animate-shake glow-error" />
            ) : null}
          </div>
        )}
      </div>
      {fieldState.showError && error?.message && (
        <p className="text-sm text-error flex items-center gap-2 animate-slide-up bg-error/10 px-3 py-2 rounded-md border border-error/20">
          <AlertCircle className="h-3 w-3 animate-pulse" />
          {error.message}
        </p>
      )}
    </div>
  );
}

interface SelectFormFieldProps {
  id: string;
  label: string;
  fieldState: FormFieldState;
  error?: FieldError;
  register: any;
  options: Array<{ value: string; label: string }>;
}

export function SelectFormField({ id, label, fieldState, error, register, options }: SelectFormFieldProps) {
  return (
    <div className="space-y-3 animate-fade-in">
      <Label htmlFor={id} className="text-text-primary font-medium">{label}</Label>
      <div className="relative">
        <select
          id={id}
          {...register}
          className={`w-full px-4 py-3 backdrop-blur-sm border-2 rounded-lg transition-all duration-300 hover:border-brand-500 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 focus:glow-sm appearance-none cursor-pointer text-(--input-foreground) ${
            fieldState.showError
              ? 'border-error focus:border-error focus:ring-error/20 glow-error'
              : ''
          }`}
        >
          {options.map((option) => (
            <option key={option.value} value={option.value} className="bg-bg-primary">
              {option.label}
            </option>
          ))}
        </select>
        <div className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none">
          <svg className="h-4 w-4 text-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      {fieldState.showError && error?.message && (
        <p className="text-sm text-error flex items-center gap-2 animate-slide-up bg-error/10 px-3 py-2 rounded-md border border-error/20">
          <AlertCircle className="h-3 w-3 animate-pulse" />
          {error.message}
        </p>
      )}
    </div>
  );
}