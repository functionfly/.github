import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  Copy,
  Check,
  RotateCcw,
  Trash2,
  Edit,
  Key,
  Clock,
  Shield,
  Globe,
  Loader2,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  APIKey,
  API_KEY_TYPE_LABELS,
  PERMISSION_LABELS,
  RESOURCE_TYPE_LABELS,
  RotationReason,
} from "@/types/api-key";
import { apiKeysService } from "@/services/api-keys";

interface APIKeyDetailsProps {
  onRotate?: (key: APIKey) => void;
  onDelete?: (key: APIKey) => void;
}

export function APIKeyDetails({ onRotate, onDelete }: APIKeyDetailsProps) {
  const { keyId } = useParams<{ keyId: string }>();
  const queryClient = useQueryClient();
  const [copied, setCopied] = useState(false);

  const { data: apiKey, isLoading, error, refetch } = useQuery({
    queryKey: ["api-key", keyId],
    queryFn: () => apiKeysService.getKey(keyId!),
    enabled: !!keyId,
  });

  const toggleActiveMutation = useMutation({
    mutationFn: (isActive: boolean) =>
      apiKeysService.updateKey(keyId!, { is_active: isActive }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-key", keyId] });
    },
  });

  const handleCopyKey = async () => {
    // Note: We can't copy the actual key here as it's not available in the API response
    // This would be used when showing the key after creation
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getKeyTypeBadgeVariant = (type: string) => {
    switch (type) {
      case "platform":
        return "default";
      case "function":
        return "secondary";
      default:
        return "outline";
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !apiKey) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <AlertCircle className="w-8 h-8 text-red-500 mb-2" />
        <p className="text-red-500 mb-4">
          {error instanceof Error ? error.message : "Failed to load API key"}
        </p>
        <Button variant="outline" onClick={() => refetch()}>
          Retry
        </Button>
      </div>
    );
  }

  const key = apiKey as APIKey;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to="/dashboard/api-keys">
          <Button variant="ghost" size="icon" aria-label="Back to API keys">
            <ArrowLeft className="w-5 h-5" />
          </Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
              <Key className="w-6 h-6 text-primary" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">{key.name}</h1>
              {key.description && (
                <p className="text-muted-foreground">{key.description}</p>
              )}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={key.is_active ? "default" : "secondary"}>
            {key.is_active ? "Active" : "Inactive"}
          </Badge>
          <Badge variant={getKeyTypeBadgeVariant(key.key_type)}>
            {API_KEY_TYPE_LABELS[key.key_type]}
          </Badge>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="flex gap-3">
        {onRotate && (
          <Button variant="outline" onClick={() => onRotate(key)}>
            <RotateCcw className="w-4 h-4 mr-2" />
            Rotate Key
          </Button>
        )}
        {onDelete && (
          <Button
            variant="outline"
            className="text-red-600 hover:text-red-600"
            onClick={() => onDelete(key)}
          >
            <Trash2 className="w-4 h-4 mr-2" />
            Delete Key
          </Button>
        )}
      </div>

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Created
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {formatDate(key.created_at)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Last Rotated
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {formatDate(key.last_rotated_at)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Last Used
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {formatDate(key.last_used_at)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Expires
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {key.expires_at ? formatDate(key.expires_at) : "Never"}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5" />
            Settings
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <Label>Active</Label>
              <p className="text-sm text-muted-foreground">
                Inactive keys cannot be used for authentication
              </p>
            </div>
            <Switch
              checked={key.is_active}
              onCheckedChange={(checked) => toggleActiveMutation.mutate(checked)}
              disabled={toggleActiveMutation.isPending}
            />
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="space-y-1">
              <Label className="text-muted-foreground">Rotation Frequency</Label>
              <p className="font-medium">{key.rotation_frequency_days} days</p>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Rate Limit (RPM)</Label>
              <p className="font-medium">
                {key.rate_limit_rpm?.toLocaleString() || "1,000"}
              </p>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Rate Limit (RPH)</Label>
              <p className="font-medium">
                {key.rate_limit_rph?.toLocaleString() || "60,000"}
              </p>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Rate Limit (RPD)</Label>
              <p className="font-medium">
                {key.rate_limit_rpd?.toLocaleString() || "1,000,000"}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Permissions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5" />
            Permissions
          </CardTitle>
        </CardHeader>
        <CardContent>
          {key.permissions && key.permissions.length > 0 ? (
            <div className="space-y-2">
              {key.permissions.map((permission) => (
                <div
                  key={permission.id}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <Badge variant="outline">
                      {RESOURCE_TYPE_LABELS[permission.resource_type]}
                    </Badge>
                    <span className="text-sm">
                      {PERMISSION_LABELS[permission.permission]}
                    </span>
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {formatDate(permission.created_at)}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">
              No permissions configured
            </p>
          )}
        </CardContent>
      </Card>

      {/* Environments */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="w-5 h-5" />
            Environments
          </CardTitle>
        </CardHeader>
        <CardContent>
          {key.environments && key.environments.length > 0 ? (
            <div className="space-y-2">
              {key.environments.map((env) => (
                <div
                  key={env.id}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <span className="font-medium">{env.environment_name}</span>
                  <span className="text-xs text-muted-foreground">
                    Added {formatDate(env.created_at)}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">
              No environments linked
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
