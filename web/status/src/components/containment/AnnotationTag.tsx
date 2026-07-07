import React from 'react';

interface AnnotationTagProps {
  label: string;
  detail?: string;
  className?: string;
}

export const AnnotationTag: React.FC<AnnotationTagProps> = ({
  label,
  detail,
  className = '',
}) => {
  const classes = ['annotation-tag', className].filter(Boolean).join(' ');

  return (
    <div className={classes}>
      <span>{label}</span>
      {detail && (
        <>
          <span className="mx-1 opacity-40">/</span>
          <span className="annotation-tag__subtitle">{detail}</span>
        </>
      )}
    </div>
  );
};
