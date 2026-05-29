import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LanguageDetectorBadge } from './LanguageDetectorBadge';

describe('LanguageDetectorBadge', () => {
  it('shows idle copy before code is pasted', () => {
    render(<LanguageDetectorBadge language={null} confidence={0} hasCode={false} />);
    expect(screen.getByText('Paste code to detect')).toBeInTheDocument();
  });

  it('shows parsing copy while analyzing', () => {
    render(<LanguageDetectorBadge language={null} confidence={0} hasCode isParsing />);
    expect(screen.getByText('Detecting language...')).toBeInTheDocument();
  });

  it('shows detected language name', () => {
    render(<LanguageDetectorBadge language="python" confidence={92} hasCode />);
    expect(screen.getByText('Python')).toBeInTheDocument();
    expect(screen.getByText('92%')).toBeInTheDocument();
  });
});
