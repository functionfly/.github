import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { CreditCard, DollarSign, Users, TrendingUp, Plus, MoreVertical, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { StatCard } from "@/components/common/StatCard";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { billingApi, type PricingTier, type Subscription, type Invoice, type Coupon } from "@/api/admin";

export function AdminBillingPage() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("overview");
  const queryClient = useQueryClient();

  const { data: tiersData, isLoading: tiersLoading } = useQuery({
    queryKey: ['pricing-tiers'],
    queryFn: () => billingApi.listPricingTiers(),
  });

  const { data: subscriptionsData, isLoading: subscriptionsLoading } = useQuery({
    queryKey: ['subscriptions'],
    queryFn: () => billingApi.listSubscriptions(),
  });

  const { data: invoicesData, isLoading: invoicesLoading } = useQuery({
    queryKey: ['invoices'],
    queryFn: () => billingApi.listInvoices(),
  });

  const { data: couponsData, isLoading: couponsLoading } = useQuery({
    queryKey: ['coupons'],
    queryFn: () => billingApi.listCoupons(),
  });

  const tiers = tiersData?.tiers || [];
  const subscriptions = subscriptionsData?.subscriptions || [];
  const invoices = invoicesData?.invoices || [];
  const coupons = couponsData?.coupons || [];

  const formatCurrency = (cents: number) => {
    return `$${(cents / 100).toFixed(2)}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusBadgeColor = (status: string) => {
    switch (status) {
      case "paid":
        return "bg-emerald-500/10 text-emerald-400";
      case "open":
        return "bg-blue-500/10 text-blue-400";
      case "active":
        return "bg-green-500/10 text-green-400";
      case "canceled":
        return "bg-red-500/10 text-red-400";
      default:
        return "bg-gray-500/10 text-text-secondary";
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">Billing</h1>
          <p className="text-text-secondary">Manage pricing, subscriptions, invoices, and revenue</p>
        </div>
        <div></div> {/* Empty div to balance the layout */}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Monthly Revenue"
          value={`$${(invoices.filter(inv => inv.status === 'paid').reduce((sum, inv) => sum + inv.amount_paid_cents, 0) / 100).toFixed(2)}`}
          change={{ value: 0, label: "from last month" }}
          icon={<DollarSign className="w-5 h-5 text-[#6366f1]" />}
          trend="neutral"
        />
        <StatCard
          title="Active Subscriptions"
          value={subscriptions.filter(sub => sub.status === 'active').length}
          change={{ value: 0, label: "from last month" }}
          icon={<CreditCard className="w-5 h-5 text-[#6366f1]" />}
          trend="neutral"
        />
        <StatCard
          title="Paying Tenants"
          value={new Set(subscriptions.map(sub => sub.tenant_id)).size}
          change={{ value: 0, label: "from last month" }}
          icon={<Users className="w-5 h-5 text-[#6366f1]" />}
          trend="neutral"
        />
        <StatCard
          title="Total Invoices"
          value={invoices.length}
          change={{ value: 0, label: "from last month" }}
          icon={<TrendingUp className="w-5 h-5 text-[#6366f1]" />}
          trend="neutral"
        />
      </div>

      {/* Main Content */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="bg-bg-secondary border border-white/8">
          <TabsTrigger value="overview" className="text-text-primary data-[state=active]:bg-[#6366f1]">Overview</TabsTrigger>
          <TabsTrigger value="tiers" className="text-text-primary data-[state=active]:bg-[#6366f1]">Pricing Tiers</TabsTrigger>
          <TabsTrigger value="subscriptions" className="text-text-primary data-[state=active]:bg-[#6366f1]">Subscriptions</TabsTrigger>
          <TabsTrigger value="invoices" className="text-text-primary data-[state=active]:bg-[#6366f1]">Invoices</TabsTrigger>
          <TabsTrigger value="coupons" className="text-text-primary data-[state=active]:bg-[#6366f1]">Coupons</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Recent Revenue */}
            <Card>
              <CardHeader>
                <CardTitle>Revenue Overview</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex justify-between items-center">
                    <span className="text-text-secondary">This Month</span>
                    <span className="text-2xl font-bold text-text-primary">$19.00</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-text-secondary">Last Month</span>
                    <span className="text-text-primary">$0.00</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-text-secondary">Growth</span>
                    <Badge className="bg-emerald-500/10 text-emerald-400">+100%</Badge>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Subscription Status */}
            <Card>
              <CardHeader>
                <CardTitle>Subscription Status</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {tiers.map((tier) => (
                    <div key={tier.id} className="flex justify-between items-center">
                      <span className="text-text-secondary">{tier.name}</span>
                      <span className="text-text-primary">
                        {subscriptions.filter(sub => sub.pricing_tier_id === tier.id && sub.status === 'active').length}
                      </span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="tiers" className="space-y-6">
          <div className="flex justify-between items-center">
            <h2 className="text-xl font-semibold text-text-primary">Pricing Tiers</h2>
            <Button className="bg-[#6366f1] hover:bg-[#5855eb]">
              <Plus className="w-4 h-4 mr-2" />
              Add Tier
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {tiersLoading ? (
              <div className="col-span-full flex items-center justify-center py-8">
                <LoadingSpinner />
              </div>
            ) : tiers.map((tier) => (
              <Card key={tier.id}>
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div>
                      <CardTitle className="text-text-primary">{tier.name}</CardTitle>
                      <p className="text-text-secondary text-sm">{tier.description}</p>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-text-primary">
                        {formatCurrency(tier.price_cents)}
                      </div>
                      <div className="text-sm text-text-muted">per month</div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    {Object.entries(tier.features).map(([key, value]) => (
                      <div key={key} className="flex justify-between text-sm">
                        <span className="text-text-secondary capitalize">
                          {key.replace('_', ' ')}
                        </span>
                        <span className="text-text-primary">
                          {typeof value === 'number' && value === -1 ? 'Unlimited' : value as ReactNode}
                        </span>
                      </div>
                    ))}
                  </div>
                  <div className="mt-4 flex justify-between items-center">
                    <Badge className={tier.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-red-500/10 text-red-400"}>
                      {tier.is_active ? "Active" : "Inactive"}
                    </Badge>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="bg-bg-tertiary border-white/8">
                        <DropdownMenuItem className="text-text-primary hover:bg-white/5">Edit</DropdownMenuItem>
                        <DropdownMenuItem className="text-red-400 hover:bg-red-500/10">
                          {tier.is_active ? "Deactivate" : "Activate"}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="subscriptions" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>All Subscriptions ({subscriptions.length})</CardTitle>
            </CardHeader>
            <CardContent>
              {subscriptionsLoading ? (
                <div className="flex items-center justify-center py-8">
                  <LoadingSpinner />
                </div>
              ) : (
                <div className="space-y-4">
                  {subscriptions.map((subscription) => (
                    <div
                      key={subscription.id}
                      className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8"
                    >
                      <div>
                        <p className="font-medium text-text-primary">Subscription {subscription.id.slice(-8)}</p>
                        <p className="text-sm text-text-muted">
                          Tenant: {subscription.tenant_id.slice(0, 8)}... •
                          Period: {formatDate(subscription.current_period_start)} - {formatDate(subscription.current_period_end)}
                        </p>
                      </div>
                      <div className="flex items-center gap-4">
                        <Badge className={getStatusBadgeColor(subscription.status)}>
                          {subscription.status}
                        </Badge>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreVertical className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent className="bg-bg-tertiary border-white/8">
                            <DropdownMenuItem className="text-text-primary hover:bg-white/5">View Details</DropdownMenuItem>
                            <DropdownMenuItem className="text-text-primary hover:bg-white/5">Change Plan</DropdownMenuItem>
                            <DropdownMenuItem className="text-red-400 hover:bg-red-500/10">Cancel</DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="invoices" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>All Invoices ({invoices.length})</CardTitle>
            </CardHeader>
            <CardContent>
              {invoicesLoading ? (
                <div className="flex items-center justify-center py-8">
                  <LoadingSpinner />
                </div>
              ) : (
                <div className="space-y-4">
                  {invoices.map((invoice) => (
                    <div
                      key={invoice.id}
                      className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8"
                    >
                      <div>
                        <p className="font-medium text-text-primary">Invoice {invoice.id.slice(-8)}</p>
                        <p className="text-sm text-text-muted">
                          Tenant: {invoice.tenant_id.slice(0, 8)}... •
                          Period: {invoice.period_start ? formatDate(invoice.period_start) : "N/A"}
                        </p>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <div className="text-lg font-semibold text-text-primary">
                            {formatCurrency(invoice.amount_due_cents)}
                          </div>
                          <Badge className={getStatusBadgeColor(invoice.status)}>
                            {invoice.status}
                          </Badge>
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreVertical className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent className="bg-bg-tertiary border-white/8">
                            <DropdownMenuItem className="text-text-primary hover:bg-white/5">View PDF</DropdownMenuItem>
                            <DropdownMenuItem className="text-text-primary hover:bg-white/5">Send Email</DropdownMenuItem>
                            <DropdownMenuItem className="text-text-primary hover:bg-white/5">Mark Paid</DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="coupons" className="space-y-6">
          <div className="flex justify-between items-center">
            <h2 className="text-xl font-semibold text-text-primary">Coupons</h2>
            <Button className="bg-[#6366f1] hover:bg-[#5855eb]">
              <Plus className="w-4 h-4 mr-2" />
              Create Coupon
            </Button>
          </div>

          <Card>
            <CardContent className="pt-6">
              {couponsLoading ? (
                <div className="flex items-center justify-center py-8">
                  <LoadingSpinner />
                </div>
              ) : (
                <div className="space-y-4">
                  {coupons.map((coupon) => (
                    <div
                      key={coupon.id}
                      className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-medium text-text-primary">{coupon.code}</p>
                          <Badge className={coupon.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-red-500/10 text-red-400"}>
                            {coupon.is_active ? "Active" : "Inactive"}
                          </Badge>
                        </div>
                        <p className="text-sm text-text-muted">{coupon.description}</p>
                        <p className="text-xs text-text-muted">
                          {coupon.discount_value}{coupon.discount_type === "percent" ? "%" : ` USD`} off •
                          {coupon.times_redeemed} redeemed
                          {coupon.max_redemptions && ` / ${coupon.max_redemptions}`}
                        </p>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className="bg-bg-tertiary border-white/8">
                          <DropdownMenuItem className="text-text-primary hover:bg-white/5">Edit</DropdownMenuItem>
                          <DropdownMenuItem className="text-text-primary hover:bg-white/5">View Usage</DropdownMenuItem>
                          <DropdownMenuItem className="text-red-400 hover:bg-red-500/10">
                            {coupon.is_active ? "Deactivate" : "Activate"}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
