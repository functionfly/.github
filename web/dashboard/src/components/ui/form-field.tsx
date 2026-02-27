import { ReactNode, InputHTMLAttributes } from 'react';
import { FieldError, UseFormRegisterReturn, FieldErrors } from 'react-hook-form';
import { Input } from './input';
import { Label } from './label';
import { cn } from '@/lib/utils';
import { AlertCircle, CheckCircle2 } from 'lucide-react';

interface FormFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'name'> {
  label: string;
  name: string;
  error?: FieldError | any; // Allow broader error types from react-hook-form
  success?: boolean;
  required?: boolean;
  className?: string;
  children?: ReactNode;
  icon?: ReactNode;
  helperText?: string;
  showValidationIcon?: boolean;
  registerProps?: UseFormRegisterReturn;
}

export function FormField({
  label,
  name,
  type = 'text',
  placeholder,
  error,
  success,
  required,
  disabled,
  className,
  children,
  icon,
  helperText,
  showValidationIcon = true,
  registerProps,
  ...inputProps
}: FormFieldProps) {
  const hasError = !!error;
  const hasSuccess = success && !hasError;
  const showIcon = showValidationIcon && (hasError || hasSuccess);

  // Extract name from registerProps if it exists, otherwise use the name prop
  const { name: registerName, ...registerRest } = registerProps || {};
  const fieldName = registerName || name;

  return (
    <div className={cn('space-y-2', className)}>
      <Label htmlFor={fieldName} className={cn(
        'flex items-center gap-2',
        hasError && 'text-red-600 dark:text-red-400',
        hasSuccess && 'text-green-600 dark:text-green-400'
      )}>
        {label}
        {required && <span className="text-red-500">*</span>}
      </Label>

      <div className="relative">
        {children || (
          <Input
            id={fieldName}
            name={fieldName}
            type={type}
            placeholder={placeholder}
            disabled={disabled}
            className={cn(
              'transition-colors',
              hasError && 'border-red-500 focus:border-red-500 focus:ring-red-500',
              hasSuccess && 'border-green-500 focus:border-green-500 focus:ring-green-500'
            )}
            {...(registerProps ? {
              onChange: registerProps.onChange,
              onBlur: registerProps.onBlur,
              ref: registerProps.ref,
            } : {})}
            {...inputProps}
          />
        )}

        {icon && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted">
            {icon}
          </div>
        )}

        {showIcon && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2">
            {hasError && <AlertCircle className="w-4 h-4 text-red-500" />}
            {hasSuccess && <CheckCircle2 className="w-4 h-4 text-green-500" />}
          </div>
        )}
      </div>

      {/* Helper text */}
      {(helperText || hasError) && (
        <div className={cn(
          'text-xs',
          hasError && 'text-red-600 dark:text-red-400',
          !hasError && 'text-text-muted'
        )}>
          {(error as any)?.message || helperText}
        </div>
      )}
    </div>
  );
}