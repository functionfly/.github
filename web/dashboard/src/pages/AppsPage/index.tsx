import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Plus, Building2, Loader2, ExternalLink, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { appsApi } from "@/api/apps";
import type { App } from "@/types";
import { ROUTES } from "@/lib/constants";
import { toast } from "sonner";

function CreateAppModal({
  onSuccess,
}: {
  onSuccess: (app: App) => void;
}) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Auto-generate slug from name
  const handleNameChange = (value: string) => {
    setName(value);
    // Generate slug: lowercase, replace spaces with hyphens, remove non-alphanumeric
    const generatedSlug = value
      .toLowerCase()
      .replace(/\s+/g, "-")
      .replace(/[^a-z0-9-]/g, "");
    setSlug(generatedSlug);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError("App name is required");
      return;
    }

    if (name.length < 3 || name.length > 50) {
      setError("App name must be between 3 and 50 characters");
      return;
    }

    if (!slug.trim()) {
      setError("App slug is required");
      return;
    }

    if (!/^[a-z0-9-]+$/.test(slug)) {
      setError("Slug can only contain lowercase letters, numbers, and hyphens");
      return;
    }

    setIsSubmitting(true);

    try {
      const app = await appsApi.create({ name, slug });
      toast.success(`App "${name}" created successfully`);
      setOpen(false);
      setName("");
      setSlug("");
      onSuccess(app);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create app");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      setName("");
      setSlug("");
      setError(null);
    }
    setOpen(isOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="w-4 h-4 mr-2" />
          Create App
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New App</DialogTitle>
          <DialogDescription>
            Create a new app to organize your functions and manage deployments.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="flex items-center gap-2 p-3 text-sm text-red-500 bg-red-50 rounded-lg">
              <AlertCircle className="w-4 h-4" />
              {error}
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="name">App Name</Label>
            <Input
              id="name"
              placeholder="My Awesome App"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              disabled={isSubmitting}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="slug">App Slug</Label>
            <Input
              id="slug"
              placeholder="my-awesome-app"
              value={slug}
              onChange={(e) => setSlug(e.target.value.toLowerCase())}
              disabled={isSubmitting}
            />
            <p className="text-xs text-muted-foreground">
              Used in URLs: /apps/[slug]
            </p>
          </div>
          <div className="flex justify-end gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
              Create App
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function AppCard({ app }: { app: App }) {
  return (
    <Link to={`${ROUTES.APPS}/${app.id}`}>
      <Card className="hover:border-brand-500 transition-colors cursor-pointer">
        <CardContent className="p-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-brand-100 dark:bg-brand-900 flex items-center justify-center">
              <Building2 className="w-5 h-5 text-brand-600" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="font-medium truncate">{app.name}</h3>
              <p className="text-sm text-muted-foreground truncate">
                {app.slug}
              </p>
            </div>
            <ExternalLink className="w-4 h-4 text-muted-foreground" />
          </div>
          <div className="mt-3 text-xs text-muted-foreground">
            Created {new Date(app.createdAt).toLocaleDateString()}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

function EmptyState({ onCreateClick }: { onCreateClick: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="w-16 h-16 rounded-full bg-muted flex items-center justify-center mb-4">
        <Building2 className="w-8 h-8 text-muted-foreground" />
      </div>
      <h3 className="text-lg font-semibold mb-2">No apps yet</h3>
      <p className="text-muted-foreground mb-6 max-w-md">
        Create your first app to start organizing your functions and managing
        multi-cloud deployments.
      </p>
      <Button onClick={onCreateClick}>
        <Plus className="w-4 h-4 mr-2" />
        Create Your First App
      </Button>
    </div>
  );
}

export function AppsPage() {
  const [createModalOpen, setCreateModalOpen] = useState(false);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["apps"],
    queryFn: async () => {
      const response = await appsApi.list();
      return response.apps;
    },
  });

  const handleCreateSuccess = () => {
    refetch();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Apps</h1>
          <p className="text-muted-foreground">
            Manage your applications and their deployments
          </p>
        </div>
        <CreateAppModal onSuccess={handleCreateSuccess} />
      </div>

      {/* Error State */}
      {error && (
        <div className="flex items-center gap-2 p-4 text-red-500 bg-red-50 rounded-lg">
          <AlertCircle className="w-5 h-5" />
          <span>Failed to load apps. Please try again.</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
        </div>
      )}

      {/* Apps List */}
      {!isLoading && !error && (
        <>
          {data && data.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {data.map((app) => (
                <AppCard key={app.id} app={app} />
              ))}
            </div>
          ) : (
            <EmptyState onCreateClick={() => setCreateModalOpen(true)} />
          )}
        </>
      )}
    </div>
  );
}
