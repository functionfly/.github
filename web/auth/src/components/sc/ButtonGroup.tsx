import React from "react";

interface ButtonGroupProps {
  children: React.ReactNode;
  className?: string;
  responsive?: boolean;
  centered?: boolean;
}

export const ButtonGroup: React.FC<ButtonGroupProps> = ({
  children,
  className = "",
  responsive = true,
  centered = false,
}) => {
  return (
    <div
      className={`
        sealed-button-group
        ${responsive ? "sealed-button-group-responsive" : ""}
        ${centered ? "sealed-button-group-centered" : ""}
        ${className}
      `}
    >
      {children}
    </div>
  );
};
