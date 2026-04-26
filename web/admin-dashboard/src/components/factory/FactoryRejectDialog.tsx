/**
 * Factory Reject Dialog Component
 * Confirmation dialog for rejecting opportunities with undo support
 */

import { useState, useEffect } from 'react';
import { XCircle, Undo2 } from 'lucide-react';

interface RejectionRecord {
  id: string;
  reason: string;
  timestamp: number;
}

interface FactoryRejectDialogProps {
  open: boolean;
  reason: string;
  isPending: boolean;
  onOpenChange: (open: boolean) => void;
  onReasonChange: (reason: string) => void;
  onConfirm: () => void;
  onUndo?: (id: string) => void;
}

export function FactoryRejectDialog({
  open,
  reason,
  isPending,
  onOpenChange,
  onReasonChange,
  onConfirm,
  onUndo,
}: FactoryRejectDialogProps) {
  const [rejectionRecords, setRejectionRecords] = useState<RejectionRecord[]>([]);

  useEffect(() => {
    if (!open && rejectionRecords.length > 0) {
      const lastRecord = rejectionRecords[rejectionRecords.length - 1];
      const timeLeft = 5000 - (Date.now() - lastRecord.timestamp);
      if (timeLeft > 0 && onUndo) {
        const timeout = setTimeout(() => {
          setRejectionRecords((prev) => prev.slice(0, -1));
        }, timeLeft);
        return () => clearTimeout(timeout);
      }
    }
  }, [open, rejectionRecords, onUndo]);

  if (!open) return null;

  const handleConfirm = () => {
    if (!reason.trim() || isPending) return;
    const record: RejectionRecord = {
      id: `rejection-${Date.now()}`,
      reason,
      timestamp: Date.now(),
    };
    setRejectionRecords((prev) => [...prev, record]);
    onConfirm();
  };

  return (
    <div className="fixed inset-0 bg-black/50 dark:bg-black/70 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full mx-4 border border-gray-200 dark:border-gray-700">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Reject Opportunity</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Please provide a reason for rejecting this opportunity.
          </p>
        </div>
        <div className="px-6 py-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Rejection Reason
          </label>
          <textarea
            value={reason}
            onChange={(e) => onReasonChange(e.target.value)}
            placeholder="Enter the reason for rejection..."
            rows={4}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          {rejectionRecords.length > 0 && (
            <div className="mt-3 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
              <div className="flex items-center gap-2 text-yellow-800 dark:text-yellow-300 mb-2">
                <Undo2 className="h-4 w-4" />
                <span className="text-sm font-medium">Undo Available</span>
              </div>
              <p className="text-xs text-yellow-700 dark:text-yellow-400 mb-2">
                You can undo your last rejection within 5 seconds.
              </p>
              {rejectionRecords.map((record) => {
                const timeLeft = Math.max(0, 5000 - (Date.now() - record.timestamp));
                return (
                  <div key={record.id} className="flex items-center justify-between">
                    <span className="text-sm text-gray-700 dark:text-gray-300 truncate flex-1">
                      {record.reason.substring(0, 30)}...
                    </span>
                    <button
                      onClick={() => {
                        onUndo?.(record.id);
                        setRejectionRecords((prev) => prev.filter((r) => r.id !== record.id));
                      }}
                      className="ml-2 px-2 py-1 text-xs bg-yellow-200 dark:bg-yellow-800 text-yellow-800 dark:text-yellow-200 rounded hover:bg-yellow-300 dark:hover:bg-yellow-700 transition-colors"
                    >
                      Undo ({Math.ceil(timeLeft / 1000)}s)
                    </button>
                  </div>
                );
              })}
            </div>
          )}
        </div>
        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
          <button
            onClick={() => onOpenChange(false)}
            className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-gray-700 dark:text-gray-300"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={!reason.trim() || isPending}
            className="px-4 py-2 bg-red-600 dark:bg-red-600 text-white rounded-lg hover:bg-red-700 dark:hover:bg-red-700 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            <XCircle className="h-4 w-4" />
            {isPending ? 'Rejecting...' : 'Reject Opportunity'}
          </button>
        </div>
      </div>
    </div>
  );
}
