import React from 'react';

interface ButtonGroupProps {
  children: React.ReactNode;
  className?: string;
  /** Stack buttons vertically on mobile (<480px) */
  responsive?: boolean;
  /** Center the button group within its container */
  centered?: boolean;
}

/**
 * ButtonGroup - Layout container for button pairs.
 *
 * Layout rules:
 * - Primary action always comes first (left in LTR, top in stacked)
 * - Gap between paired buttons: var(--space-3) (12px)
 * - On screens under 480px width, stacks buttons full-width vertically
 * - Never has more than one SealedButton (primary) - additional actions use FrameButton
 */
export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  children,
  className = '',
  responsive = true,
  centered = false,
}) => {
  return (
    <div
      className={`
        sealed-button-group
        ${responsive ? 'sealed-button-group-responsive' : ''}
        ${centered ? 'sealed-button-group-centered' : ''}
        ${className}
      `}
    >
      {children}
    </div>
  );
};
