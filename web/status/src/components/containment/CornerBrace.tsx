import React from 'react';

type Position = 'tl' | 'tr' | 'bl' | 'br';

interface CornerBraceProps {
  position: Position;
  className?: string;
}

export const CornerBrace: React.FC<CornerBraceProps> = ({
  position,
  className = '',
}) => {
  const classes = ['corner-brace', `corner-brace--${position}`, className].filter(Boolean).join(' ');

  return (
    <div
      className={classes}
      aria-hidden="true"
    />
  );
};
