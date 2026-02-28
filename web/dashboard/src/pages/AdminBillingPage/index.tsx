import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import {
  CreditCard,
  DollarSign,
  Users,
  TrendingUp,
  Plus,
  MoreVertical,
  ArrowLeft,
  RefreshCw,
  Download,
  CheckCircle,
  XCircle,
  Clock,
  Tag,
  BarChart3,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuSeparator } from "@/components/ui/dropdown-menu";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { Skeleton } from "@/components/ui/skeleton";
import { billingApi, type PricingTier, type Subscription, type Invoice, type Coupon } from "@/api/admin";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from "recharts";

const revenueHistory = [
  { month: "Aug", mrr: 1200, arr: 14400 },
  { month: "Sep", mrr: 1800, arr: 21600 },
  { month: "Oct", mrr: 1500, arr: 18000 },
  { month: "Nov", mrr: 2200, arr: 26400 },
  { month: "Dec", mrr: 2800, arr: 33600 },
  { month: "Jan", mrr: 3100, arr: 37200 },
  { month: "Feb", mrr: 3400, arr: 40800 },
];

const PIE_COLORS = ["#6366f1", "#8b5cf6", "#d946ef", "#10b981"];

const STATUS_CONFIG: Record<string, { label: string; color: string; icon: React.ComponentType<{ className?: string }> }> = {
  paid: { label: "Paid", color: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20", icon: CheckCircle },
  open: { label: "Open", color: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20", icon: Clock },
  active: { label: "Active", color: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20", icon: CheckCircle },
  canceled: { label: "Canceled", color: "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20", icon: XCircle },
  draft: { label: "Draft", color: "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20", icon: Clock },
  void: { label: "Void", color: "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20", icon: XCircle },
};

function StatusBadge({ status }: { status: string }) {
  const config = STATUS_CONFIG[status] || STATUS_CONFIG.draft;
  const Icon = config.icon;
  return (
    <Badge className={cn("text-xs border font-medium", config.color)}>
      <Icon className="w-3 h-3 mr-1" />
      {config.label}
    </Badge>
  );
}

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

  const formatCurrency = (cents: number) => `$${(cents / 100).toFixed(2)}`;
  const formatDate = (dateString: string) => new Date(dateString).toLocaleDateString();

  const totalRevenue = invoices.filter(inv => inv.status === 'paid').reduce((sum, inv) => sum + inv.amount_paid_cents, 0);
  const activeSubscriptionsCount = subscriptions.filter(sub => sub.status === 'active').length;
  const payingTenants = new Set(subscriptions.map(sub => sub.tenant_id)).size;

  // Build pie data from tiers
  const tierPieData = tiers.map((tier, i) => ({
    name: tier.name,
    value: subscriptions.filter(sub => sub.pricing_tier_id === tier.id && sub.status === 'active').length,
    color: PIE_COLORS[i % PIE_COLORS.length],
  })).filter(d => d.value > 0);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            onClick={() => navigate('/admin')}
            className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Billing</h1>
            <p className="text-sm text-text-secondary">Manage pricing, subscriptions, and revenue</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              queryClient.invalidateQueries({ queryKey: ['pricing-tiers'] });
              queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
              queryClient.invalidateQueries({ queryKey: ['invoices'] });
            }}
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          {
            title: "Monthly Revenue",
            value: formatCurrency(totalRevenue),
            icon: DollarSign,
            iconColor: "text-emerald-500",
            bg: "bg-emerald-500/10",
            sub: "from paid invoices",
          },
          {
            title: "Active Subscriptions",
            value: activeSubscriptionsCount,
            icon: CreditCard,
            iconColor: "text-purple-500",
            bg: "bg-purple-500/10",
            sub: "currently active",
          },
          {
            title: "Paying Tenants",
            value: payingTenants,
            icon: Users,
            iconColor: "text-blue-500",
            bg: "bg-blue-500/10",
            sub: "unique organizations",
          },
          {
            title: "Total Invoices",
            value: invoices.length,
            icon: TrendingUp,
            iconColor: "text-amber-500",
            bg: "bg-amber-500/10",
            sub: "all time",
          },
        ].map((stat) => (
          <Card key={stat.title} className="glass-card hover-lift">
            <CardContent className="p-5">
              <div className="flex items-center justify-between mb-2">
                <p className="text-xs font-medium text-text-secondary">{stat.title}</p>
                <div className={cn("p-1.5 rounded-lg", stat.bg)}>
                  <stat.icon className={cn("w-4 h-4", stat.iconColor)} />
                </div>
              </div>
              <p className="text-2xl font-bold text-text-primary">{stat.value}</p>
              <p className="text-xs text-text-muted mt-1">{stat.sub}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Main Content */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="bg-bg-secondary border border-border-default">
          <TabsTrigger value="overview" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white text-text-secondary">
            Overview
          </TabsTrigger>
          <TabsTrigger value="tiers" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white text-text-secondary">
            Pricing Tiers
          </TabsTrigger>
          <TabsTrigger value="subscriptions" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white text-text-secondary">
            Subscriptions
          </TabsTrigger>
          <TabsTrigger value="invoices" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white text-text-secondary">
            Invoices
          </TabsTrigger>
          <TabsTrigger value="coupons" className="data-[state=active]:bg-brand-500 data-[state=active]:text-white text-text-secondary">
            Coupons
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6 mt-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Revenue Chart */}
            <Card className="lg:col-span-2 glass-card">
              <CardHeader className="pb-2">
                <CardTitle className="text-text-primary flex items-center gap-2">
                  <BarChart3 className="w-5 h-5 text-purple-500" />
                  Revenue Trend
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={220}>
                  <AreaChart data={revenueHistory} margin={{ top: 5, right: 5, left: -10, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorMrr" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                    <XAxis dataKey="month" tick={{ fill: "var(--text-muted)", fontSize: 12 }} axisLine={false} tickLine={false} />
                    <YAxis tick={{ fill: "var(--text-muted)", fontSize: 12 }} axisLine={false} tickLine={false} tickFormatter={(v) => `$${v}`} />
                    <Tooltip
                      contentStyle={{
                        background: "var(--bg-tertiary)",
                        border: "1px solid var(--border-default)",
                        borderRadius: "8px",
                        color: "var(--text-primary)",
                        fontSize: "12px",
                      }}
                      formatter={(value) => [`$${value}`, "MRR"]}
                    />
                    <Area type="monotone" dataKey="mrr" stroke="#6366f1" strokeWidth={2} fill="url(#colorMrr)" name="MRR" />
                  </AreaChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>

            {/* Subscription Distribution */}
            <Card className="glass-card">
              <CardHeader className="pb-2">
                <CardTitle className="text-text-primary text-base">Plan Distribution</CardTitle>
              </CardHeader>
              <CardContent>
                {tierPieData.length > 0 ? (
                  <>
                    <ResponsiveContainer width="100%" height={160}>
                      <PieChart>
                        <Pie
                          data={tierPieData}
                          cx="50%"
                          cy="50%"
                          innerRadius={45}
                          outerRadius={70}
                          paddingAngle={3}
                          dataKey="value"
                        >
                          {tierPieData.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={entry.color} />
                          ))}
                        </Pie>
                        <Tooltip
                          contentStyle={{
                            background: "var(--bg-tertiary)",
                            border: "1px solid var(--border-default)",
                            borderRadius: "8px",
                            color: "var(--text-primary)",
                            fontSize: "12px",
                          }}
                        />
                      </PieChart>
                    </ResponsiveContainer>
                    <div className="space-y-2 mt-2">
                      {tierPieData.map((entry) => (
                        <div key={entry.name} className="flex items-center justify-between text-sm">
                          <div className="flex items-center gap-2">
                            <span className="w-2.5 h-2.5 rounded-full" style={{ background: entry.color }} />
                            <span className="text-text-secondary">{entry.name}</span>
                          </div>
                          <span className="font-medium text-text-primary">{entry.value}</span>
                        </div>
                      ))}
                    </div>
                  </>
                ) : (
                  <div className="flex flex-col items-center justify-center h-40 text-text-muted">
                    <CreditCard className="w-8 h-8 mb-2 opacity-30" />
                    <p className="text-sm">No active subscriptions</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Revenue Summary */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card className="glass-card">
              <CardHeader>
                <CardTitle className="text-text-primary">Revenue Summary</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {[
                  { label: "This Month", value: "$3,400", change: "+9.7%", positive: true },
                  { label: "Last Month", value: "$3,100", change: "+10.7%", positive: true },
                  { label: "YTD Revenue", value: "$19,000", change: "+45%", positive: true },
                  { label: "ARR", value: "$40,800", change: "+23%", positive: true },
                ].map((item) => (
                  <div key={item.label} className="flex items-center justify-between py-2 border-b border-border-subtle last:border-0">
                    <span className="text-text-secondary">{item.label}</span>
                    <div className="flex items-center gap-3">
                      <span className="font-semibold text-text-primary">{item.value}</span>
                      <Badge className={cn(
                        "text-xs border",
                        item.positive
                          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                          : "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20"
                      )}>
                        {item.change}
                      </Badge>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="glass-card">
              <CardHeader>
                <CardTitle className="text-text-primary">Subscription Status</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {tiers.map((tier) => {
                  const count = subscriptions.filter(sub => sub.pricing_tier_id === tier.id && sub.status === 'active').length;
                  const total = subscriptions.filter(sub => sub.pricing_tier_id === tier.id).length;
                  return (
                    <div key={tier.id} className="space-y-1.5">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-text-secondary font-medium">{tier.name}</span>
                        <span className="text-text-primary font-semibold">{count} active / {total} total</span>
                      </div>
                      <div className="h-1.5 bg-bg-secondary rounded-full overflow-hidden">
                        <div
                          className="h-full bg-brand-500 rounded-full transition-all"
                          style={{ width: total > 0 ? `${(count / total) * 100}%` : "0%" }}
                        />
                      </div>
                    </div>
                  );
                })}
                {tiers.length === 0 && (
                  <p className="text-text-muted text-sm text-center py-4">No pricing tiers configured</p>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="tiers" className="space-y-6 mt-6">
          <div className="flex justify-between items-center">
            <h2 className="text-xl font-semibold text-text-primary">Pricing Tiers</h2>
            <Button className="bg-brand-500 hover:bg-brand-600 text-white">
              <Plus className="w-4 h-4 mr-2" />
              Add Tier
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {tiersLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <Card key={i} className="glass-card">
                  <CardContent className="p-6">
                    <Skeleton className="h-6 w-32 mb-4" />
                    <Skeleton className="h-10 w-24 mb-4" />
                    <div className="space-y-2">
                      {Array.from({ length: 4 }).map((_, j) => (
                        <Skeleton key={j} className="h-4 w-full" />
                      ))}
                    </div>
                  </CardContent>
                </Card>
              ))
            ) : tiers.map((tier) => (
              <Card key={tier.id} className="glass-card hover-lift">
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div>
                      <CardTitle className="text-text-primary">{tier.name}</CardTitle>
                      <p className="text-text-secondary text-sm mt-1">{tier.description}</p>
                    </div>
                    <div className="text-right">
                      <div className="text-2xl font-bold text-text-primary">
                        {formatCurrency(tier.price_cents)}
                      </div>
                      <div className="text-xs text-text-muted">per month</div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 mb-4">
                    {Object.entries(tier.features).map(([key, value]) => (
                      <div key={key} className="flex justify-between text-sm py-1 border-b border-border-subtle last:border-0">
                        <span className="text-text-secondary capitalize">
                          {key.replace(/_/g, ' ')}
                        </span>
                        <span className="text-text-primary font-medium">
                          {typeof value === 'number' && value === -1 ? 'Unlimited' : value as ReactNode}
                        </span>
                      </div>
                    ))}
                  </div>
                  <div className="flex justify-between items-center">
                    <StatusBadge status={tier.is_active ? "active" : "canceled"} />
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="bg-bg-tertiary border-border-default" align="end">
                        <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">Edit</DropdownMenuItem>
                        <DropdownMenuSeparator className="bg-border-subtle" />
                        <DropdownMenuItem className={cn("cursor-pointer", tier.is_active ? "text-red-500 hover:bg-red-500/10" : "text-emerald-500 hover:bg-emerald-500/10")}>
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

        <TabsContent value="subscriptions" className="space-y-6 mt-6">
          <Card className="glass-card">
            <CardHeader>
              <CardTitle className="text-text-primary">
                All Subscriptions
                <span className="ml-2 text-sm font-normal text-text-muted">({subscriptions.length})</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {subscriptionsLoading ? (
                <div className="flex items-center justify-center py-12">
                  <LoadingSpinner />
                </div>
              ) : subscriptions.length === 0 ? (
                <div className="text-center py-12 text-text-muted">
                  <CreditCard className="w-10 h-10 mx-auto mb-3 opacity-30" />
                  <p>No subscriptions found</p>
                </div>
              ) : (
                <div className="divide-y divide-border-subtle">
                  {subscriptions.map((subscription) => (
                    <div
                      key={subscription.id}
                      className="flex items-center justify-between px-6 py-4 hover:bg-bg-hover transition-colors"
                    >
                      <div>
                        <p className="font-medium text-text-primary">
                          Subscription <span className="font-mono text-sm">{subscription.id.slice(-8)}</span>
                        </p>
                        <p className="text-sm text-text-muted mt-0.5">
                          Tenant: <span className="font-mono">{subscription.tenant_id.slice(0, 8)}...</span>
                          {" · "}
                          {formatDate(subscription.current_period_start)} – {formatDate(subscription.current_period_end)}
                        </p>
                      </div>
                      <div className="flex items-center gap-3">
                        <StatusBadge status={subscription.status} />
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary h-8 w-8 p-0">
                              <MoreVertical className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent className="bg-bg-tertiary border-border-default" align="end">
                            <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">View Details</DropdownMenuItem>
                            <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">Change Plan</DropdownMenuItem>
                            <DropdownMenuSeparator className="bg-border-subtle" />
                            <DropdownMenuItem className="text-red-500 hover:bg-red-500/10 cursor-pointer">Cancel</DropdownMenuItem>
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

        <TabsContent value="invoices" className="space-y-6 mt-6">
          <Card className="glass-card">
            <CardHeader>
              <CardTitle className="text-text-primary">
                All Invoices
                <span className="ml-2 text-sm font-normal text-text-muted">({invoices.length})</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {invoicesLoading ? (
                <div className="flex items-center justify-center py-12">
                  <LoadingSpinner />
                </div>
              ) : invoices.length === 0 ? (
                <div className="text-center py-12 text-text-muted">
                  <DollarSign className="w-10 h-10 mx-auto mb-3 opacity-30" />
                  <p>No invoices found</p>
                </div>
              ) : (
                <div className="divide-y divide-border-subtle">
                  {invoices.map((invoice) => (
                    <div
                      key={invoice.id}
                      className="flex items-center justify-between px-6 py-4 hover:bg-bg-hover transition-colors"
                    >
                      <div>
                        <p className="font-medium text-text-primary">
                          Invoice <span className="font-mono text-sm">{invoice.id.slice(-8)}</span>
                        </p>
                        <p className="text-sm text-text-muted mt-0.5">
                          Tenant: <span className="font-mono">{invoice.tenant_id.slice(0, 8)}...</span>
                          {invoice.period_start && ` · ${formatDate(invoice.period_start)}`}
                        </p>
                      </div>
                      <div className="flex items-center gap-3">
                        <div className="text-right">
                          <div className="font-semibold text-text-primary">
                            {formatCurrency(invoice.amount_due_cents)}
                          </div>
                          <StatusBadge status={invoice.status} />
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary h-8 w-8 p-0">
                              <MoreVertical className="w-4 h-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent className="bg-bg-tertiary border-border-default" align="end">
                            <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">View PDF</DropdownMenuItem>
                            <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">Send Email</DropdownMenuItem>
                            <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">Mark Paid</DropdownMenuItem>
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

        <TabsContent value="coupons" className="space-y-6 mt-6">
          <div className="flex justify-between items-center">
            <h2 className="text-xl font-semibold text-text-primary">Coupons</h2>
            <Button className="bg-brand-500 hover:bg-brand-600 text-white">
              <Plus className="w-4 h-4 mr-2" />
              Create Coupon
            </Button>
          </div>

          <Card className="glass-card">
            <CardContent className="p-0">
              {couponsLoading ? (
                <div className="flex items-center justify-center py-12">
                  <LoadingSpinner />
                </div>
              ) : coupons.length === 0 ? (
                <div className="text-center py-12 text-text-muted">
                  <Tag className="w-10 h-10 mx-auto mb-3 opacity-30" />
                  <p>No coupons found</p>
                </div>
              ) : (
                <div className="divide-y divide-border-subtle">
                  {coupons.map((coupon) => (
                    <div
                      key={coupon.id}
                      className="flex items-center justify-between px-6 py-4 hover:bg-bg-hover transition-colors"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="font-mono font-semibold text-text-primary">{coupon.code}</p>
                          <StatusBadge status={coupon.is_active ? "active" : "canceled"} />
                        </div>
                        <p className="text-sm text-text-muted mt-0.5">{coupon.description}</p>
                        <p className="text-xs text-text-muted mt-0.5">
                          <span className="font-medium text-text-secondary">
                            {coupon.discount_value}{coupon.discount_type === "percent" ? "%" : " USD"} off
                          </span>
                          {" · "}
                          {coupon.times_redeemed} redeemed
                          {coupon.max_redemptions && ` / ${coupon.max_redemptions} max`}
                        </p>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary h-8 w-8 p-0">
                            <MoreVertical className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className="bg-bg-tertiary border-border-default" align="end">
                          <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">Edit</DropdownMenuItem>
                          <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">View Usage</DropdownMenuItem>
                          <DropdownMenuSeparator className="bg-border-subtle" />
                          <DropdownMenuItem className={cn("cursor-pointer", coupon.is_active ? "text-red-500 hover:bg-red-500/10" : "text-emerald-500 hover:bg-emerald-500/10")}>
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
