export interface AnnotationTagProps {
  primary: string;
  secondary?: string;
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left';
  className?: string;
}

export function AnnotationTag({
  primary,
  secondary,
  position = 'top-right',
  className = '',
}: AnnotationTagProps) {
  return (
    <div className={`annotation-tag annotation-tag--${position} ${className}`}>
      <span className="annotation-tag__primary">{primary}</span>
      {secondary && <span className="annotation-tag__secondary">{secondary}</span>}
    </div>
  );
}