"use client"

import { useCallback, useEffect, useState } from 'react'
import { loadStripe, type Stripe, type StripeElements } from '@stripe/stripe-js'
import {
  createSetupIntent,
  listPaymentMethods,
  removePaymentMethod,
  setDefaultPaymentMethod,
  type PaymentMethod,
} from '@/api/billing'
import { Button } from '@/components/ui/button'
import { CardBrandIcon } from '@/components/ui/payment-method-card'
import { HelpTooltip } from '@/components/ui/help-tooltip'
import { Loader2, Plus, Trash2, Star, CreditCard, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY || '')

interface PaymentMethodManagerProps {
  returnUrl?: string
  className?: string
}

export function PaymentMethodManager({
  returnUrl,
  className,
}: PaymentMethodManagerProps) {
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([])
  const [loading, setLoading] = useState(true)
  const [addingCard, setAddingCard] = useState(false)
  const [removingId, setRemovingId] = useState<string | null>(null)
  const [settingDefaultId, setSettingDefaultId] = useState<string | null>(null)
  const [stripe, setStripe] = useState<Stripe | null>(null)
  const [elements, setElements] = useState<StripeElements | null>(null)
  const [clientSecret, setClientSecret] = useState<string>('')
  const [cardComplete, setCardComplete] = useState(false)
  const [cardError, setCardError] = useState<string | null>(null)
  const [processing, setProcessing] = useState(false)

  const fetchPaymentMethods = useCallback(async () => {
    setLoading(true)
    try {
      const data = await listPaymentMethods()
      setPaymentMethods(data.payment_methods || [])
    } catch {
      toast.error('Failed to load payment methods')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchPaymentMethods()
  }, [fetchPaymentMethods])

  const initializeStripeElements = useCallback(async () => {
    const stripeInstance = await stripePromise
    if (!stripeInstance) {
      setCardError('Stripe is not configured. Please add VITE_STRIPE_PUBLISHABLE_KEY to your environment.')
      return
    }

    try {
      const result = await createSetupIntent()
      setClientSecret(result.client_secret)

      const elementsInstance = stripeInstance.elements({
        clientSecret: result.client_secret,
        appearance: {
          theme: 'stripe',
          variables: {
            colorPrimary: '#6366f1',
            colorBackground: '#1a1a2e',
            colorText: '#e2e8f0',
            colorDanger: '#ef4444',
            fontFamily: 'system-ui, sans-serif',
            borderRadius: '8px',
          },
        },
      })

      setStripe(stripeInstance)
      setElements(elementsInstance)

      const cardElement = elementsInstance.getElement('card')
      if (cardElement) {
        cardElement.on('change', (event) => {
          setCardComplete(event.complete)
          setCardError(event.error?.message || null)
        })
      }
    } catch {
      toast.error('Failed to initialize payment form')
      setAddingCard(false)
    }
  }, [])

  const handleAddCard = useCallback(async () => {
    setAddingCard(true)
    setCardError(null)
    await initializeStripeElements()
  }, [initializeStripeElements])

  const handleCancelAdd = useCallback(() => {
    setAddingCard(false)
    setElements(null)
    setStripe(null)
    setClientSecret('')
    setCardComplete(false)
    setCardError(null)
  }, [])

  const handleConfirmCard = useCallback(async () => {
    if (!stripe || !elements || !clientSecret) return

    const cardElement = elements.getElement('card')
    if (!cardElement) return

    setProcessing(true)
    setCardError(null)

    try {
      const { error, setupIntent } = await stripe.confirmCardSetup(clientSecret, {
        payment_method: {
          card: cardElement,
        },
      })

      if (error) {
        setCardError(error.message || 'Failed to add card')
        toast.error(error.message || 'Failed to add card')
      } else if (setupIntent && setupIntent.status === 'succeeded') {
        toast.success('Card added successfully')
        handleCancelAdd()
        fetchPaymentMethods()
      }
    } catch {
      toast.error('Failed to add card')
    } finally {
      setProcessing(false)
    }
  }, [stripe, elements, clientSecret, handleCancelAdd, fetchPaymentMethods])

  const handleRemoveCard = useCallback(
    async (paymentMethodId: string) => {
      setRemovingId(paymentMethodId)
      try {
        await removePaymentMethod(paymentMethodId)
        toast.success('Card removed')
        fetchPaymentMethods()
      } catch {
        toast.error('Failed to remove card')
      } finally {
        setRemovingId(null)
      }
    },
    [fetchPaymentMethods]
  )

  const handleSetDefault = useCallback(
    async (paymentMethodId: string) => {
      setSettingDefaultId(paymentMethodId)
      try {
        await setDefaultPaymentMethod(paymentMethodId)
        toast.success('Default card updated')
        fetchPaymentMethods()
      } catch {
        toast.error('Failed to update default card')
      } finally {
        setSettingDefaultId(null)
      }
    },
    [fetchPaymentMethods]
  )

  const isExpired = (expMonth: number, expYear: number) => {
    return new Date(expYear, expMonth - 1) < new Date()
  }

  return (
    <div className={cn('space-y-4', className)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h3 className="font-medium text-text-primary">Payment Methods</h3>
          <HelpTooltip
            content="Add, remove, or set a default payment method. Card details are securely processed via Stripe."
            side="top"
          />
        </div>
        {!addingCard && (
          <Button size="sm" variant="outline" onClick={handleAddCard} className="gap-1.5">
            <Plus className="w-4 h-4" />
            Add Card
          </Button>
        )}
      </div>

      {loading ? (
        <div className="flex items-center justify-center p-6">
          <Loader2 className="w-5 h-5 animate-spin text-text-muted" />
        </div>
      ) : paymentMethods.length === 0 && !addingCard ? (
        <div className="flex flex-col items-center justify-center p-6 rounded-lg border border-dashed border-border-default bg-bg-secondary">
          <CreditCard className="w-8 h-8 text-text-muted mb-2" />
          <p className="text-sm text-text-muted text-center">No payment methods saved</p>
          <Button size="sm" variant="outline" onClick={handleAddCard} className="mt-3 gap-1.5">
            <Plus className="w-4 h-4" />
            Add your first card
          </Button>
        </div>
      ) : (
        <div className="space-y-2">
          {paymentMethods.map((pm) => (
            <div
              key={pm.stripe_payment_method_id}
              className={cn(
                'flex items-center justify-between p-3 rounded-lg border',
                pm.is_default
                  ? 'border-brand-500 bg-brand-500/10'
                  : 'border-border-default bg-bg-secondary'
              )}
            >
              <div className="flex items-center gap-3">
                <CardBrandIcon brand={pm.brand} size="md" />
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-text-primary">
                      •••• {pm.last4}
                    </span>
                    {pm.is_default && (
                      <span className="text-xs text-brand-500 font-medium">Default</span>
                    )}
                  </div>
                  {pm.exp_month && pm.exp_year && (
                    <p
                      className={cn(
                        'text-xs',
                        isExpired(pm.exp_month, pm.exp_year) ? 'text-red-400' : 'text-text-muted'
                      )}
                    >
                      {isExpired(pm.exp_month, pm.exp_year) ? 'Expired' : 'Expires'}{' '}
                      {pm.exp_month.toString().padStart(2, '0')}/{pm.exp_year}
                    </p>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2">
                {!pm.is_default && (
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => handleSetDefault(pm.stripe_payment_method_id!)}
                    disabled={settingDefaultId === pm.stripe_payment_method_id}
                    className="gap-1 text-xs"
                  >
                    {settingDefaultId === pm.stripe_payment_method_id ? (
                      <Loader2 className="w-3 h-3 animate-spin" />
                    ) : (
                      <Star className="w-3 h-3" />
                    )}
                    Set default
                  </Button>
                )}
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleRemoveCard(pm.stripe_payment_method_id!)}
                  disabled={removingId === pm.stripe_payment_method_id || pm.is_default}
                  className="gap-1 text-xs text-red-400 hover:text-red-300 hover:bg-red-500/10"
                >
                  {removingId === pm.stripe_payment_method_id ? (
                    <Loader2 className="w-3 h-3 animate-spin" />
                  ) : (
                    <Trash2 className="w-3 h-3" />
                  )}
                  Remove
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {addingCard && elements && (
        <div className="p-4 rounded-lg border border-border-default bg-bg-secondary space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-text-primary">Add New Card</p>
            <Button size="sm" variant="ghost" onClick={handleCancelAdd}>
              Cancel
            </Button>
          </div>

          <div className="p-3 rounded-lg border border-border-default bg-bg-tertiary">
            <div id="card-element" />
          </div>

          {cardError && (
            <div className="flex items-center gap-2 p-2 rounded-md bg-red-500/10 border border-red-500/20">
              <AlertCircle className="w-4 h-4 text-red-400 shrink-0" />
              <p className="text-sm text-red-400">{cardError}</p>
            </div>
          )}

          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={handleCancelAdd} disabled={processing}>
              Cancel
            </Button>
            <Button
              onClick={handleConfirmCard}
              disabled={!cardComplete || processing}
              className="gap-1.5"
            >
              {processing ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Adding...
                </>
              ) : (
                <>
                  <Plus className="w-4 h-4" />
                  Add Card
                </>
              )}
            </Button>
          </div>
        </div>
      )}

      {addingCard && !elements && (
        <div className="flex items-center justify-center p-4">
          <Loader2 className="w-5 h-5 animate-spin text-text-muted" />
          <span className="ml-2 text-sm text-text-muted">Loading payment form...</span>
        </div>
      )}
    </div>
  )
}