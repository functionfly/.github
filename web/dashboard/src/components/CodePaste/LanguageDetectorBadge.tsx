import { SUPPORTED_LANGUAGES, SupportedLanguage } from '../../types/codePaste';
import './LanguageDetectorBadge.css';

interface LanguageDetectorBadgeProps {
  language: string | null;
  confidence: number;
  hasCode?: boolean;
  isParsing?: boolean;
  onLanguageChange?: (language: string) => void;
  showConfidence?: boolean;
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
  unknown: '❓',
};

export function LanguageDetectorBadge({
  language,
  confidence,
  hasCode = false,
  isParsing = false,
  onLanguageChange,
  showConfidence = true,
}: LanguageDetectorBadgeProps) {
	const displayLanguage = language || 'unknown';
	const icon = languageIcons[displayLanguage] || '📄';
	let languageName: string;
	if (!hasCode) {
		languageName = 'Paste code to detect';
	} else if (isParsing) {
		languageName = 'Detecting language...';
	} else if (language) {
		languageName = SUPPORTED_LANGUAGES[language as SupportedLanguage] || language;
	} else {
		languageName = 'Detecting...';
	}

	const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
		if (onLanguageChange) {
			onLanguageChange(e.target.value);
		}
	};

	return (
		<div className="language-detector-badge">
			<div className="language-info">
				<span className="language-icon">{icon}</span>
				<span className="language-name">{languageName}</span>
				{showConfidence && hasCode && !isParsing && confidence > 0 && (
					<span className="confidence-badge">
						{Math.round(confidence)}%
					</span>
				)}
			</div>
      {onLanguageChange && (
        <select
          className="language-select"
          value={displayLanguage === 'unknown' ? 'auto' : displayLanguage}
          onChange={handleChange}
        >
          <option value="auto">Auto-detect</option>
          <option disabled>---</option>
          {Object.entries(SUPPORTED_LANGUAGES).map(([key, label]) => (
            <option key={key} value={key}>
              {label}
            </option>
          ))}
        </select>
      )}
    </div>
  );
}