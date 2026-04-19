"use client"

import * as React from "react"
import { ChevronDown, ChevronUp } from "lucide-react"
import { cn } from "@/lib/utils"

interface CollapsibleSectionProps {
  title: string
  icon?: React.ReactNode
  defaultOpen?: boolean
  children: React.ReactNode
  className?: string
  headerRight?: React.ReactNode
  variant?: "default" | "highlighted"
}

export function CollapsibleSection({
  title,
  icon,
  defaultOpen = false,
  children,
  className,
  headerRight,
  variant = "default",
}: CollapsibleSectionProps) {
  const [isOpen, setIsOpen] = React.useState(defaultOpen)

  return (
    <div
      className={cn(
        "rounded-lg border transition-all",
        variant === "default"
          ? "bg-bg-secondary border-border-default"
          : "bg-gradient-to-br from-amber-500/5 to-orange-500/5 border-brand-500/20",
        className
      )}
    >
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between p-4 text-left hover:bg-bg-hover/50 transition-colors rounded-lg"
      >
        <div className="flex items-center gap-2">
          {icon && <span className="text-brand-500">{icon}</span>}
          <span className="text-sm font-medium text-text-primary">{title}</span>
        </div>
        <div className="flex items-center gap-2">
          {headerRight}
          {isOpen ? (
            <ChevronUp className="w-4 h-4 text-text-muted" />
          ) : (
            <ChevronDown className="w-4 h-4 text-text-muted" />
          )}
        </div>
      </button>
      {isOpen && <div className="px-4 pb-4">{children}</div>}
    </div>
  )
}