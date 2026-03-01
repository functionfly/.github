import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  CheckCircle,
  XCircle,
  AlertTriangle,
  Shield,
  Zap,
  BarChart3,
  Headphones,
  Globe,
  Lock,
  FileText,
  Database,
  Clock,
  Users,
  Layers,
  Rocket,
  Crown,
  Star,
  ChevronRight,
  RefreshCw,
  Search,
  Filter,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { Skeleton } from "@/components/ui/skeleton";
import { featuresApi, type Feature, type PlanInfo } from "@/api/admin";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// Icon mapping for feature categories
const CATEGORY_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  core: Zap,
  security: Lock,
  analytics: BarChart3,
  support: Headphones,
};

// Plan colors
const PLAN_COLORS: Record<string, { bg: string; text: string; border: string; icon: React.ComponentType<{ className?: string }> }> = {
  starter: { bg: "bg-slate-500/10", text: "text-slate-600 dark:text-slate-400", border: "border-slate-500/20", icon: Star },
  pro: { bg: "bg-blue-500/10", text: "text-blue-600 dark:text-blue-400", border: "border-blue-500/20", icon: Rocket },
  enterprise: { bg: "bg-purple-500/10", text: "text-purple-600 dark:text-purple-400", border: "border-purple-500/20", icon: Crown },
  agent_starter: { bg: "bg-green-500/10", text: "text-green-600 dark:text-green-400", border: "border-green-500/20", icon: Users },
  agent_scale: { bg: "bg-emerald-500/10", text: "text-emerald-600 dark:text-emerald-400", border: "border-emerald-500/20", icon: Layers },
  agent_pro: { bg: "bg-indigo-500/10", text: "text-indigo-600 dark:text-indigo-400", border: "border-indigo-500/20", icon: Globe },
  agent_enterprise: { bg: "bg-amber-500/10", text: "text-amber-600 dark:text-amber-400", border: "border-amber-500/20", icon: Lock },
};

function FeatureCard({ feature, availablePlans }: { feature: Feature; availablePlans: string[] }) {
  const CategoryIcon = CATEGORY_ICONS[feature.category] || Zap;
  
  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3">
            <div className="p-2 rounded-lg bg-primary/10">
              <CategoryIcon className="w-5 h-5 text-primary" />
            </div>
            <div>
              <h3 className="font-semibold">{feature.name}</h3>
              <p className="text-sm text-muted-foreground mt-1">{feature.description}</p>
              <Badge variant="outline" className="mt-2 text-xs capitalize">
                {feature.category}
              </Badge>
            </div>
          </div>
        </div>
        
        <div className="mt-4 pt-4 border-t">
          <p className="text-xs text-muted-foreground mb-2">Available on:</p>
          <div className="flex flex-wrap gap-2">
            {availablePlans.map(plan => {
              const color = PLAN_COLORS[plan] || PLAN_COLORS.starter;
              const PlanIcon = color.icon;
              return (
                <Badge key={plan} className={cn("text-xs", color.bg, color.text, color.border)}>
                  <PlanIcon className="w-3 h-3 mr-1" />
                  {plan.replace('_', ' ')}
                </Badge>
              );
            })}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function PlanCard({ plan }: { plan: PlanInfo }) {
  const color = PLAN_COLORS[plan.plan] || PLAN_COLORS.starter;
  const PlanIcon = color.icon;
  
  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardContent className="pt-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className={cn("p-2 rounded-lg", color.bg)}>
              <PlanIcon className={cn("w-5 h-5", color.text)} />
            </div>
            <div>
              <h3 className="font-semibold capitalize">{plan.plan.replace('_', ' ')}</h3>
              <p className="text-sm text-muted-foreground">
                {plan.is_agent ? 'Agent Plan' : plan.is_enterprise ? 'Enterprise' : plan.is_pro ? 'Pro' : 'Starter'}
              </p>
            </div>
          </div>
          <Badge variant="outline" className="text-lg font-semibold">
            {plan.feature_count} features
          </Badge>
        </div>
        
        <div className="space-y-2">
          <div className="flex gap-2 flex-wrap">
            {plan.features.slice(0, 5).map(feature => (
              <Badge key={feature} variant="secondary" className="text-xs">
                {feature.replace(/_/g, ' ')}
              </Badge>
            ))}
            {plan.features.length > 5 && (
              <Badge variant="secondary" className="text-xs">
                +{plan.features.length - 5} more
              </Badge>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function AdminFeaturesPage() {
  const [activeTab, setActiveTab] = useState("overview");
  const [searchQuery, setSearchQuery] = useState("");
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [selectedPlan, setSelectedPlan] = useState<string>("all");

  // Fetch all features
  const { data: featuresData, isLoading: featuresLoading } = useQuery({
    queryKey: ['admin-features'],
    queryFn: () => featuresApi.listFeatures(),
  });

  // Fetch all plans info
  const { data: plansData, isLoading: plansLoading } = useQuery({
    queryKey: ['admin-plans'],
    queryFn: () => featuresApi.getAllPlansInfo(),
  });

  // Get features by plan
  const { data: selectedPlanData } = useQuery({
    queryKey: ['admin-plan-features', selectedPlan],
    queryFn: () => featuresApi.getPlanInfo(selectedPlan),
    enabled: selectedPlan !== "all",
  });

  const features = featuresData?.features || [];
  const plans = plansData?.plans || [];

  // Filter features by search and category
  const filteredFeatures = features.filter(feature => {
    const matchesSearch = feature.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      feature.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = categoryFilter === "all" || feature.category === categoryFilter;
    return matchesSearch && matchesCategory;
  });

  // Group features by category
  const featuresByCategory = filteredFeatures.reduce((acc, feature) => {
    if (!acc[feature.category]) {
      acc[feature.category] = [];
    }
    acc[feature.category].push(feature);
    return acc;
  }, {} as Record<string, Feature[]>);

  // Get available plans for each feature
  const getAvailablePlans = (featureKey: string): string[] => {
    return plans
      .filter(plan => plan.features.includes(featureKey))
      .map(plan => plan.plan);
  };

  const isLoading = featuresLoading || plansLoading;

  if (isLoading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center gap-2 mb-6">
          <Skeleton className="h-8 w-8" />
          <Skeleton className="h-10 w-64" />
        </div>
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-primary/10">
            <Shield className="w-6 h-6 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">Feature Management</h1>
            <p className="text-muted-foreground">Manage tier-specific features and permissions</p>
          </div>
        </div>
        <Button variant="outline" size="sm">
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-blue-500/10">
                <Layers className="w-6 h-6 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Features</p>
                <p className="text-2xl font-bold">{features.length}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-purple-500/10">
                <Crown className="w-6 h-6 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Enterprise Only</p>
                <p className="text-2xl font-bold">
                  {plans.find(p => p.plan === 'enterprise')?.feature_count || 0}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-green-500/10">
                <Star className="w-6 h-6 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Starter Features</p>
                <p className="text-2xl font-bold">
                  {plans.find(p => p.plan === 'starter')?.feature_count || 0}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-amber-500/10">
                <Users className="w-6 h-6 text-amber-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Agent Plans</p>
                <p className="text-2xl font-bold">
                  {plans.filter(p => p.is_agent).length}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="mb-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="features">All Features</TabsTrigger>
          <TabsTrigger value="plans">Plans</TabsTrigger>
          <TabsTrigger value="compare">Compare</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {plans.map(plan => (
              <PlanCard key={plan.plan} plan={plan} />
            ))}
          </div>
        </TabsContent>

        {/* All Features Tab */}
        <TabsContent value="features">
          {/* Filters */}
          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Search features..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className="w-[180px]">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue placeholder="Category" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Categories</SelectItem>
                <SelectItem value="core">Core</SelectItem>
                <SelectItem value="security">Security</SelectItem>
                <SelectItem value="analytics">Analytics</SelectItem>
                <SelectItem value="support">Support</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Features by Category */}
          <div className="space-y-8">
            {Object.entries(featuresByCategory).map(([category, categoryFeatures]) => {
              const CategoryIcon = CATEGORY_ICONS[category] || Zap;
              return (
                <div key={category}>
                  <div className="flex items-center gap-2 mb-4">
                    <CategoryIcon className="w-5 h-5" />
                    <h2 className="text-lg font-semibold capitalize">{category}</h2>
                    <Badge variant="secondary">{categoryFeatures.length}</Badge>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {categoryFeatures.map(feature => (
                      <FeatureCard
                        key={feature.key}
                        feature={feature}
                        availablePlans={getAvailablePlans(feature.key)}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </TabsContent>

        {/* Plans Tab */}
        <TabsContent value="plans">
          <div className="mb-4">
            <Select value={selectedPlan} onValueChange={setSelectedPlan}>
              <SelectTrigger className="w-[250px]">
                <SelectValue placeholder="Select a plan" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Select a plan to view details</SelectItem>
                {plans.map(plan => (
                  <SelectItem key={plan.plan} value={plan.plan}>
                    {plan.plan.replace('_', ' ')} ({plan.feature_count} features)
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {selectedPlan === "all" ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {plans.map(plan => (
                <PlanCard key={plan.plan} plan={plan} />
              ))}
            </div>
          ) : selectedPlanData ? (
            <Card>
              <CardHeader>
                <CardTitle className="capitalize flex items-center gap-2">
                  {(() => {
                    const color = PLAN_COLORS[selectedPlan] || PLAN_COLORS.starter;
                    const PlanIcon = color.icon;
                    return <PlanIcon className={cn("w-5 h-5", color.text)} />;
                  })()}
                  {selectedPlan.replace('_', ' ')} Plan Details
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <h4 className="font-medium mb-2">Feature Count</h4>
                    <p className="text-3xl font-bold">{selectedPlanData.feature_count}</p>
                  </div>
                  <div>
                    <h4 className="font-medium mb-2">All Features</h4>
                    <div className="flex flex-wrap gap-2">
                      {selectedPlanData.features.map(feature => (
                        <Badge key={feature} variant="outline" className="text-sm">
                          {feature.replace(/_/g, ' ')}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ) : (
            <LoadingSpinner />
          )}
        </TabsContent>

        {/* Compare Tab */}
        <TabsContent value="compare">
          <Card>
            <CardHeader>
              <CardTitle>Plan Comparison</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr>
                      <th className="text-left p-3 border-b">Feature</th>
                      {plans.filter(p => !p.is_agent).map(plan => (
                        <th key={plan.plan} className="text-center p-3 border-b capitalize">
                          {plan.plan}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {features.map(feature => {
                      const starterHas = plans.find(p => p.plan === 'starter')?.features.includes(feature.key);
                      const proHas = plans.find(p => p.plan === 'pro')?.features.includes(feature.key);
                      const enterpriseHas = plans.find(p => p.plan === 'enterprise')?.features.includes(feature.key);
                      
                      return (
                        <tr key={feature.key} className="border-b">
                          <td className="p-3">
                            <div className="font-medium">{feature.name}</div>
                            <div className="text-xs text-muted-foreground">{feature.description}</div>
                          </td>
                          <td className="text-center p-3">
                            {starterHas ? (
                              <CheckCircle className="w-5 h-5 text-green-500 mx-auto" />
                            ) : (
                              <XCircle className="w-5 h-5 text-red-300 mx-auto" />
                            )}
                          </td>
                          <td className="text-center p-3">
                            {proHas ? (
                              <CheckCircle className="w-5 h-5 text-green-500 mx-auto" />
                            ) : (
                              <XCircle className="w-5 h-5 text-red-300 mx-auto" />
                            )}
                          </td>
                          <td className="text-center p-3">
                            {enterpriseHas ? (
                              <CheckCircle className="w-5 h-5 text-green-500 mx-auto" />
                            ) : (
                              <XCircle className="w-5 h-5 text-red-300 mx-auto" />
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default AdminFeaturesPage;
