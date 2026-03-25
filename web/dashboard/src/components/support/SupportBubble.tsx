import { useSupportChat } from './SupportChat';

// ============================================================================
// Support Bubble Trigger
// ============================================================================

interface SupportBubbleProps {
  className?: string;
  showLabel?: boolean;
}

export function SupportBubble({ className = '', showLabel = true }: SupportBubbleProps) {
  const { openChat, isOpen } = useSupportChat();

  if (isOpen) return null;

  return (
    <button
      onClick={() => openChat()}
      className={`support-bubble ${className}`}
      style={{
        position: 'fixed',
        bottom: '24px',
        right: '24px',
        width: '60px',
        height: '60px',
        borderRadius: '50%',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        border: 'none',
        boxShadow: '0 4px 20px rgba(102, 126, 234, 0.4)',
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 9998,
        transition: 'transform 0.2s ease, box-shadow 0.2s ease',
      }}
      aria-label="Open support chat"
    >
      <svg
        width="28"
        height="28"
        viewBox="0 0 24 24"
        fill="none"
        stroke="white"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
        <circle cx="12" cy="10" r="1" fill="white" />
        <circle cx="8" cy="10" r="1" fill="white" />
        <circle cx="16" cy="10" r="1" fill="white" />
      </svg>
      {showLabel && (
        <span
          style={{
            position: 'absolute',
            top: '-8px',
            right: '-8px',
            background: '#10b981',
            color: 'white',
            fontSize: '10px',
            padding: '2px 6px',
            borderRadius: '10px',
            fontWeight: 600,
          }}
        >
          AI
        </span>
      )}
      <style>{`
        .support-bubble:hover {
          transform: scale(1.05);
          box-shadow: 0 6px 25px rgba(102, 126, 234, 0.5);
        }
      `}</style>
    </button>
  );
}

export default SupportBubble;
