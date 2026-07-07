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
      className={`${className}`}
      style={{
        fontFamily: 'var(--font-display)',
        fontSize: '22px',
        fontWeight: 500,
        lineHeight: 1.25,
        color: 'var(--text)',
      }}
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
      className={`${className}`}
      style={{
        marginTop: 'var(--space-2)',
        fontSize: '15px',
        lineHeight: 1.6,
        color: 'var(--text-dim)',
      }}
    >
      {children}
    </p>
  );
};
