import { useState } from "react";
import { Link } from "react-router-dom";
import {
  Key,
  Copy,
  Check,
  MoreVertical,
  RotateCcw,
  Trash2,
  Edit,
  Eye,
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
import { cn } from "@/lib/utils";

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

  const handleCopyKey = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    // Note: We can't copy the actual key here as it's not available in the API response
    // This would be used in the details view where the key is shown
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
                <CardDescription className="text-xs">
                  {apiKey.key_prefix}•••••••••
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
                <DropdownMenuItem onClick={handleEdit}>
                  <Edit className="w-4 h-4 mr-2" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={handleRotate}>
                  <RotateCcw className="w-4 h-4 mr-2" />
                  Rotate
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={handleDelete}
                  className="text-red-600 focus:text-red-600"
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
