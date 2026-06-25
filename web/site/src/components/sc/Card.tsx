import React from 'react';

interface CardProps {
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

export const Card: React.FC<CardProps> = ({
  children,
  className = '',
  style,
}) => {
  return (
    <div
      className={`relative ${className}`}
      style={{
        background: 'var(--panel-raised)',
        borderRadius: 'var(--radius)',
        border: '1px solid var(--panel-edge)',
        padding: 'var(--space-5)',
        boxShadow: 'none',
        ...style,
      }}
    >
      {children}
    </div>
  );
};

interface CardTitleProps {
  children: React.ReactNode;
  className?: string;
}

export const CardTitle: React.FC<CardTitleProps> = ({
  children,
  className = '',
}) => {
  return (
    <h3
      className={`ff-font-display text-[22px] font-medium leading-[1.25] text-[var(--text)] ${className}`}
    >
      {children}
    </h3>
  );
};

interface CardDescriptionProps {
  children: React.ReactNode;
  className?: string;
}

export const CardDescription: React.FC<CardDescriptionProps> = ({
  children,
  className = '',
}) => {
  return (
    <p
      className={`mt-[var(--space-2)] text-[15px] leading-[1.6] text-[var(--text-dim)] ${className}`}
    >
      {children}
    </p>
  );
};
