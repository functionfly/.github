interface AnnotationTagProps {
  title: string;
  subtitle?: string;
}

export function AnnotationTag({ title, subtitle }: AnnotationTagProps) {
  return (
    <div className="annotation-tag">
      {subtitle && <div className="annotation-tag__subtitle">{subtitle}</div>}
      <div>{title}</div>
    </div>
  );
}
