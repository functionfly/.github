import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import {
  Plus,
  Search,
  Filter,
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle2,
  MoreHorizontal,
  Edit3,
  Trash2,
  RefreshCw,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useIncidents, useCreateIncident, useUpdateIncident, incidentKeys } from '@/hooks/useStatus';
import { useSeverityColors } from '@/hooks/useStatus';
import type { Incident, IncidentSeverity, IncidentStatus, CreateIncidentRequest } from '@/api/status';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { toast } from 'sonner';

// Status badge component
function StatusBadge({ status }: { status: IncidentStatus }) {
  const config = {
    investigating: { color: 'bg-red-500/10 text-red-400 border-red-500/30', label: 'Investigating' },
    identified: { color: 'bg-amber-500/10 text-amber-400 border-amber-500/30', label: 'Identified' },
    monitoring: { color: 'bg-blue-500/10 text-blue-400 border-blue-500/30', label: 'Monitoring' },
    resolved: { color: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30', label: 'Resolved' },
  };

  const { color, label } = config[status];

  return (
    <Badge variant="outline" className={cn(color)}>
      {label}
    </Badge>
  );
}

// Severity badge component
function SeverityBadge({ severity }: { severity: IncidentSeverity }) {
  const colors = useSeverityColors();
  const config = colors[severity];

  return (
    <Badge variant="outline" className={cn(config.border, config.text, 'bg-opacity-10')}>
      {config.label}
    </Badge>
  );
}

// Create incident dialog
function CreateIncidentDialog({ onCreated }: { onCreated: () => void }) {
  const [open, setOpen] = useState(false);
  const createIncident = useCreateIncident();

  const [formData, setFormData] = useState<CreateIncidentRequest>({
    title: '',
    description: '',
    severity: 'medium',
    status: 'investigating',
    affected_components: [],
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.title || !formData.description) {
      toast.error('Please fill in all required fields');
      return;
    }

    try {
      await createIncident.mutateAsync(formData);
      toast.success('Incident created successfully');
      setOpen(false);
      onCreated();
      setFormData({
        title: '',
        description: '',
        severity: 'medium',
        status: 'investigating',
        affected_components: [],
      });
    } catch (error) {
      toast.error('Failed to create incident');
    }
  };

  const componentOptions = [
    'Orchestrator',
    'Health Monitor',
    'Database',
    'Cache',
    'Caddy',
    'Cloudflare',
    'Vercel',
    'Fly',
    'Deno Deploy',
  ];

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Create Incident
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Create New Incident</DialogTitle>
          <DialogDescription>
            Report a new service disruption or issue.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="title">Title *</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="e.g., Database connectivity issues"
              required
            />
          </div>
          <div>
            <Label htmlFor="description">Description *</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Describe the issue and its impact..."
              rows={4}
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label htmlFor="severity">Severity</Label>
              <Select
                value={formData.severity}
                onValueChange={(v) =>
                  setFormData({ ...formData, severity: v as IncidentSeverity })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="critical">Critical</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="status">Initial Status</Label>
              <Select
                value={formData.status}
                onValueChange={(v) =>
                  setFormData({ ...formData, status: v as IncidentStatus })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="investigating">Investigating</SelectItem>
                  <SelectItem value="identified">Identified</SelectItem>
                  <SelectItem value="monitoring">Monitoring</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div>
            <Label>Affected Components</Label>
            <div className="mt-2 flex flex-wrap gap-2">
              {componentOptions.map((component) => (
                <Badge
                  key={component}
                  variant={
                    formData.affected_components.includes(component)
                      ? 'default'
                      : 'outline'
                  }
                  className="cursor-pointer"
                  onClick={() => {
                    const updated = formData.affected_components.includes(component)
                      ? formData.affected_components.filter((c) => c !== component)
                      : [...formData.affected_components, component];
                    setFormData({ ...formData, affected_components: updated });
                  }}
                >
                  {component}
                </Badge>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createIncident.isPending}>
              {createIncident.isPending ? 'Creating...' : 'Create Incident'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Update incident dialog
function UpdateIncidentDialog({
  incident,
  onUpdated,
}: {
  incident: Incident;
  onUpdated: () => void;
}) {
  const [open, setOpen] = useState(false);
  const updateIncident = useUpdateIncident();

  const [formData, setFormData] = useState({
    status: incident.status,
    message: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await updateIncident.mutateAsync({
        id: incident.id,
        data: {
          status: formData.status,
          message: formData.message || undefined,
        },
      });
      toast.success('Incident updated successfully');
      setOpen(false);
      onUpdated();
    } catch (error) {
      toast.error('Failed to update incident');
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
          <Edit3 className="mr-2 h-4 w-4" />
          Update Status
        </DropdownMenuItem>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Update Incident</DialogTitle>
          <DialogDescription>
            Update the status of "{incident.title}"
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="status">New Status</Label>
            <Select
              value={formData.status}
              onValueChange={(v) => setFormData({ ...formData, status: v as IncidentStatus })}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="investigating">Investigating</SelectItem>
                <SelectItem value="identified">Identified</SelectItem>
                <SelectItem value="monitoring">Monitoring</SelectItem>
                <SelectItem value="resolved">Resolved</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label htmlFor="message">Update Message (Optional)</Label>
            <Textarea
              id="message"
              value={formData.message}
              onChange={(e) => setFormData({ ...formData, message: e.target.value })}
              placeholder="Add an update message..."
              rows={3}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateIncident.isPending}>
              {updateIncident.isPending ? 'Updating...' : 'Update'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Main page component
export default function AdminIncidentsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<IncidentStatus | 'all'>('all');
  const [severityFilter, setSeverityFilter] = useState<IncidentSeverity | 'all'>('all');

  const queryClient = useQueryClient();

  const {
    data: incidentsData,
    isLoading,
    refetch,
  } = useIncidents({
    status: statusFilter,
    severity: severityFilter,
    limit: 50,
  });

  // Filter by search query locally
  const filteredIncidents =
    incidentsData?.incidents.filter(
      (incident) =>
        !searchQuery ||
        incident.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        incident.description.toLowerCase().includes(searchQuery.toLowerCase())
    ) || [];

  const activeIncidents = filteredIncidents.filter((i) => i.status !== 'resolved');
  const resolvedIncidents = filteredIncidents.filter((i) => i.status === 'resolved');

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Incident Management</h1>
          <p className="mt-1 text-sm text-text-secondary">
            Manage and track service incidents
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isLoading}
          >
            <RefreshCw className={cn('mr-2 h-4 w-4', isLoading && 'animate-spin')} />
            Refresh
          </Button>
          <CreateIncidentDialog onCreated={() => refetch()} />
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted">
              Total Incidents
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-text-primary">
              {incidentsData?.total || 0}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted">
              Active Incidents
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-400">{activeIncidents.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted">
              Critical Issues
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-orange-400">
              {filteredIncidents.filter((i) => i.severity === 'critical').length}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-muted">
              Resolved Today
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-emerald-400">
              {
                resolvedIncidents.filter((i) => {
                  const today = new Date();
                  const resolved = i.resolved_at ? new Date(i.resolved_at) : null;
                  return (
                    resolved &&
                    resolved.toDateString() === today.toDateString()
                  );
                }).length
              }
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-center gap-3">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
              <Input
                placeholder="Search incidents..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as IncidentStatus | 'all')}
              className="h-10 rounded-md border border-border-subtle bg-bg-secondary px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500"
            >
              <option value="all">All Statuses</option>
              <option value="investigating">Investigating</option>
              <option value="identified">Identified</option>
              <option value="monitoring">Monitoring</option>
              <option value="resolved">Resolved</option>
            </select>
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value as IncidentSeverity | 'all')}
              className="h-10 rounded-md border border-border-subtle bg-bg-secondary px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500"
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
            </select>
          </div>
        </CardContent>
      </Card>

      {/* Incidents Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Incidents</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : filteredIncidents.length === 0 ? (
            <div className="py-12 text-center">
              <CheckCircle2 className="mx-auto h-12 w-12 text-emerald-400" />
              <h3 className="mt-4 font-medium text-text-primary">No incidents found</h3>
              <p className="mt-1 text-sm text-text-secondary">
                {searchQuery || statusFilter !== 'all' || severityFilter !== 'all'
                  ? 'Try adjusting your filters.'
                  : 'Great! No incidents to report.'}
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Severity</TableHead>
                    <TableHead>Affected</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredIncidents.map((incident) => (
                    <TableRow key={incident.id}>
                      <TableCell>
                        <div>
                          <p className="font-medium text-text-primary">{incident.title}</p>
                          <p className="text-xs text-text-muted line-clamp-1">
                            {incident.description}
                          </p>
                        </div>
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={incident.status} />
                      </TableCell>
                      <TableCell>
                        <SeverityBadge severity={incident.severity} />
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {incident.affected_components.slice(0, 2).map((comp) => (
                            <Badge key={comp} variant="secondary" className="text-xs">
                              {comp}
                            </Badge>
                          ))}
                          {incident.affected_components.length > 2 && (
                            <Badge variant="secondary" className="text-xs">
                              +{incident.affected_components.length - 2}
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-text-muted">
                        {new Date(incident.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <UpdateIncidentDialog
                              incident={incident}
                              onUpdated={() => refetch()}
                            />
                            <DropdownMenuItem className="text-red-400">
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
