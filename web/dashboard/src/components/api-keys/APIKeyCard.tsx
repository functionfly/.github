import { useState } from "react";
import { Link } from "react-router-dom";
import { toast } from "sonner";
import {
  Key,
  Copy,
  Check,
  MoreVertical,
  RotateCcw,
  Trash2,
  Edit,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import {
  APIKey,
  API_KEY_TYPE_LABELS,
} from "@/types/api-key";

interface APIKeyCardProps {
  apiKey: APIKey;
  onRotate?: (key: APIKey) => void;
  onDelete?: (key: APIKey) => void;
  onEdit?: (key: APIKey) => void;
}

export function APIKeyCard({
  apiKey,
  onRotate,
  onDelete,
  onEdit,
}: APIKeyCardProps) {
  const [copied, setCopied] = useState(false);

  // Copy the visible identifier (the key prefix). The full plaintext is only
  // available at creation/rotation time and is not stored on the API key
  // record. We surface this clearly so users know what they're copying.
  const handleCopyKey = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const value = `${apiKey.key_prefix}${"•".repeat(12)}`;
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      toast.success("API key identifier copied to clipboard");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Failed to copy. Your browser may block clipboard access.");
    }
  };

  const handleRotate = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onRotate?.(apiKey);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onDelete?.(apiKey);
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    onEdit?.(apiKey);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const getKeyTypeBadgeVariant = (type: string) => {
    switch (type) {
      case "platform":
        return "default";
      case "function":
        return "secondary";
      case "agent":
        return "outline";
      case "environment":
        return "outline";
      case "oauth":
        return "outline";
      default:
        return "secondary";
    }
  };

  return (
    <Link to={`/dashboard/api-keys/${apiKey.id}`}>
      <Card className="hover:shadow-md transition-shadow cursor-pointer">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                <Key className="w-5 h-5 text-primary" />
              </div>
              <div>
                <CardTitle className="text-lg">{apiKey.name}</CardTitle>
                <CardDescription className="text-xs flex items-center gap-1">
                  {apiKey.key_prefix}{"•".repeat(12)}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-5 w-5"
                    onClick={handleCopyKey}
                    aria-label="Copy key identifier"
                  >
                    {copied ? (
                      <Check className="w-3 h-3 text-green-600" />
                    ) : (
                      <Copy className="w-3 h-3" />
                    )}
                  </Button>
                </CardDescription>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={apiKey.is_active ? "default" : "secondary"}>
                {apiKey.is_active ? "Active" : "Inactive"}
              </Badge>
              <Badge variant={getKeyTypeBadgeVariant(apiKey.key_type)}>
                {API_KEY_TYPE_LABELS[apiKey.key_type]}
              </Badge>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {apiKey.description && (
            <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
              {apiKey.description}
            </p>
          )}
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center gap-4">
              <span>Created: {formatDate(apiKey.created_at)}</span>
              <span>Last used: {formatDate(apiKey.last_used_at)}</span>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <MoreVertical className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {onEdit && (
                  <DropdownMenuItem onClick={handleEdit}>
                    <Edit className="w-4 h-4 mr-2" />
                    Edit
                  </DropdownMenuItem>
                )}
                {onRotate && (
                  <DropdownMenuItem onClick={handleRotate}>
                    <RotateCcw className="w-4 h-4 mr-2" />
                    Rotate
                  </DropdownMenuItem>
                )}
                {onDelete && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      onClick={handleDelete}
                      className="text-red-600 focus:text-red-600"
                    >
                      <Trash2 className="w-4 h-4 mr-2" />
                      Delete
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
