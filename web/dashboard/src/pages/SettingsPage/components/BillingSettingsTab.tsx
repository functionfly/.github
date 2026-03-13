import { useState } from "react";
import {
  CreditCard,
  Calendar,
  Download,
  AlertCircle,
  Building2,
  Trash2,
} from "lucide-react";
import {
  createBillingPortalSession,
  getBillingPortalErrorMessage,
  getSubscription,
  getSubscriptionErrorMessage,
  listInvoices,
  getInvoicesErrorMessage,
  cancelSubscription,
  type Subscription,
  type Invoice,
} from "@/api/billing";
import { EnterpriseSettingsSection } from "@/components/enterprise";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { useAuthStore } from "@/stores/authStore";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { formatDate, formatCurrency } from "../settings-utils";

export interface BillingSettingsTabProps {
  /** Return URL after Stripe billing portal (e.g. current page). */
  returnUrl: string;
  /** Plan name to show when there is no subscription (from parent me query). */
  displayPlan: string;
}

export function BillingSettingsTab({ returnUrl, displayPlan }: BillingSettingsTabProps) {
  const user = useAuthStore((state) => state.user);
  const [billingPortalLoading, setBillingPortalLoading] = useState(false);
  const [contactModalOpen, setContactModalOpen] = useState(false);
  const [contactForm, setContactForm] = useState({
    name: user?.name || "",
    email: user?.email || "",
    company: "",
    message: "",
  });
  const [contactSubmitting, setContactSubmitting] = useState(false);
  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [cancelSubmitting, setCancelSubmitting] = useState(false);

  const { data: subscriptionData, isLoading: subscriptionLoading, error: subscriptionError } = useQuery({
    queryKey: ["billing", "subscription"],
    queryFn: () => getSubscription(),
    retry: false,
  });

  const { data: invoicesData, isLoading: invoicesLoading, error: invoicesError } = useQuery({
    queryKey: ["billing", "invoices"],
    queryFn: () => listInvoices(10, 0),
    retry: false,
  });

  const subscription = subscriptionData as Subscription | undefined;
  const invoices = (invoicesData as { invoices: Invoice[] })?.invoices ?? [];

  const handleContactSales = async () => {
    if (!contactForm.name || !contactForm.email || !contactForm.message) {
      toast.error("Please fill in all required fields");
      return;
    }
    setContactSubmitting(true);
    try {
      const subject = encodeURIComponent(`Enterprise Plan Inquiry from ${contactForm.name}`);
      const body = encodeURIComponent(
        `Name: ${contactForm.name}\nEmail: ${contactForm.email}\nCompany: ${contactForm.company || "Not provided"}\n\nMessage:\n${contactForm.message}`
      );
      window.location.href = `mailto:sales@functionfly.com?subject=${subject}&body=${body}`;
      setContactModalOpen(false);
      toast.success("Thank you for your interest! Our sales team will contact you soon.");
    } catch {
      toast.error("Failed to submit contact request. Please email us directly at sales@functionfly.com");
    } finally {
      setContactSubmitting(false);
    }
  };

  const handleCancelSubscription = async () => {
    setCancelSubmitting(true);
    try {
      await cancelSubscription(false);
      toast.success("Subscription will be cancelled at the end of the billing period");
      setCancelModalOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to cancel subscription");
    } finally {
      setCancelSubmitting(false);
    }
  };

  const openPortal = async (urlPath: string) => {
    setBillingPortalLoading(true);
    try {
      const { url } = await createBillingPortalSession(urlPath);
      window.location.href = url;
    } catch (e: unknown) {
      setBillingPortalLoading(false);
      toast.error(getBillingPortalErrorMessage(e));
    }
  };

  return (
    <div className="space-y-6">
      <EnterpriseSettingsSection />
      <Card>
        <CardHeader>
          <CardTitle>Current Plan</CardTitle>
          <CardDescription className="text-text-secondary">Manage your subscription</CardDescription>
        </CardHeader>
        <CardContent>
          {subscriptionLoading ? (
            <div className="flex items-center justify-center p-4">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-[#6366f1]" />
            </div>
          ) : subscriptionError ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <AlertCircle className="w-5 h-5 text-amber-500" />
              <p className="text-amber-500 text-sm">{getSubscriptionErrorMessage(subscriptionError)}</p>
            </div>
          ) : subscription ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-border-default">
                <div>
                  <h3 className="font-semibold text-text-primary capitalize">{subscription.plan} Plan</h3>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge
                      variant={subscription.status === "active" ? "default" : "secondary"}
                      className={
                        subscription.status === "active"
                          ? "bg-green-500/20 text-green-600 dark:text-green-400 border border-green-500/40 dark:border-green-500/30"
                          : ""
                      }
                    >
                      {subscription.status}
                    </Badge>
                    {subscription.cancel_at_period_end && (
                      <Badge variant="outline" className="border-amber-500/50 text-amber-600 dark:text-amber-400">
                        Cancels at period end
                      </Badge>
                    )}
                  </div>
                </div>
                <Badge>Current</Badge>
              </div>
              {(subscription.current_period_start || subscription.current_period_end) && (
                <div className="grid grid-cols-2 gap-4 p-4 rounded-lg bg-bg-secondary border border-border-default">
                  <div className="flex items-center gap-3">
                    <Calendar className="w-5 h-5 text-text-muted" />
                    <div>
                      <p className="text-sm text-text-muted">Current Period Start</p>
                      <p className="text-text-primary font-medium">{formatDate(subscription.current_period_start)}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Calendar className="w-5 h-5 text-text-muted" />
                    <div>
                      <p className="text-sm text-text-muted">Next Billing Date</p>
                      <p className="text-text-primary font-medium">{formatDate(subscription.current_period_end)}</p>
                    </div>
                  </div>
                </div>
              )}
              {subscription.payment_method && (
                <div className="p-4 rounded-lg bg-bg-secondary border border-border-default">
                  <p className="text-sm text-text-muted mb-2">Payment Method</p>
                  <div className="flex items-center gap-3">
                    <CreditCard className="w-5 h-5 text-[#6366f1]" />
                    <div>
                      <p className="text-text-primary font-medium capitalize">
                        {subscription.payment_method.brand} •••• {subscription.payment_method.last4}
                      </p>
                      {subscription.payment_method.exp_month && subscription.payment_method.exp_year && (
                        <p className="text-sm text-text-muted">
                          Expires {subscription.payment_method.exp_month}/{subscription.payment_method.exp_year}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="flex items-center justify-between p-4 rounded-lg bg-linear-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-border-default">
              <div>
                <h3 className="font-semibold text-text-primary capitalize">{displayPlan} Plan</h3>
                <p className="text-sm text-text-secondary">
                  {displayPlan === "free" ? "Free forever" : "Active subscription"}
                </p>
              </div>
              <Badge>Current</Badge>
            </div>
          )}
          <div className="mt-6 flex flex-wrap gap-3">
            <Button
              variant="default"
              onClick={() => openPortal(returnUrl)}
              disabled={billingPortalLoading}
            >
              {billingPortalLoading ? "Opening…" : "Manage billing"}
            </Button>
            <Button
              variant="outline"
              className="settings-upgrade-btn border-border-strong"
              disabled={billingPortalLoading}
              onClick={() => openPortal(`${window.location.origin}/pricing`)}
            >
              {billingPortalLoading ? "Opening…" : "Upgrade Plan"}
            </Button>
            <Dialog open={contactModalOpen} onOpenChange={setContactModalOpen}>
              <DialogTrigger asChild>
                <Button
                  variant="outline"
                  className="border-border-strong border-[#6366f1]/50 text-[#6366f1] hover:bg-[#6366f1]/10"
                >
                  <Building2 className="w-4 h-4 mr-2" />
                  Contact Sales
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                  <DialogTitle>Contact Sales</DialogTitle>
                  <DialogDescription>
                    Interested in our Enterprise plan? Fill out the form below and our team will get back to you
                    within 24 hours.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="contact-name">Name *</Label>
                    <Input
                      id="contact-name"
                      value={contactForm.name}
                      onChange={(e) => setContactForm({ ...contactForm, name: e.target.value })}
                      placeholder="Your name"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="contact-email">Email *</Label>
                    <Input
                      id="contact-email"
                      type="email"
                      value={contactForm.email}
                      onChange={(e) => setContactForm({ ...contactForm, email: e.target.value })}
                      placeholder="your@email.com"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="contact-company">Company</Label>
                    <Input
                      id="contact-company"
                      value={contactForm.company}
                      onChange={(e) => setContactForm({ ...contactForm, company: e.target.value })}
                      placeholder="Your company name"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="contact-message">Message *</Label>
                    <Textarea
                      id="contact-message"
                      value={contactForm.message}
                      onChange={(e) => setContactForm({ ...contactForm, message: e.target.value })}
                      placeholder="Tell us about your needs..."
                      rows={4}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setContactModalOpen(false)}>
                    Cancel
                  </Button>
                  <Button
                    onClick={handleContactSales}
                    disabled={contactSubmitting}
                    className="bg-[#6366f1] hover:bg-[#6366f1]/90"
                  >
                    {contactSubmitting ? "Sending..." : "Send Message"}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
            {subscription && subscription.status === "active" && !subscription.cancel_at_period_end && (
              <Dialog open={cancelModalOpen} onOpenChange={setCancelModalOpen}>
                <DialogTrigger asChild>
                  <Button
                    variant="outline"
                    className="border-border-strong border-red-500/50 text-red-400 hover:bg-red-500/10"
                  >
                    <Trash2 className="w-4 h-4 mr-2" />
                    Cancel Subscription
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[500px]">
                  <DialogHeader>
                    <DialogTitle>Cancel Subscription</DialogTitle>
                    <DialogDescription>
                      Are you sure you want to cancel your subscription? You'll lose access to premium features at
                      the end of your billing period.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="py-4">
                    <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
                      <p className="text-amber-400 text-sm">
                        Your subscription will remain active until {formatDate(subscription.current_period_end)}. After
                        that, you'll be downgraded to the free plan.
                      </p>
                    </div>
                  </div>
                  <DialogFooter>
                    <Button variant="outline" onClick={() => setCancelModalOpen(false)}>
                      Keep Subscription
                    </Button>
                    <Button
                      onClick={handleCancelSubscription}
                      disabled={cancelSubmitting}
                      variant="destructive"
                    >
                      {cancelSubmitting ? "Cancelling..." : "Confirm Cancellation"}
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Invoices</CardTitle>
          <CardDescription className="text-text-secondary">
            View and download your past invoices
          </CardDescription>
        </CardHeader>
        <CardContent>
          {invoicesLoading ? (
            <div className="flex items-center justify-center p-4">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-[#6366f1]" />
            </div>
          ) : invoicesError ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <AlertCircle className="w-5 h-5 text-amber-500" />
              <p className="text-amber-500 text-sm">{getInvoicesErrorMessage(invoicesError)}</p>
            </div>
          ) : invoices.length === 0 ? (
            <div className="text-center p-6">
              <CreditCard className="w-12 h-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No invoices yet</p>
              <p className="text-sm text-text-muted">Your invoices will appear here after your first payment</p>
            </div>
          ) : (
            <div className="space-y-3">
              {invoices.map((invoice) => (
                <div
                  key={invoice.id}
                  className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-default hover:border-border-strong transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-[#6366f1]/20 flex items-center justify-center">
                      <CreditCard className="w-5 h-5 text-[#6366f1]" />
                    </div>
                    <div>
                      <p className="font-medium text-text-primary">{formatCurrency(invoice.amount, invoice.currency)}</p>
                      <p className="text-sm text-text-muted">{formatDate(invoice.invoice_date)}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge
                      variant={invoice.status === "paid" ? "default" : "secondary"}
                      className={
                        invoice.status === "paid" ? "bg-green-500/20 text-green-400 border-green-500/30" : ""
                      }
                    >
                      {invoice.status}
                    </Badge>
                    {(invoice.invoice_pdf || invoice.hosted_invoice_url) && (
                      <Button variant="ghost" size="sm" asChild>
                        <a
                          href={invoice.invoice_pdf || invoice.hosted_invoice_url || "#"}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1"
                        >
                          <Download className="w-4 h-4" />
                          Download
                        </a>
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
