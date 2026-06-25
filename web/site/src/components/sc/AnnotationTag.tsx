import React from 'react';

interface AnnotationTagProps {
  label: string;
  detail?: string;
}

export const AnnotationTag: React.FC<AnnotationTagProps> = ({
  label,
  detail,
}) => {
  return (
    <div
      className="absolute hidden lg:block font-[var(--font-mono)] text-[11px] uppercase tracking-widest text-[var(--text-faint)] opacity-60"
      style={{
        top: 'var(--space-7)',
        right: 'var(--space-7)',
        lineHeight: 1.7,
        textAlign: 'right',
      }}
    >
      <span>{label}</span>
      {detail && (
        <>
          <span className="mx-1 opacity-40">/</span>
          <span className="text-[var(--text-dim)]">{detail}</span>
        </>
      )}
    </div>
  );
};
