import { cn } from "@/lib/utils";

export interface ProtocolVersionTagProps {
  /** Protocol version */
  version: string;
  /** Whether it's the latest version */
  latest?: boolean;
  /** Custom className */
  className?: string;
  /** Click handler */
  onClick?: () => void;
}

export function ProtocolVersionTag({
  version,
  latest = false,
  className,
  onClick,
}: ProtocolVersionTagProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 text-xs font-medium px-2 py-0.5 rounded-md border",
        latest
          ? "bg-brand-500/10 text-brand-500 border-brand-500/20"
          : "bg-bg-secondary text-muted-foreground border-border-subtle",
        onClick && "cursor-pointer hover:opacity-80 transition-opacity",
        className
      )}
      onClick={onClick}
    >
      <span className="font-mono">{version}</span>
      {latest && (
        <span className="text-[10px] uppercase tracking-wider">latest</span>
      )}
    </span>
  );
}
