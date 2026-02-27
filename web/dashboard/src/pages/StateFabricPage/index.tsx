import { useState } from "react";
import { Plus, Database, Zap, Network, Activity, Settings, MoreVertical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatCard } from "@/components/common/StatCard";
import { StatusBadge } from "@/components/common/StatusBadge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const mockStateFabrics = [
  {
    id: 1,
    name: "User Sessions",
    description: "Real-time user session management and synchronization",
    status: "online" as const,
    type: "session",
    stores: 3,
    pipelines: 2,
    throughput: "2.5K ops/sec",
    latency: "12ms",
    lastUpdated: "2 minutes ago",
  },
  {
    id: 2,
    name: "Product Catalog",
    description: "Distributed product data and inventory management",
    status: "online" as const,
    type: "catalog",
    stores: 5,
    pipelines: 4,
    throughput: "5.1K ops/sec",
    latency: "8ms",
    lastUpdated: "1 hour ago",
  },
  {
    id: 3,
    name: "Analytics Cache",
    description: "High-performance analytics data caching and aggregation",
    status: "degraded" as const,
    type: "cache",
    stores: 2,
    pipelines: 1,
    throughput: "1.8K ops/sec",
    latency: "45ms",
    lastUpdated: "5 minutes ago",
  },
  {
    id: 4,
    name: "Order Processing",
    description: "Event-driven order processing and workflow orchestration",
    status: "offline" as const,
    type: "workflow",
    stores: 4,
    pipelines: 6,
    throughput: "0 ops/sec",
    latency: "-",
    lastUpdated: "2 hours ago",
  },
];

const stats = [
  {
    title: "Total State Fabrics",
    value: "12",
    change: { value: 2, label: "from last month" },
    icon: <Database className="w-5 h-5 text-blue-500" />,
    trend: "up" as const,
  },
  {
    title: "Active Pipelines",
    value: "28",
    change: { value: 5, label: "from last week" },
    icon: <Network className="w-5 h-5 text-green-500" />,
    trend: "up" as const,
  },
  {
    title: "Total Throughput",
    value: "45.2K",
    change: { value: 12, label: "from yesterday" },
    icon: <Zap className="w-5 h-5 text-yellow-500" />,
    trend: "up" as const,
  },
  {
    title: "Avg Latency",
    value: "15ms",
    change: { value: -3, label: "from last week" },
    icon: <Activity className="w-5 h-5 text-purple-500" />,
    trend: "up" as const,
  },
];

const getTypeIcon = (type: string) => {
  switch (type) {
    case "session":
      return "👤";
    case "catalog":
      return "📦";
    case "cache":
      return "⚡";
    case "workflow":
      return "🔄";
    default:
      return "🧵";
  }
};

const getTypeColor = (type: string) => {
  switch (type) {
    case "session":
      return "bg-blue-500/10 border-blue-500/20";
    case "catalog":
      return "bg-green-500/10 border-green-500/20";
    case "cache":
      return "bg-yellow-500/10 border-yellow-500/20";
    case "workflow":
      return "bg-purple-500/10 border-purple-500/20";
    default:
      return "bg-gray-500/10 border-gray-500/20";
  }
};

export function StateFabricPage() {
  const [searchQuery, setSearchQuery] = useState("");

  const filteredFabrics = mockStateFabrics.filter((fabric) =>
    fabric.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    fabric.description.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">State Fabric</h1>
          <p className="text-text-secondary">Manage state and data orchestration across your applications</p>
        </div>
        <Button className="gap-2">
          <Plus className="w-4 h-4" />
          Create State Fabric
        </Button>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <StatCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <input
            type="text"
            placeholder="Search state fabrics..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500"
          />
        </div>
        <Button variant="outline">Filter</Button>
        <Button variant="outline">Sort</Button>
      </div>

      {/* State Fabrics Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {filteredFabrics.map((fabric) => (
          <Card key={fabric.id} className="hover:border-brand-500/30 transition-colors">
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-lg ${getTypeColor(fabric.type)}`}>
                    {getTypeIcon(fabric.type)}
                  </div>
                  <div>
                    <CardTitle className="text-lg text-text-primary">{fabric.name}</CardTitle>
                    <p className="text-sm text-text-secondary">{fabric.description}</p>
                  </div>
                </div>
                <StatusBadge status={fabric.status} />
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Metrics */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <p className="text-xs text-text-muted uppercase tracking-wide">Throughput</p>
                  <p className="text-lg font-semibold text-text-primary">{fabric.throughput}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-xs text-text-muted uppercase tracking-wide">Latency</p>
                  <p className="text-lg font-semibold text-text-primary">{fabric.latency}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-xs text-text-muted uppercase tracking-wide">Stores</p>
                  <p className="text-lg font-semibold text-text-primary">{fabric.stores}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-xs text-text-muted uppercase tracking-wide">Pipelines</p>
                  <p className="text-lg font-semibold text-text-primary">{fabric.pipelines}</p>
                </div>
              </div>

              {/* Footer */}
              <div className="flex items-center justify-between pt-4 border-t border-border-subtle">
                <p className="text-xs text-text-muted">
                  Updated {fabric.lastUpdated}
                </p>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="text-text-secondary">
                      <MoreVertical className="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="bg-bg-tertiary border-white/8">
                    <DropdownMenuItem className="gap-2">
                      <Settings className="w-4 h-4" />
                      Configure
                    </DropdownMenuItem>
                    <DropdownMenuItem className="gap-2">
                      <Activity className="w-4 h-4" />
                      View Metrics
                    </DropdownMenuItem>
                    <DropdownMenuItem className="gap-2">
                      <Database className="w-4 h-4" />
                      Manage Stores
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredFabrics.length === 0 && (
        <Card className="p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Database className="w-8 h-8 text-text-muted" />
          </div>
          <h3 className="text-lg font-medium text-text-primary mb-2">No state fabrics yet</h3>
          <p className="text-text-secondary mb-6">
            {searchQuery ? "No fabrics match your search." : "Create your first state fabric to get started with data orchestration."}
          </p>
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create State Fabric
          </Button>
        </Card>
      )}
    </div>
  );
}