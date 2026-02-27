import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { AlertCircle, CheckCircle2, Loader2, RefreshCw } from "lucide-react";

interface IOType {
  type: string;
  example?: any;
  schema?: any;
  required?: boolean;
  description?: string;
  title?: string;
}

interface ValidationError {
  field: string;
  message: string;
}

interface ManifestInputFormProps {
  inputSpec: IOType;
  value: any;
  onChange: (value: any) => void;
  onSubmit?: (value: any) => Promise<void> | void;
  onValidate?: (value: any) => ValidationError[] | Promise<ValidationError[]>;
  className?: string;
  title?: string;
  description?: string;
  submitLabel?: string;
  showSubmitButton?: boolean;
  disabled?: boolean;
  loading?: boolean;
  errors?: ValidationError[];
}

export function ManifestInputForm({
  inputSpec,
  value,
  onChange,
  onSubmit,
  onValidate,
  className = "",
  title,
  description,
  submitLabel = "Submit",
  showSubmitButton = false,
  disabled = false,
  loading = false,
  errors = []
}: ManifestInputFormProps) {
  const [internalValue, setInternalValue] = useState(value || getDefaultValue(inputSpec));
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    setInternalValue(value || getDefaultValue(inputSpec));
  }, [value, inputSpec]);

  useEffect(() => {
    setValidationErrors(errors);
  }, [errors]);

  const handleChange = async (newValue: any) => {
    setInternalValue(newValue);
    onChange(newValue);

    // Real-time validation if validator is provided
    if (onValidate) {
      try {
        const validationResult = await onValidate(newValue);
        setValidationErrors(validationResult);
      } catch (error) {
        console.error("Validation error:", error);
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (disabled || loading || isSubmitting) return;

    setIsSubmitting(true);
    try {
      if (onSubmit) {
        await onSubmit(internalValue);
      }
    } catch (error) {
      console.error("Submit error:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const getFieldError = (fieldName: string) => {
    return validationErrors.find(error => error.field === fieldName);
  };

  const formContent = (
    <>
      {inputSpec.type === 'string' && (
        <div className="space-y-2">
          <Label htmlFor="input-string" className="text-sm font-medium">
            {inputSpec.title || 'Input'}
            {inputSpec.required && <span className="text-red-500 ml-1">*</span>}
          </Label>
          {inputSpec.description && (
            <p className="text-xs text-text-secondary">{inputSpec.description}</p>
          )}
          {inputSpec.example && typeof inputSpec.example === 'string' && inputSpec.example.length > 100 ? (
            <Textarea
              id="input-string"
              value={internalValue || ''}
              onChange={(e) => handleChange(e.target.value)}
              placeholder={inputSpec.example}
              className={cn(
                "font-mono text-sm",
                disabled && "opacity-50 cursor-not-allowed"
              )}
              rows={6}
              disabled={disabled}
            />
          ) : (
            <Input
              id="input-string"
              type="text"
              value={internalValue || ''}
              onChange={(e) => handleChange(e.target.value)}
              placeholder={inputSpec.example || 'Enter input...'}
              className={cn(disabled && "opacity-50 cursor-not-allowed")}
              disabled={disabled}
            />
          )}
        </div>
      )}

      {inputSpec.type === 'object' && inputSpec.schema?.properties && (
        <div className="space-y-6">
          {Object.entries(inputSpec.schema.properties).map(([key, propSpec]: [string, any]) => {
            const fieldError = getFieldError(key);
            const isRequired = inputSpec.schema.required?.includes(key);

            return (
              <div key={key} className="space-y-2">
                <Label htmlFor={`input-${key}`} className="text-sm font-medium">
                  {propSpec.title || key}
                  {isRequired && <span className="text-red-500 ml-1">*</span>}
                </Label>
                {propSpec.description && (
                  <p className="text-xs text-text-secondary">{propSpec.description}</p>
                )}
                <FieldRenderer
                  fieldKey={key}
                  fieldSpec={propSpec}
                  value={internalValue?.[key]}
                  onChange={(fieldValue) => {
                    handleChange({
                      ...internalValue,
                      [key]: fieldValue
                    });
                  }}
                  disabled={disabled}
                  error={fieldError?.message}
                />
                {fieldError && (
                  <div className="flex items-center gap-2 text-xs text-red-500">
                    <AlertCircle className="w-3 h-3" />
                    {fieldError.message}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Fallback for unsupported types */}
      <div className="space-y-2">
        <Label htmlFor="input-raw" className="text-sm font-medium">
          {inputSpec.title || 'Input (JSON)'}
          {inputSpec.required && <span className="text-red-500 ml-1">*</span>}
        </Label>
        {inputSpec.description && (
          <p className="text-xs text-text-secondary">{inputSpec.description}</p>
        )}
        <Textarea
          id="input-raw"
          value={typeof internalValue === 'string' ? internalValue : JSON.stringify(internalValue, null, 2)}
          onChange={(e) => {
            try {
              handleChange(JSON.parse(e.target.value));
            } catch {
              handleChange(e.target.value);
            }
          }}
          placeholder={inputSpec.example ? JSON.stringify(inputSpec.example, null, 2) : 'Enter JSON input...'}
          className={cn(
            "font-mono text-sm",
            disabled && "opacity-50 cursor-not-allowed"
          )}
          rows={8}
          disabled={disabled}
        />
        <div className="flex items-center gap-2 text-xs text-text-secondary">
          <AlertCircle className="w-3 h-3" />
          Enter valid JSON format
        </div>
      </div>
    </>
  );

  return (
    <Card className={cn("", className)}>
      {(title || description) && (
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              {title && <CardTitle className="text-lg">{title}</CardTitle>}
              {description && (
                <p className="text-sm text-text-secondary mt-1">{description}</p>
              )}
            </div>
            {validationErrors.length > 0 && (
              <Badge variant="destructive" className="text-xs">
                {validationErrors.length} error{validationErrors.length !== 1 ? 's' : ''}
              </Badge>
            )}
          </div>
        </CardHeader>
      )}

      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {formContent}

          {/* Validation Errors Summary */}
          {validationErrors.length > 0 && (
            <Alert className="border-red-500/20 bg-red-500/10">
              <AlertCircle className="h-4 w-4 text-red-500" />
              <AlertDescription className="text-red-600 dark:text-red-400">
                <div className="space-y-1">
                  {validationErrors.map((error, index) => (
                    <div key={index} className="text-sm">
                      <strong>{error.field}:</strong> {error.message}
                    </div>
                  ))}
                </div>
              </AlertDescription>
            </Alert>
          )}

          {/* Submit Button */}
          {showSubmitButton && (
            <div className="flex justify-end pt-4">
              <Button
                type="submit"
                disabled={disabled || loading || isSubmitting || validationErrors.length > 0}
                className="min-w-[120px]"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Submitting...
                  </>
                ) : loading ? (
                  <>
                    <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                    Loading...
                  </>
                ) : (
                  submitLabel
                )}
              </Button>
            </div>
          )}
        </form>
      </CardContent>
    </Card>
  );
}

interface FieldRendererProps {
  fieldKey: string;
  fieldSpec: any;
  value: any;
  onChange: (value: any) => void;
  disabled?: boolean;
  error?: string;
}

function FieldRenderer({ fieldKey, fieldSpec, value, onChange, disabled = false, error }: FieldRendererProps) {
  const fieldType = fieldSpec.type || 'string';
  const defaultValue = getDefaultValue(fieldSpec);

  useEffect(() => {
    if (value === undefined && defaultValue !== undefined) {
      onChange(defaultValue);
    }
  }, [value, defaultValue, onChange]);

  const currentValue = value !== undefined ? value : defaultValue;

  switch (fieldType) {
    case 'string':
      if (fieldSpec.enum) {
        return (
          <Select value={currentValue || ''} onValueChange={onChange} disabled={disabled}>
            <SelectTrigger className={cn("mt-1", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}>
              <SelectValue placeholder={`Select ${fieldKey}...`} />
            </SelectTrigger>
            <SelectContent>
              {fieldSpec.enum.map((option: string) => (
                <SelectItem key={option} value={option}>
                  {option}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        );
      }

      return fieldSpec.example && typeof fieldSpec.example === 'string' && fieldSpec.example.length > 50 ? (
        <Textarea
          id={`input-${fieldKey}`}
          value={currentValue || ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder={fieldSpec.example}
          className={cn("mt-1 font-mono text-sm", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}
          rows={3}
          disabled={disabled}
        />
      ) : (
        <Input
          id={`input-${fieldKey}`}
          type="text"
          value={currentValue || ''}
          onChange={(e) => onChange(e.target.value)}
          placeholder={fieldSpec.example || `Enter ${fieldKey}...`}
          className={cn("mt-1", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}
          disabled={disabled}
        />
      );

    case 'number':
    case 'integer':
      return (
        <Input
          id={`input-${fieldKey}`}
          type="number"
          value={currentValue || ''}
          onChange={(e) => onChange(e.target.value ? Number(e.target.value) : undefined)}
          placeholder={fieldSpec.example?.toString() || `Enter ${fieldKey}...`}
          className={cn("mt-1", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}
          disabled={disabled}
        />
      );

    case 'boolean':
      return (
        <div className="flex items-center space-x-2 mt-1">
          <Checkbox
            id={`input-${fieldKey}`}
            checked={!!currentValue}
            onCheckedChange={onChange}
            disabled={disabled}
            className={cn(error && "border-red-500")}
          />
          <Label
            htmlFor={`input-${fieldKey}`}
            className={cn("text-sm", disabled && "opacity-50 cursor-not-allowed")}
          >
            {fieldSpec.title || fieldKey}
          </Label>
        </div>
      );

    case 'array':
      // Simple array input as JSON string for now
      return (
        <Textarea
          id={`input-${fieldKey}`}
          value={Array.isArray(currentValue) ? JSON.stringify(currentValue, null, 2) : currentValue || ''}
          onChange={(e) => {
            try {
              onChange(JSON.parse(e.target.value));
            } catch {
              onChange(e.target.value);
            }
          }}
          placeholder={fieldSpec.example ? JSON.stringify(fieldSpec.example, null, 2) : `Enter ${fieldKey} array...`}
          className={cn("mt-1 font-mono text-sm", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}
          rows={3}
          disabled={disabled}
        />
      );

    case 'object':
      // Nested object as JSON string for now
      return (
        <Textarea
          id={`input-${fieldKey}`}
          value={typeof currentValue === 'object' ? JSON.stringify(currentValue, null, 2) : currentValue || ''}
          onChange={(e) => {
            try {
              onChange(JSON.parse(e.target.value));
            } catch {
              onChange(e.target.value);
            }
          }}
          placeholder={fieldSpec.example ? JSON.stringify(fieldSpec.example, null, 2) : `Enter ${fieldKey} object...`}
          className={cn("mt-1 font-mono text-sm", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}
          rows={4}
          disabled={disabled}
        />
      );

    default:
      return (
        <Input
          id={`input-${fieldKey}`}
          type="text"
          value={typeof currentValue === 'string' ? currentValue : JSON.stringify(currentValue || '')}
          onChange={(e) => onChange(e.target.value)}
          placeholder={`Enter ${fieldKey}...`}
          className={cn("mt-1", error && "border-red-500", disabled && "opacity-50 cursor-not-allowed")}
          disabled={disabled}
        />
      );
  }
}

function getDefaultValue(spec: IOType): any {
  if (spec.example !== undefined) {
    return spec.example;
  }

  switch (spec.type) {
    case 'string':
      return '';
    case 'number':
    case 'integer':
      return 0;
    case 'boolean':
      return false;
    case 'object':
      return {};
    case 'array':
      return [];
    default:
      return null;
  }
}