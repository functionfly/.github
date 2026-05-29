import { MAX_CODE_BYTES, formatCodeSize, getCodeByteLength } from '../../utils/codePasteValidation';
import './CodeSizeIndicator.css';

interface CodeSizeIndicatorProps {
  code: string;
}

export function CodeSizeIndicator({ code }: CodeSizeIndicatorProps) {
  const bytes = getCodeByteLength(code);
  const isOverLimit = bytes > MAX_CODE_BYTES;
  const ratio = Math.min(bytes / MAX_CODE_BYTES, 1);

  return (
    <div
      className={`code-size-indicator${isOverLimit ? ' code-size-indicator--over' : ''}`}
      aria-live="polite"
    >
      <div className="code-size-indicator__bar" aria-hidden="true">
        <div
          className="code-size-indicator__fill"
          style={{ width: `${ratio * 100}%` }}
        />
      </div>
      <span className="code-size-indicator__label">
        {formatCodeSize(bytes)} / {formatCodeSize(MAX_CODE_BYTES)}
      </span>
    </div>
  );
}
