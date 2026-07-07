import { useState, useRef, useEffect } from 'react';
import { languages, type Language } from '../lib/i18n/languages';
import { getLocaleFromUrl } from '../i18n/index';

interface LanguagePickerProps {
  currentLocale?: string;
}

export default function LanguagePicker({ currentLocale }: LanguagePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const locale = currentLocale || 'en';
  const currentLanguage = languages.find(l => l.code === locale) || languages[0];

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  function handleLocaleChange(newLocale: string) {
    const currentPath = window.location.pathname;
    const pathParts = currentPath.split('/').filter(Boolean);

    if (languages.some(l => l.code === pathParts[0])) {
      pathParts[0] = newLocale;
    } else {
      pathParts.unshift(newLocale);
    }

    window.location.pathname = '/' + pathParts.join('/');
  }

  return (
    <div className="language-picker" ref={dropdownRef}>
      <button
        className="language-picker__trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-label="Select language"
      >
        <span className="language-picker__current">
          {currentLanguage.nativeName}
        </span>
        <svg
          className={`language-picker__chevron ${isOpen ? 'is-open' : ''}`}
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>

      {isOpen && (
        <ul className="language-picker__dropdown" role="listbox">
          {languages.map((lang: Language) => (
            <li key={lang.code}>
              <button
                className={`language-picker__option ${lang.code === locale ? 'is-active' : ''}`}
                onClick={() => {
                  handleLocaleChange(lang.code);
                  setIsOpen(false);
                }}
                role="option"
                aria-selected={lang.code === locale}
              >
                <span className="language-picker__option-native">{lang.nativeName}</span>
                <span className="language-picker__option-name">{lang.name}</span>
              </button>
            </li>
          ))}
        </ul>
      )}

      <style>{`
        .language-picker {
          position: relative;
        }

        .language-picker__trigger {
          display: flex;
          align-items: center;
          gap: 0.35rem;
          padding: 0.4rem 0.6rem;
          background: rgba(var(--ff-tarmac-rgb, 13, 17, 23), 0.6);
          backdrop-filter: blur(8px);
          border: 1px solid var(--color-border);
          border-radius: 6px;
          color: var(--color-text-muted);
          font-size: 0.8rem;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s;
        }

        .language-picker__trigger:hover {
          color: var(--color-text);
          border-color: var(--ff-flame);
        }

        .language-picker__chevron {
          transition: transform 0.2s;
        }

        .language-picker__chevron.is-open {
          transform: rotate(180deg);
        }

        .language-picker__dropdown {
          position: absolute;
          top: 100%;
          right: 0;
          margin-top: 0.5rem;
          min-width: 160px;
          max-height: 280px;
          overflow-y: auto;
          background: var(--color-bg);
          border: 1px solid var(--color-border);
          border-radius: 10px;
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
          z-index: 1000;
          list-style: none;
          padding: 0.5rem;
        }

        .language-picker__option {
          display: flex;
          flex-direction: column;
          align-items: flex-start;
          width: 100%;
          padding: 0.6rem 0.75rem;
          background: transparent;
          border: none;
          border-radius: 6px;
          color: var(--color-text-muted);
          cursor: pointer;
          transition: all 0.15s;
          text-align: left;
        }

        .language-picker__option:hover {
          background: rgba(var(--ff-flame-rgb), 0.1);
          color: var(--color-text);
        }

        .language-picker__option.is-active {
          background: rgba(var(--ff-flame-rgb), 0.15);
          color: var(--ff-flame);
        }

        .language-picker__option-native {
          font-size: 0.875rem;
          font-weight: 500;
        }

        .language-picker__option-name {
          font-size: 0.7rem;
          opacity: 0.7;
          margin-top: 0.15rem;
        }

        [data-theme='light'] .language-picker__trigger {
          background: rgba(255, 255, 255, 0.8);
          border-color: var(--color-border);
          color: var(--color-text-muted);
        }

        [data-theme='light'] .language-picker__trigger:hover {
          border-color: var(--ff-flame);
        }

        [data-theme='light'] .language-picker__dropdown {
          background: var(--color-bg);
          border-color: var(--color-border);
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
        }

        [data-theme='light'] .language-picker__option {
          color: var(--color-text-muted);
        }

        [data-theme='light'] .language-picker__option:hover {
          background: rgba(var(--ff-flame-rgb), 0.08);
          color: var(--color-text);
        }

        [data-theme='light'] .language-picker__option.is-active {
          background: rgba(var(--ff-flame-rgb), 0.12);
          color: var(--ff-flame);
        }
      `}</style>
    </div>
  );
}