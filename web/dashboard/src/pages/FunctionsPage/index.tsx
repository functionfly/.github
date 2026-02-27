import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, MoreVertical, Rocket, Edit, Trash2, Eye } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { StatusBadge } from "@/components/common/StatusBadge";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const mockFunctions = [
  {
    id: 1,
    name: "api-handler",
    providers: ["workers", "vercel"],
    status: "online" as const,
    lastDeployed: "2 hours ago",
    requests: "12.5K",
  },
  {
    id: 2,
    name: "auth-service",
    providers: ["workers", "fly"],
    status: "online" as const,
    lastDeployed: "1 day ago",
    requests: "8.2K",
  },
  {
    id: 3,
    name: "webhook-handler",
    providers: ["vercel"],
    status: "degraded" as const,
    lastDeployed: "3 days ago",
    requests: "3.1K",
  },
  {
    id: 4,
    name: "image-processor",
    providers: ["fly"],
    status: "offline" as const,
    lastDeployed: "1 week ago",
    requests: "0",
  },
];

export function FunctionsPage() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");

  const filteredFunctions = mockFunctions.filter((fn) =>
    fn.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Functions</h1>
          <p className="text-text-secondary">Manage and deploy your edge functions</p>
        </div>
        <Button className="gap-2" onClick={() => navigate("/functions/new")}>
          <Plus className="w-4 h-4" />
          Deploy New
        </Button>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search functions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <Button variant="outline">Filter</Button>
      </div>

      {/* Functions List */}
      <div className="space-y-3">
        {filteredFunctions.map((fn) => (
          <Card key={fn.id} className="hover:border-[#6366f1]/30 transition-colors">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-lg bg-[#6366f1]/10 flex items-center justify-center">
                    <Rocket className="w-5 h-5 text-[#6366f1]" />
                  </div>
                  <div>
                    <h3 className="font-medium text-white">{fn.name}</h3>
                    <p className="text-sm text-text-muted">
                      Last deployed {fn.lastDeployed} • {fn.requests} requests
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-6">
                  {/* Providers */}
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-text-muted">Providers:</span>
                    <div className="flex -space-x-1">
                      {fn.providers.map((provider) => (
                        <div
                          key={provider}
                          className="w-6 h-6 rounded-full bg-bg-tertiary border border-white/8 flex items-center justify-center"
                        >
                          <ProviderIcon provider={provider} size="sm" />
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Status */}
                  <StatusBadge status={fn.status} />

                  {/* Actions */}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon" className="text-text-secondary">
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="bg-bg-tertiary border-white/8">
                      <DropdownMenuItem className="gap-2" onClick={() => navigate(`/functions/${fn.id}`)}>
                        <Eye className="w-4 h-4" />
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem className="gap-2" onClick={() => navigate(`/functions/${fn.id}/edit`)}>
                        <Edit className="w-4 h-4" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem className="gap-2">
                        <Rocket className="w-4 h-4" />
                        Redeploy
                      </DropdownMenuItem>
                      <DropdownMenuItem className="gap-2 text-red-400">
                        <Trash2 className="w-4 h-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredFunctions.length === 0 && (
        <Card className="p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Rocket className="w-8 h-8 text-text-muted" />
          </div>
          <h3 className="text-lg font-medium text-white mb-2">No functions yet</h3>
          <p className="text-text-secondary mb-6">Deploy your first function to get started</p>
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Deploy Function
          </Button>
        </Card>
      )}
    </div>
  );
}
