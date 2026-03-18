import { FieldErrors } from 'react-hook-form';
import { AlertCircle } from 'lucide-react';

interface FormValidationSummaryProps {
  errors: FieldErrors;
  touchedFields: Record<string, any>;
  isValid: boolean;
}

export function FormValidationSummary({ errors, touchedFields, isValid }: FormValidationSummaryProps) {
  if (isValid || Object.keys(touchedFields).length === 0) return null;

  return (
    <div className="mt-4 p-3 bg-warning/5 border border-warning/20 rounded-md">
      <p className="text-sm text-warning font-medium mb-2">
        Please review the following:
      </p>
      <ul className="text-sm text-warning space-y-1">
        {Object.entries(errors).map(([field, error]) => {
          const message =
            typeof error === "object" && error && "message" in error
              ? (error as { message?: unknown }).message
              : undefined;

          if (typeof message !== "string") return null;

          return (
            <li key={field} className="flex items-center gap-1">
              <AlertCircle className="h-3 w-3" />
              {message}
            </li>
          );
        })}
      </ul>
    </div>
  );
}