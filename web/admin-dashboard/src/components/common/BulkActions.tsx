import { useState } from 'react';
import { MoreHorizontal, Download, Trash2, Ban, CheckCircle, X } from 'lucide-react';

export interface BulkAction {
  id: string;
  label: string;
  icon: React.ReactNode;
  variant?: 'default' | 'danger' | 'warning' | 'success';
  requiresConfirmation?: boolean;
  confirmMessage?: string;
}

interface BulkActionsProps {
  selectedCount: number;
  actions: BulkAction[];
  onAction: (actionId: string, selectedIds: string[]) => void;
  disabled?: boolean;
}

export function BulkActions({
  selectedCount,
  actions,
  onAction,
  disabled = false,
}: BulkActionsProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [confirmingAction, setConfirmingAction] = useState<string | null>(null);

  if (selectedCount === 0) {
    return null;
  }

  const handleActionClick = (action: BulkAction) => {
    if (action.requiresConfirmation) {
      setConfirmingAction(action.id);
    } else {
      // This would be triggered by parent component
      onAction(action.id, []);
    }
  };

  const handleConfirm = (actionId: string) => {
    onAction(actionId, []);
    setConfirmingAction(null);
  };

  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-gray-600">
        {selectedCount} selected
      </span>

      <div className="relative">
        <button
          onClick={() => setIsOpen(!isOpen)}
          disabled={disabled}
          className="flex items-center gap-2 px-3 py-1.5 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50"
        >
          <MoreHorizontal className="w-4 h-4" />
          Bulk Actions
        </button>

        {isOpen && (
          <>
            <div
              className="fixed inset-0 z-10"
              onClick={() => setIsOpen(false)}
            />
            <div className="absolute right-0 mt-1 w-48 bg-white rounded-lg shadow-lg border z-20">
              {actions.map((action) => (
                <button
                  key={action.id}
                  onClick={() => handleActionClick(action)}
                  className={`w-full flex items-center gap-2 px-4 py-2 text-sm text-left hover:bg-gray-50 first:rounded-t-lg last:rounded-b-lg ${
                    action.variant === 'danger' ? 'text-red-600' : ''
                  }`}
                >
                  {action.icon}
                  {action.label}
                </button>
              ))}
            </div>
          </>
        )}
      </div>

      {/* Confirmation Dialog */}
      {confirmingAction && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-30">
          <div className="bg-white rounded-lg p-6 max-w-sm">
            <h3 className="text-lg font-semibold mb-2">Confirm Action</h3>
            <p className="text-gray-600 mb-4">
              {actions.find((a) => a.id === confirmingAction)?.confirmMessage ||
                `Are you sure you want to perform this action on ${selectedCount} items?`}
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setConfirmingAction(null)}
                className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
              >
                Cancel
              </button>
              <button
                onClick={() => handleConfirm(confirmingAction)}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default BulkActions;
