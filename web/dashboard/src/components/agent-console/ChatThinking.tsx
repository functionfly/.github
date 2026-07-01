import { Brain, ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from 'react';

interface ChatThinkingProps {
  content: string;
  tokens: number;
}

export function ChatThinking({ content, tokens }: ChatThinkingProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="ach-thinking">
      <button className="ach-thinking__header" onClick={() => setExpanded(!expanded)}>
        <Brain style={{ width: 14, height: 14 }} />
        <span className="ach-thinking__label">Thinking</span>
        <span className="ach-thinking__tokens">{tokens.toLocaleString()} tokens</span>
        {expanded ? <ChevronUp style={{ width: 14, height: 14 }} /> : <ChevronDown style={{ width: 14, height: 14 }} />}
      </button>
      {expanded && (
        <div className="ach-thinking__content">
          <pre className="ach-thinking__text">{content}</pre>
        </div>
      )}
    </div>
  );
}
