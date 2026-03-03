import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, MoreVertical, ExternalLink, Eye, EyeOff, Edit, Trash2, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { FormField } from "@/components/ui/form-field";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { StatCard } from "@/components/common/StatCard";
import { Textarea } from "@/components/ui/textarea";
import { useAuthStore } from "@/stores/authStore";
import { useRedirectForm } from "@/hooks/useAdminForms";

// Types for redirect data
interface Redirect {
  _id: string;
  source: string;
  destination: string;
  statusCode: number;
  matchType: "exact" | "prefix" | "regex";
  enabled: boolean;
  notes?: string;
  createdBy: { name: string };
  createdAt: string;
  updatedAt?: string;
}

interface RedirectStats {
  title: string;
  value: number;
  change: { value: number; label: string };
  icon: React.ReactNode;
}

export function AdminRedirectsPage() {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const [redirects, setRedirects] = useState<Redirect[]>([]);
  const [stats, setStats] = useState<RedirectStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [editingRedirect, setEditingRedirect] = useState<Redirect | null>(null);

  // Ensure user is authenticated
  if (!user) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Redirect Management</h1>
            <p className="text-red-600">You must be logged in to access this page.</p>
          </div>
        </div>
      </div>
    );
  }

  // Fetch redirects (mock data since Sanity was removed)
  useEffect(() => {
    const fetchRedirects = async () => {
      try {
        setLoading(true);

        // Mock redirect data
        const mockData = [
          {
            _id: "1",
            source: "/old-blog-post",
            destination: "/blog/new-blog-post",
            statusCode: 301,
            matchType: "exact" as const,
            enabled: true,
            notes: "Old blog URL redirect",
            createdBy: { name: "Admin" },
            createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
            updatedAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString()
          },
          {
            _id: "2",
            source: "/api/v1/*",
            destination: "/api/v2/$1",
            statusCode: 302,
            matchType: "regex" as const,
            enabled: true,
            notes: "API version redirect",
            createdBy: { name: "Admin" },
            createdAt: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000).toISOString()
          },
          {
            _id: "3",
            source: "/products",
            destination: "/tools",
            statusCode: 301,
            matchType: "exact" as const,
            enabled: false,
            notes: "Temporarily disabled for maintenance",
            createdBy: { name: "Admin" },
            createdAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString()
          }
        ];

        const data = mockData;

        // Calculate stats
        const totalRedirects = data.length;
        const activeRedirects = data.filter((r: Redirect) => r.enabled).length;
        const permanentRedirects = data.filter((r: Redirect) => r.statusCode === 301).length;
        const temporaryRedirects = data.filter((r: Redirect) => r.statusCode === 302).length;

        const calculatedStats: RedirectStats[] = [
          {
            title: "Total Redirects",
            value: totalRedirects,
            change: { value: 0, label: "from last month" },
            icon: <ExternalLink className="w-5 h-5 text-brand-500" />,
          },
          {
            title: "Active Redirects",
            value: activeRedirects,
            change: { value: 0, label: "from last month" },
            icon: <Eye className="w-5 h-5 text-status-online" />,
          },
          {
            title: "301 Redirects",
            value: permanentRedirects,
            change: { value: 0, label: "from last month" },
            icon: <Badge className="w-5 h-5 text-status-degraded" />,
          },
          {
            title: "302 Redirects",
            value: temporaryRedirects,
            change: { value: 0, label: "from last month" },
            icon: <Badge className="w-5 h-5 text-status-offline" />,
          },
        ];

        setRedirects(data);
        setStats(calculatedStats);
        setError(null);
      } catch (err) {
        console.error('Failed to fetch redirects:', err);
        setError('Failed to load redirects. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    fetchRedirects();
  }, []);

  const filteredRedirects = redirects.filter(redirect =>
    redirect.source.toLowerCase().includes(searchQuery.toLowerCase()) ||
    redirect.destination.toLowerCase().includes(searchQuery.toLowerCase()) ||
    redirect.notes?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleToggleEnabled = async (redirectId: string) => {
    try {
      if (!user) {
        setError('You must be logged in to perform this action.');
        return;
      }

      const redirect = redirects.find(r => r._id === redirectId);
      if (!redirect) return;

      const now = new Date().toISOString();

      // Mock update operation (Sanity removed)
      await new Promise(resolve => setTimeout(resolve, 500)); // Simulate API delay

      // Update local state
      setRedirects(redirects.map(r =>
        r._id === redirectId
          ? { ...r, enabled: !r.enabled, updatedAt: now }
          : r
      ));
    } catch (err) {
      console.error('Failed to toggle redirect:', err);
      setError('Failed to update redirect. Please try again.');
    }
  };

  const handleDelete = async (redirectId: string) => {
    try {
      if (!user) {
        setError('You must be logged in to perform this action.');
        return;
      }

      // Mock delete operation (Sanity removed)
      await new Promise(resolve => setTimeout(resolve, 300)); // Simulate API delay

      // Update local state
      setRedirects(redirects.filter(redirect => redirect._id !== redirectId));
    } catch (err) {
      console.error('Failed to delete redirect:', err);
      setError('Failed to delete redirect. Please try again.');
    }
  };

  const RedirectForm = ({ redirect, onClose }: { redirect?: Redirect, onClose: () => void }) => {
    const user = useAuthStore((state) => state.user);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const {
      register,
      handleSubmit,
      watch,
      formState: { errors, isValid, isSubmitting: formSubmitting },
      setValue,
    } = useRedirectForm({
      source: redirect?.source || "",
      destination: redirect?.destination || "",
      statusCode: redirect?.statusCode || 301,
      matchType: redirect?.matchType || "exact",
      enabled: redirect?.enabled ?? true,
      notes: redirect?.notes || "",
    });

    const onSubmit = async (data: any) => {
      setIsSubmitting(true);

      try {
        if (!user) {
          setError('You must be logged in to perform this action.');
          return;
        }

        const now = new Date().toISOString();
        const currentUserId = user.id;

        if (redirect) {
          // Update existing
          const updatedRedirect = {
            ...data,
            updatedAt: now,
            updatedBy: { _type: 'reference', _ref: currentUserId }
          };

          // Mock update operation (Sanity removed)
          await new Promise(resolve => setTimeout(resolve, 500)); // Simulate API delay

          // Update local state
          setRedirects(redirects.map(r =>
            r._id === redirect._id ? { ...r, ...data, updatedAt: now } : r
          ));
        } else {
          // Create new
          const newRedirect = {
            _type: 'redirect',
            ...data,
            createdAt: now,
            updatedAt: now,
            createdBy: { _type: 'reference', _ref: currentUserId },
            updatedBy: { _type: 'reference', _ref: currentUserId }
          };

          // Mock create operation (Sanity removed)
          const result = { _id: Date.now().toString() }; // Mock ID
          await new Promise(resolve => setTimeout(resolve, 500)); // Simulate API delay

          // Add to local state
          const redirectWithRefs: Redirect = {
            _id: result._id,
            source: data.source,
            destination: data.destination,
            statusCode: data.statusCode,
            matchType: data.matchType,
            enabled: data.enabled,
            notes: data.notes,
            createdBy: { name: user.name },
            createdAt: now,
            updatedAt: now,
          };

          setRedirects([...redirects, redirectWithRefs]);
        }
        onClose();
      } catch (err) {
        console.error('Failed to save redirect:', err);
        setError('Failed to save redirect. Please try again.');
      } finally {
        setIsSubmitting(false);
      }
    };

    return (
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <FormField
          label="Source Path"
          name="source"
          type="text"
          placeholder="/old-path"
          error={errors.source}
          success={!errors.source && watch('source')}
          required
          helperText="Must start with '/' (e.g., /old-path)"
          {...register('source')}
        />

        <FormField
          label="Destination URL"
          name="destination"
          type="url"
          placeholder="https://example.com/new-path"
          error={errors.destination}
          success={!errors.destination && watch('destination')}
          required
          helperText="Full URL including protocol"
          {...register('destination')}
        />

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="statusCode">Status Code</Label>
            <Select
              value={watch('statusCode')?.toString()}
              onValueChange={(value) => setValue('statusCode', parseInt(value) as any)}
            >
              <SelectTrigger className={errors.statusCode ? 'border-red-500' : ''}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="301">301 - Permanent</SelectItem>
                <SelectItem value="302">302 - Temporary</SelectItem>
                <SelectItem value="307">307 - Temporary (preserve method)</SelectItem>
                <SelectItem value="308">308 - Permanent (preserve method)</SelectItem>
              </SelectContent>
            </Select>
            {errors.statusCode && (
              <div className="text-xs text-red-600 dark:text-red-400">
                {(errors.statusCode as any)?.message}
              </div>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="matchType">Match Type</Label>
            <Select
              value={watch('matchType')}
              onValueChange={(value) => setValue('matchType', value as 'exact' | 'prefix' | 'regex')}
            >
              <SelectTrigger className={errors.matchType ? 'border-red-500' : ''}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="exact">Exact Match</SelectItem>
                <SelectItem value="prefix">Prefix Match</SelectItem>
                <SelectItem value="regex">Regex Match</SelectItem>
              </SelectContent>
            </Select>
            {errors.matchType && (
              <div className="text-xs text-red-600 dark:text-red-400">
                {(errors.matchType as any)?.message}
              </div>
            )}
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="notes">Notes</Label>
          <Textarea
            id="notes"
            placeholder="Internal notes about this redirect"
            rows={3}
            className={errors.notes ? 'border-red-500' : ''}
            {...register('notes')}
          />
          {errors.notes && (
            <div className="text-xs text-red-600 dark:text-red-400">
              {(errors.notes as any)?.message}
            </div>
          )}
        </div>

        <div className="flex items-center space-x-2">
          <Checkbox
            id="enabled"
            checked={watch('enabled')}
            onCheckedChange={(checked) => setValue('enabled', checked as boolean)}
          />
          <Label htmlFor="enabled" className="cursor-pointer">Enabled</Label>
        </div>

        <div className="flex justify-end space-x-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={isSubmitting || formSubmitting}>
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting || formSubmitting || !isValid}>
            {isSubmitting || formSubmitting ? (
              <LoadingSpinner text="Saving..." size="sm" />
            ) : (
              (redirect ? 'Update' : 'Create') + ' Redirect'
            )}
          </Button>
        </div>
      </form>
    );
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Redirect Management</h1>
            <p className="text-muted-foreground">Loading redirects...</p>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <div className="animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-8 bg-gray-200 rounded w-1/2"></div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Redirect Management</h1>
            <p className="text-red-600">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
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
          <h1 className="text-3xl font-bold text-text-primary">Redirect Management</h1>
          <p className="text-text-secondary">
            Manage URL redirects for your site
          </p>
        </div>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              Add Redirect
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Create New Redirect</DialogTitle>
            </DialogHeader>
            <RedirectForm onClose={() => setIsCreateDialogOpen(false)} />
          </DialogContent>
        </Dialog>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      {/* Search and Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Redirects</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-4 mb-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
              <Input
                placeholder="Search redirects..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>

          <div className="space-y-4">
            {filteredRedirects.map((redirect) => (
              <div key={redirect._id} className="flex items-center justify-between p-4 border rounded-lg">
                <div className="flex-1">
                  <div className="flex items-center space-x-2">
                    <code className="text-sm bg-muted px-2 py-1 rounded">
                      {redirect.source}
                    </code>
                    <span className="text-muted-foreground">→</span>
                    <code className="text-sm bg-muted px-2 py-1 rounded">
                      {redirect.destination}
                    </code>
                    <Badge variant={redirect.statusCode === 301 ? "default" : "secondary"}>
                      {redirect.statusCode}
                    </Badge>
                    <Badge variant="outline">
                      {redirect.matchType}
                    </Badge>
                    {redirect.enabled ? (
                      <Eye className="w-4 h-4 text-green-600" />
                    ) : (
                      <EyeOff className="w-4 h-4 text-text-muted" />
                    )}
                  </div>
                  {redirect.notes && (
                    <p className="text-sm text-muted-foreground mt-1">
                      {redirect.notes}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground mt-1">
                    Created by {redirect.createdBy.name} on {new Date(redirect.createdAt).toLocaleDateString()}
                  </p>
                </div>

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="sm" aria-label="Redirect options">
                      <MoreVertical className="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => setEditingRedirect(redirect)}>
                      <Edit className="w-4 h-4 mr-2" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => handleToggleEnabled(redirect._id)}>
                      {redirect.enabled ? (
                        <>
                          <EyeOff className="w-4 h-4 mr-2" />
                          Disable
                        </>
                      ) : (
                        <>
                          <Eye className="w-4 h-4 mr-2" />
                          Enable
                        </>
                      )}
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => handleDelete(redirect._id)}
                      className="text-red-600"
                    >
                      <Trash2 className="w-4 h-4 mr-2" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={!!editingRedirect} onOpenChange={() => setEditingRedirect(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Redirect</DialogTitle>
          </DialogHeader>
          {editingRedirect && (
            <RedirectForm
              redirect={editingRedirect}
              onClose={() => setEditingRedirect(null)}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
