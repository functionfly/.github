import * as React from "react"
import { HelpCircle } from "lucide-react"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "./tooltip"
import { cn } from "@/lib/utils"

interface HelpTooltipProps {
  content: string
  className?: string
  side?: "top" | "right" | "bottom" | "left"
  children?: React.ReactNode
}

export function HelpTooltip({
  content,
  className,
  side = "top",
  children
}: HelpTooltipProps) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {children || (
            <HelpCircle className={cn(
              "w-4 h-4 text-text-muted hover:text-text-accent cursor-help shrink-0",
              className
            )} />
          )}
        </TooltipTrigger>
        <TooltipContent side={side} className="max-w-xs">
          <p className="text-sm">{content}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
