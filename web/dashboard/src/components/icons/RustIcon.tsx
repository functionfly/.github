import React from 'react';

export const RustIcon: React.FC<{ className?: string }> = ({ className = "w-6 h-6" }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <title>Rust</title>
    <path
      d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm0 2.182a9.818 9.818 0 110 19.636 9.818 9.818 0 010-19.636z"
      fill="#DEA584"
    />
    <path
      d="M12 3.273a8.727 8.727 0 100 17.454 8.727 8.727 0 000-17.454zm3.6 4.909l-1.2 2.182H9.6l-1.2-2.182h7.2zm-4.2 3.273h3l1.8 3.272-3.3 3.273-3.3-3.273 1.8-3.272z"
      fill="#DEA584"
    />
    <circle cx="12" cy="12" r="2.5" fill="#DEA584"/>
  </svg>
);
