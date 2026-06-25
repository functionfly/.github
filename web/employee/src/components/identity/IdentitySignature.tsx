interface IdentitySignatureProps {
  signature: string;
  className?: string;
}

export function IdentitySignature({ signature, className = '' }: IdentitySignatureProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-md bg-blue-600/20 px-2.5 py-1 font-mono text-sm font-bold tracking-wider text-blue-400 ${className}`}
    >
      <span className="text-lg leading-none">&#x2314;</span>
      {signature.replace(/^⨍/, '')}
    </span>
  );
}
