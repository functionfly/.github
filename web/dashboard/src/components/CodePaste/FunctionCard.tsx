import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { clsx } from 'clsx';
import type { ParsedFunction } from '../../types/codePaste';
import './FunctionCard.css';

interface FunctionCardProps {
  function: ParsedFunction;
  isSelected: boolean;
  onSelect: () => void;
  onPreview: () => void;
  onNameChange: (name: string) => void;
}

const languageIcons: Record<string, string> = {
  python: '🐍',
  javascript: '📜',
  typescript: '📘',
  go: '🐹',
  rust: '🦀',
  ruby: '💎',
  java: '☕',
  kotlin: '🟣',
  swift: '🍎',
  cpp: '⚙️',
  c: '🔧',
};

export function FunctionCard({
  function: func,
  isSelected,
  onSelect,
  onPreview,
  onNameChange,
}: FunctionCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editedName, setEditedName] = useState(func.name);

  const handleNameSubmit = () => {
    if (editedName.trim() && editedName !== func.name) {
      onNameChange(editedName.trim());
    }
    setIsEditing(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleNameSubmit();
    } else if (e.key === 'Escape') {
      setEditedName(func.name);
      setIsEditing(false);
    }
  };

  const icon = languageIcons[func.language] || '📄';

  return (
    <motion.div
      className={clsx('function-card', {
        'function-card--selected': isSelected,
        'function-card--expanded': isExpanded,
      })}
      layout
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      transition={{ duration: 0.2 }}
    >
      <div className="function-card__header">
        <label className="function-card__checkbox-wrapper">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={onSelect}
            className="function-card__checkbox"
          />
          <span className="function-card__checkbox-custom">
            {isSelected && (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            )}
          </span>
        </label>

        <div className="function-card__icon">{icon}</div>

        <div className="function-card__info">
          {isEditing ? (
            <input
              type="text"
              value={editedName}
              onChange={(e) => setEditedName(e.target.value)}
              onBlur={handleNameSubmit}
              onKeyDown={handleKeyDown}
              className="function-card__name-input"
              autoFocus
            />
          ) : (
            <button
              className="function-card__name"
              onClick={() => setIsEditing(true)}
              onDoubleClick={() => setIsEditing(true)}
            >
              {func.name}
            </button>
          )}
          <div className="function-card__signature">{func.signature}</div>
        </div>

        <button
          className="function-card__expand-btn"
          onClick={() => setIsExpanded(!isExpanded)}
          aria-label={isExpanded ? 'Collapse' : 'Expand'}
        >
          <svg
            className={clsx('function-card__expand-icon', { 'function-card__expand-icon--open': isExpanded })}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>
      </div>

      {func.docstring && (
        <div className="function-card__docstring">{func.docstring}</div>
      )}

      <div className="function-card__meta">
        <span className="function-card__lines">
          Lines {func.start_line}-{func.end_line}
        </span>
        {func.parameters.length > 0 && (
          <span className="function-card__params">
            {func.parameters.length} parameter{func.parameters.length !== 1 ? 's' : ''}
          </span>
        )}
        {func.return_type && (
          <span className="function-card__return">
            Returns: {func.return_type}
          </span>
        )}
      </div>

      <AnimatePresence>
        {isExpanded && (
          <motion.div
            className="function-card__expanded-content"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
          >
            <div className="function-card__code-header">
              <span>Code Preview</span>
              <button className="function-card__copy-btn" onClick={onPreview}>
                Preview
              </button>
            </div>
            <pre className="function-card__code">
              <code>{func.code}</code>
            </pre>
          </motion.div>
        )}
      </AnimatePresence>

      <div className="function-card__actions">
        <button className="function-card__action-btn" onClick={() => setIsEditing(true)}>
          Edit Name
        </button>
        <button className="function-card__action-btn" onClick={onPreview}>
          Preview
        </button>
      </div>
    </motion.div>
  );
}