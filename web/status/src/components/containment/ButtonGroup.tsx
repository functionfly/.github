import React from 'react';

interface ButtonGroupProps {
  children: React.ReactNode;
  className?: string;
  responsive?: boolean;
  centered?: boolean;
}

export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  children,
  className = '',
  responsive = true,
  centered = false,
}) => {
  return (
    <div
      className={`flex ${className}`}
      style={{
        flexDirection: responsive ? undefined : 'row',
        alignItems: 'center',
        gap: 'var(--space-3)',
        justifyContent: centered ? 'center' : undefined,
      }}
    >
      {children}
      {responsive && (
        <style>{`
          @media (max-width: 479px) {
            .${className || 'button-group'} {
              flex-direction: column;
              align-items: stretch;
            }
          }
        `}</style>
      )}
    </div>
  );
};
