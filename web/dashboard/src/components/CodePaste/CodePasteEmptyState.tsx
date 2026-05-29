import { FileCode, ListChecks, Search, Sparkles } from 'lucide-react';
import './CodePasteEmptyState.css';

export function CodePasteEmptyState() {
  return (
    <div className="code-paste-empty-state">
      <Search className="code-paste-empty-state__icon" aria-hidden="true" />
      <h3>Paste code and click Parse</h3>
      <p>We&apos;ll detect functions and help you import them as deployable functions.</p>
      <div className="code-paste-empty-state__features">
        <div className="code-paste-empty-state__feature">
          <Sparkles className="code-paste-empty-state__feature-icon" aria-hidden="true" />
          <span>Auto-detect language</span>
        </div>
        <div className="code-paste-empty-state__feature">
          <FileCode className="code-paste-empty-state__feature-icon" aria-hidden="true" />
          <span>Extract function signatures</span>
        </div>
        <div className="code-paste-empty-state__feature">
          <ListChecks className="code-paste-empty-state__feature-icon" aria-hidden="true" />
          <span>Choose what to import</span>
        </div>
      </div>
    </div>
  );
}
