import React from "react";

export interface chamberProps {
  children: React.ReactNode;
  ribs?: boolean;
  nested?: boolean;
  className?: string;
  style?: React.CSSProperties;
  id?: string;
}

export const Chamber: React.FC<chamberProps> = ({
  children,
  ribs = false,
  nested = false,
  className = "",
  style,
  id,
}) => {
  const classes = [
    'chamber',
    ribs && !nested ? 'chamber--ribs' : '',
    nested ? 'chamber--nested' : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <div
      id={id}
      className={classes}
      style={style}
    >
      {ribs && !nested && (
        <div
          className="absolute inset-0 pointer-events-none select-none"
          style={{
            opacity: 0.025,
            backgroundImage:
              "repeating-linear-gradient(90deg, transparent 0px, transparent 119px, rgba(255,255,255,0.025) 120px)",
            borderRadius: "inherit",
          }}
          aria-hidden="true"
        />
      )}
      {children}
    </div>
  );
};
