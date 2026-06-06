// ReceiptPoweredBy — the subtle branding badge. Per the spec: "subtle, not
// obnoxious" and "make it a link." Whole pill is a link; opacity dips to
// 70% at rest and rises to 100% on hover. No animation (per the plan).
//
// Only rendered on the full-page variant. The embed page omits it.

const POWERED_BY_HREF = "https://functionfly.com/?utm_source=receipt&utm_medium=virality";

export function ReceiptPoweredBy() {
  return (
    <a
      href={POWERED_BY_HREF}
      target="_blank"
      rel="noopener noreferrer sponsored"
      aria-label="Powered by FunctionFly"
      data-testid="receipt-powered-by"
      className="fixed bottom-4 right-4 z-30 select-none rounded-full border border-border/40 bg-background/80 px-3 py-1.5 text-xs opacity-70 shadow-sm backdrop-blur transition-opacity duration-150 hover:opacity-100"
    >
      Powered by <span className="font-semibold">FunctionFly</span>
    </a>
  );
}

export default ReceiptPoweredBy;
