import { motion } from 'framer-motion';
import { FunctionCard } from './FunctionCard';
import type { ParsedFunction } from '../../types/codePaste';
import './FunctionList.css';

interface FunctionListProps {
  functions: ParsedFunction[];
  selectedIds: Set<string>;
  onSelectionChange: (id: string) => void;
  onSelectAll: () => void;
  onClearSelection: () => void;
  onFunctionPreview: (func: ParsedFunction) => void;
  onFunctionNameChange: (id: string, name: string) => void;
}

export function FunctionList({
  functions,
  selectedIds,
  onSelectionChange,
  onSelectAll,
  onClearSelection,
  onFunctionPreview,
  onFunctionNameChange,
}: FunctionListProps) {
  const allSelected = functions.length > 0 && functions.every((f) => selectedIds.has(f.id));
  const someSelected = functions.some((f) => selectedIds.has(f.id));

  return (
    <div className="function-list">
      <div className="function-list__header">
        <div className="function-list__title">
          <h3>Detected Functions ({functions.length})</h3>
          <p className="function-list__subtitle">Select the functions you want to import</p>
        </div>
        <div className="function-list__actions">
          <button
            className={`function-list__select-btn ${allSelected ? 'active' : ''}`}
            onClick={allSelected ? onClearSelection : onSelectAll}
          >
            {allSelected ? 'Deselect All' : 'Select All'}
          </button>
        </div>
      </div>

      <div className="function-list__content">
        {functions.length === 0 ? (
          <div className="function-list__empty">
            <div className="function-list__empty-icon">📭</div>
            <h4>No functions detected</h4>
            <p>Make sure your code contains valid function definitions.</p>
          </div>
        ) : (
          <motion.div
            className="function-list__items"
            layout
          >
            {functions.map((func) => (
              <FunctionCard
                key={func.id}
                function={func}
                isSelected={selectedIds.has(func.id)}
                onSelect={() => onSelectionChange(func.id)}
                onPreview={() => onFunctionPreview(func)}
                onNameChange={(name) => onFunctionNameChange(func.id, name)}
              />
            ))}
          </motion.div>
        )}
      </div>

      {someSelected && (
        <div className="function-list__selection-info">
          {selectedIds.size} of {functions.length} functions selected
        </div>
      )}
    </div>
  );
}