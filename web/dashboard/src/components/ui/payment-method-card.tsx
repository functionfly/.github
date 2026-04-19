"use client"

import { CreditCard } from "lucide-react"
import { cn } from "@/lib/utils"

// Card brand configurations
const CARD_BRANDS = {
  visa: { color: "#1A1F71", letter: "VISA" },
  mastercard: { color: "#EB001B", letter: "MC" },
  amex: { color: "#006FCF", letter: "AMEX" },
  discover: { color: "#FF6000", letter: "DISC" },
  diners: { color: "#0079A0", letter: "DIN" },
  jcb: { color: "#003087", letter: "JCB" },
  unionpay: { color: "#E21836", letter: "UN" },
  default: { color: "#6B7280", letter: "" },
} as const

interface CardBrandIconProps {
  brand: string
  className?: string
  showIcon?: boolean
  size?: "sm" | "md" | "lg"
}

export function CardBrandIcon({
  brand,
  className,
  showIcon = true,
  size = "md",
}: CardBrandIconProps) {
  const normalizedBrand = brand.toLowerCase().replace(/[^a-z]/g, "")
  const brandConfig =
    (Object.keys(CARD_BRANDS).find((key) => normalizedBrand.includes(key)) as keyof typeof CARD_BRANDS) ||
    "default"
  const config = CARD_BRANDS[brandConfig]

  const sizeClasses = {
    sm: "w-8 h-6 text-[8px]",
    md: "w-10 h-7 text-[10px]",
    lg: "w-12 h-8 text-xs",
  }

  return (
    <div
      className={cn(
        "relative flex items-center justify-center rounded border border-border-subtle bg-bg-secondary font-bold",
        sizeClasses[size],
        className
      )}
      style={{ color: config.color }}
      title={brand}
    >
      {config.letter || <CreditCard className="w-4 h-4" />}
      {showIcon && config.letter && (
        <div
          className="absolute inset-0 opacity-10 rounded"
          style={{ background: config.color }}
        />
      )}
    </div>
  )
}

interface PaymentMethodCardProps {
  brand: string
  last4: string
  expMonth?: number
  expYear?: number
  onUpdate?: () => void
  className?: string
}

export function PaymentMethodCard({
  brand,
  last4,
  expMonth,
  expYear,
  onUpdate,
  className,
}: PaymentMethodCardProps) {
  const isExpired =
    expMonth &&
    expYear &&
    new Date(expYear, expMonth - 1) < new Date()

  return (
    <div
      className={cn(
        "p-4 rounded-lg bg-bg-secondary border border-border-default",
        className
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <CardBrandIcon brand={brand} size="lg" />
          <div>
            <p className="text-sm font-medium text-text-primary">
              •••• {last4}
            </p>
            {expMonth && expYear && (
              <p className={cn("text-xs text-text-muted", isExpired && "text-red-400")}>
                {isExpired ? "Expired" : "Expires"} {expMonth.toString().padStart(2, "0")}/{expYear}
              </p>
            )}
          </div>
        </div>
        {onUpdate && (
          <button
            type="button"
            onClick={onUpdate}
            className="text-xs text-brand-500 hover:text-brand-500/80 transition-colors"
          >
            Update
          </button>
        )}
      </div>
    </div>
  )
}