import React from "react";

interface AnnotationTagProps {
  label: string;
  detail?: string;
  className?: string;
}

export const AnnotationTag: React.FC<AnnotationTagProps> = ({
  label,
  detail,
  className = "",
}) => {
  return (
    <div
      className={className}
      style={{
        fontFamily: "var(--font-mono)",
        fontSize: "10px",
        textTransform: "uppercase",
        letterSpacing: "0.18em",
        color: "var(--text-faint)",
        opacity: 0.6,
      }}
    >
      <span>{label}</span>
      {detail && (
        <>
          <span style={{ margin: "0 0.25rem", opacity: 0.4 }}>/</span>
          <span style={{ color: "var(--text-dim)" }}>{detail}</span>
        </>
      )}
    </div>
  );
};
