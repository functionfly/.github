// Optimized illustration component for Predictive Routing
export const PredictiveRoutingIllustration = () => (
  <svg
    width="24"
    height="24"
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="w-6 h-6"
    role="img"
    aria-label="Predictive routing network visualization"
  >
    {/* Network nodes - optimized for performance */}
    <circle cx="12" cy="12" r="3" fill="currentColor" />
    <circle cx="36" cy="12" r="3" fill="currentColor" />
    <circle cx="24" cy="36" r="3" fill="currentColor" />

    {/* Connection lines with arrows - simplified paths */}
    <path d="M15 12h18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    <path d="M31.5 9.5 34.5 12.5 31.5 15.5" stroke="currentColor" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round" />

    <path d="M24 33V15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    <path d="M21 36 24 33 27 36" stroke="currentColor" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round" />

    <path d="M15 15 21 33" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    <path d="M18 30 21 33 18 36" stroke="currentColor" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round" />

    {/* AI brain icon in center - simplified */}
    <circle cx="24" cy="24" r="8" fill="currentColor" opacity="0.2" />
    <circle cx="23.5" cy="23.5" r="3.5" fill="currentColor" />
    <circle cx="23.5" cy="22" r="0.5" fill="white" />
  </svg>
);

// Optimized illustration for Global Edge Network
export const GlobalNetworkIllustration = () => (
  <svg
    width="24"
    height="24"
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="w-6 h-6"
    role="img"
    aria-label="Global edge network visualization"
  >
    {/* Globe outline - simplified */}
    <circle cx="24" cy="24" r="16" stroke="currentColor" strokeWidth="2" fill="none" strokeLinecap="round" />
    <path d="M8 24h32" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    <path d="M24 8v32" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />

    {/* Edge nodes - optimized */}
    <circle cx="12" cy="12" r="2" fill="currentColor" />
    <circle cx="36" cy="12" r="2" fill="currentColor" />
    <circle cx="12" cy="36" r="2" fill="currentColor" />
    <circle cx="36" cy="36" r="2" fill="currentColor" />

    {/* Connection lines - simplified */}
    <path d="M14 14 22 22" stroke="currentColor" strokeWidth="1.5" opacity="0.7" strokeLinecap="round" />
    <path d="M34 14 26 22" stroke="currentColor" strokeWidth="1.5" opacity="0.7" strokeLinecap="round" />
    <path d="M14 34 22 26" stroke="currentColor" strokeWidth="1.5" opacity="0.7" strokeLinecap="round" />
    <path d="M34 34 26 26" stroke="currentColor" strokeWidth="1.5" opacity="0.7" strokeLinecap="round" />
  </svg>
);