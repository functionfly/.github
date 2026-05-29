import { AlertCircle } from 'lucide-react';
import './ImportErrorsList.css';

interface ImportErrorsListProps {
  errors: Array<{ name: string; error: string }>;
}

export function ImportErrorsList({ errors }: ImportErrorsListProps) {
  if (errors.length === 0) {
    return null;
  }

  return (
    <div className="import-errors-list" role="alert" aria-live="polite">
      <div className="import-errors-list__header">
        <AlertCircle className="import-errors-list__icon" aria-hidden="true" />
        <span>
          {errors.length} import issue{errors.length !== 1 ? 's' : ''}
        </span>
      </div>
      <ul className="import-errors-list__items">
        {errors.map((item, index) => (
          <li key={`${item.name}-${index}`}>
            {item.name ? (
              <>
                <strong>{item.name}:</strong> {item.error}
              </>
            ) : (
              item.error
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
