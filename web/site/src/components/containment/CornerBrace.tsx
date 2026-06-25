type CornerPosition = 'tl' | 'tr' | 'bl' | 'br';

interface CornerBraceProps {
  position: CornerPosition;
}

export function CornerBrace({ position }: CornerBraceProps) {
  return (
    <div
      className={`corner-brace corner-brace--${position}`}
      aria-hidden="true"
    />
  );
}
