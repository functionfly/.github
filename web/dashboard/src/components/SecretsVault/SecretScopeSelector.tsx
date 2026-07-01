/**
 * SecretScopeSelector - Multi-select component for secret scopes
 *
 * Provides a visual interface for selecting and managing secret scopes with
 * hierarchical display, search/filtering, and visual indicators for scope types.
 * Supports bulk operations and validation for required scopes.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <SecretScopeSelector
 *   selectedScopes={["production", "api"]}
 *   onChange={(scopes) => setScopes(scopes)}
 * />
 *
 * // With required scopes validation
 * <SecretScopeSelector
 *   selectedScopes={scopes}
 *   requiredScopes={["environment"]}
 *   onChange={setScopes}
 *   onValidationError={(errors) => console.log(errors)}
 * />
 *
 * // Loading state
 * <SecretScopeSelector isLoading />
 *
 * // With inherited scopes (read-only)
 * <SecretScopeSelector
 *   selectedScopes={directScopes}
 *   inheritedScopes={["global", "organization"]}
 *   onChange={setDirectScopes}
 * />
 * ```
 */

import { useState, useCallback, useMemo } from "react";
import {
  X,
  Check,
  Search,
  Layers,
  Globe,
  Box,
  Plus,
  Trash2,
  AlertCircle,
  ChevronDown,
  Shield,
  FolderTree,
  Loader2,
} from "lucide-react";
import { cn } from "@/lib/utils";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Separator } from "@/components/ui/separator";

/** Scope type for categorization */
export type ScopeType = "function" | "environment" | "global" | "organization" | "custom";

/** Scope definition with metadata */
export interface Scope {
  id: string;
  name: string;
  type: ScopeType;
  description?: string;
  parentId?: string;
  isSystem?: boolean;
}

/** Extended scope with selection state */
export interface ScopeWithSelection extends Scope {
  isSelected: boolean;
  isInherited: boolean;
}

export interface SecretScopeSelectorProps {
  /** Currently selected scope IDs */
  selectedScopes?: string[];
  /** Scope IDs inherited from parent (read-only) */
  inheritedScopes?: string[];
  /** Available scopes to select from */
  availableScopes?: Scope[];
  /** Required scope types that must be selected */
  requiredScopes?: string[];
  /** Callback when selection changes */
  onChange?: (scopes: string[]) => void;
  /** Callback when validation errors occur */
  onValidationError?: (errors: string[]) => void;
  /** Whether the component is in loading state */
  isLoading?: boolean;
  /** Whether the selector is disabled */
  disabled?: boolean;
  /** Placeholder text when no scopes selected */
  placeholder?: string;
  /** Maximum number of scopes allowed */
  maxScopes?: number;
  /** Additional CSS classes */
  className?: string;
}

/** Default scope definitions */
const DEFAULT_SCOPES: Scope[] = [
  { id: "global", name: "Global", type: "global", description: "Available across all functions and environments", isSystem: true },
  { id: "production", name: "Production", type: "environment", description: "Production environment access", isSystem: true },
  { id: "staging", name: "Staging", type: "environment", description: "Staging environment access", isSystem: true },
  { id: "development", name: "Development", type: "environment", description: "Development environment access", isSystem: true },
  { id: "api", name: "API", type: "function", description: "API function access", parentId: "global" },
  { id: "webhook", name: "Webhook Handler", type: "function", description: "Webhook processing functions", parentId: "global" },
  { id: "scheduler", name: "Scheduler", type: "function", description: "Scheduled job functions", parentId: "global" },
  { id: "worker", name: "Background Worker", type: "function", description: "Background processing functions", parentId: "global" },
];

/** Scope type configuration with icons and colors */
const scopeTypeConfig: Record<ScopeType, { icon: typeof Layers; color: string; bgColor: string }> = {
  function: { icon: Box, color: "text-blue-500", bgColor: "bg-blue-500/10" },
  environment: { icon: Layers, color: "text-green-500", bgColor: "bg-green-500/10" },
  global: { icon: Globe, color: "text-purple-500", bgColor: "bg-purple-500/10" },
  organization: { icon: Shield, color: "text-orange-500", bgColor: "bg-orange-500/10" },
  custom: { icon: FolderTree, color: "text-gray-500", bgColor: "bg-gray-500/10" },
};

/** Build scope hierarchy for display */
function buildScopeHierarchy(scopes: Scope[]): Map<string, Scope[]> {
  const hierarchy = new Map<string, Scope[]>();

  scopes.forEach((scope) => {
    const parentId = scope.parentId || "root";
    const siblings = hierarchy.get(parentId) || [];
    siblings.push(scope);
    hierarchy.set(parentId, siblings);
  });

  return hierarchy;
}

/** Validate required scopes */
function validateScopes(
  selected: string[],
  required: string[],
  available: Scope[]
): string[] {
  const errors: string[] = [];

  required.forEach((requiredId) => {
    if (!selected.includes(requiredId)) {
      const scope = available.find((s) => s.id === requiredId);
      errors.push(`Required scope "${scope?.name || requiredId}" must be selected`);
    }
  });

  return errors;
}

/**
 * Skeleton loader for the scope selector
 */
function SecretScopeSelectorSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex flex-wrap gap-2">
        <Skeleton className="h-7 w-24 rounded-full" />
        <Skeleton className="h-7 w-20 rounded-full" />
        <Skeleton className="h-7 w-28 rounded-full" />
      </div>
      <Skeleton className="h-10 w-full rounded-lg" />
    </div>
  );
}

/**
 * Scope badge component with type indicator
 */
function ScopeBadge({
  scope,
  isInherited = false,
  onRemove,
  disabled = false,
}: {
  scope: Scope;
  isInherited?: boolean;
  onRemove?: () => void;
  disabled?: boolean;
}) {
  const config = scopeTypeConfig[scope.type];
  const Icon = config.icon;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge
            variant={isInherited ? "secondary" : "default"}
            className={cn(
              "gap-1.5 px-2.5 py-1 text-xs font-medium",
              isInherited && "opacity-70 cursor-not-allowed"
            )}
          >
            <Icon className={cn("h-3 w-3", config.color)} />
            <span className={isInherited ? "italic" : ""}>
              {scope.name}
              {isInherited && " (inherited)"}
            </span>
            {!isInherited && onRemove && !disabled && (
              <button
                onClick={onRemove}
                className="ml-1 rounded-sm hover:bg-black/10 hover:bg-[rgba(255,255,255,0.04)] focus:outline-none"
                aria-label={`Remove ${scope.name}`}
              >
                <X className="h-3 w-3" />
              </button>
            )}
          </Badge>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <p className="text-xs">{scope.description || scope.name}</p>
          <p className="text-xs text-[var(--text-faint)] capitalize">{scope.type} scope</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * SecretScopeSelector component
 *
 * Multi-select dropdown for managing secret scopes with hierarchical display,
 * search functionality, and visual type indicators.
 */
export function SecretScopeSelector({
  selectedScopes = [],
  inheritedScopes = [],
  availableScopes = DEFAULT_SCOPES,
  requiredScopes = [],
  onChange,
  onValidationError,
  isLoading = false,
  disabled = false,
  placeholder = "Select scopes...",
  maxScopes,
  className,
}: SecretScopeSelectorProps) {
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  // Combine selected and inherited scopes for display
  const allSelectedScopes = useMemo(() => {
    const combined = new Set([...selectedScopes, ...inheritedScopes]);
    return Array.from(combined);
  }, [selectedScopes, inheritedScopes]);

  // Get scope objects for selected IDs
  const selectedScopeObjects = useMemo(() => {
    return allSelectedScopes
      .map((id) => availableScopes.find((s) => s.id === id))
      .filter((s): s is Scope => s !== undefined);
  }, [allSelectedScopes, availableScopes]);

  // Group scopes by type for the dropdown
  const groupedScopes = useMemo(() => {
    const filtered = availableScopes.filter(
      (scope) =>
        !searchQuery ||
        scope.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        scope.description?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const grouped = new Map<ScopeType, Scope[]>();
    filtered.forEach((scope) => {
      const existing = grouped.get(scope.type) || [];
      existing.push(scope);
      grouped.set(scope.type, existing);
    });

    return grouped;
  }, [availableScopes, searchQuery]);

  // Handle scope selection
  const handleSelect = useCallback(
    (scopeId: string) => {
      if (disabled) return;
      if (inheritedScopes.includes(scopeId)) return; // Can't toggle inherited

      const newSelected = selectedScopes.includes(scopeId)
        ? selectedScopes.filter((id) => id !== scopeId)
        : [...selectedScopes, scopeId];

      // Validate max scopes
      if (maxScopes && newSelected.length > maxScopes) {
        onValidationError?.([`Maximum ${maxScopes} scopes allowed`]);
        return;
      }

      // Validate required scopes
      const errors = validateScopes(newSelected, requiredScopes, availableScopes);
      setValidationErrors(errors);
      onValidationError?.(errors);

      onChange?.(newSelected);
    },
    [selectedScopes, inheritedScopes, disabled, maxScopes, requiredScopes, availableScopes, onChange, onValidationError]
  );

  // Handle select all in a group
  const handleSelectAll = useCallback(
    (type: ScopeType) => {
      if (disabled) return;

      const scopesOfType = availableScopes.filter((s) => s.type === type);
      const selectableScopes = scopesOfType.filter(
        (s) => !inheritedScopes.includes(s.id) && !selectedScopes.includes(s.id)
      );

      if (selectableScopes.length === 0) {
        // Deselect all of this type
        const newSelected = selectedScopes.filter(
          (id) => !scopesOfType.some((s) => s.id === id)
        );
        onChange?.(newSelected);
      } else {
        // Select all of this type
        const newScopeIds = selectableScopes.map((s) => s.id);
        const newSelected = [...selectedScopes, ...newScopeIds];

        if (maxScopes && newSelected.length > maxScopes) {
          onValidationError?.([`Maximum ${maxScopes} scopes allowed`]);
          return;
        }

        onChange?.(newSelected);
      }
    },
    [availableScopes, inheritedScopes, selectedScopes, disabled, maxScopes, onChange, onValidationError]
  );

  // Clear all selected (not inherited)
  const handleClearAll = useCallback(() => {
    if (disabled) return;

    const errors = validateScopes([], requiredScopes, availableScopes);
    setValidationErrors(errors);
    onValidationError?.(errors);

    onChange?.([]);
  }, [disabled, requiredScopes, availableScopes, onChange, onValidationError]);

  if (isLoading) {
    return <SecretScopeSelectorSkeleton className={className} />;
  }

  return (
    <div className={cn("space-y-2", className)}>
      {/* Selected scopes display */}
      <div className="flex flex-wrap gap-2 min-h-[36px]">
        {selectedScopeObjects.length === 0 ? (
          <span className="text-sm text-(--color-text-muted) italic">
            {placeholder}
          </span>
        ) : (
          selectedScopeObjects.map((scope) => (
            <ScopeBadge
              key={scope.id}
              scope={scope}
              isInherited={inheritedScopes.includes(scope.id)}
              onRemove={() => handleSelect(scope.id)}
              disabled={disabled}
            />
          ))
        )}
      </div>

      {/* Validation errors */}
      {validationErrors.length > 0 && (
        <div className="flex items-center gap-2 text-xs text-error">
          <AlertCircle className="h-3.5 w-3.5" />
          <span>{validationErrors[0]}</span>
        </div>
      )}

      {/* Scope selector dropdown */}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            disabled={disabled}
            className="w-full justify-between bg-(--color-bg-primary) border-(--border-default) hover:bg-(--color-bg-secondary)"
          >
            <span className="flex items-center gap-2 text-sm text-(--color-text-secondary)">
              <Plus className="h-4 w-4" />
              Add scopes
              {selectedScopes.length > 0 && (
                <Badge variant="secondary" className="ml-2 text-xs">
                  {selectedScopes.length}
                </Badge>
              )}
            </span>
            <ChevronDown className="h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[400px] p-0" align="start">
          <Command className="bg-(--color-bg-primary)">
            <CommandInput
              placeholder="Search scopes..."
              value={searchQuery}
              onValueChange={setSearchQuery}
              className="border-0"
            />
            <CommandList className="max-h-[300px]">
              <CommandEmpty className="py-6 text-center text-sm text-(--color-text-muted)">
                No scopes found.
              </CommandEmpty>

              {Array.from(groupedScopes.entries()).map(([type, scopes]) => {
                const config = scopeTypeConfig[type];
                const Icon = config.icon;
                const allSelected = scopes.every(
                  (s) => selectedScopes.includes(s.id) || inheritedScopes.includes(s.id)
                );

                return (
                  <CommandGroup
                    key={type}
                    heading={
                      <div className="flex items-center justify-between py-2">
                        <div className="flex items-center gap-2">
                          <Icon className={cn("h-4 w-4", config.color)} />
                          <span className="text-xs font-semibold uppercase tracking-wide text-(--color-text-muted)">
                            {type} Scopes
                          </span>
                        </div>
                        <button
                          onClick={() => handleSelectAll(type)}
                          className="text-xs text-[var(--status-ok)] hover:text-[var(--accent)]"
                          disabled={disabled}
                        >
                          {allSelected ? "Deselect all" : "Select all"}
                        </button>
                      </div>
                    }
                    className="border-b border-(--border-subtle) last:border-0"
                  >
                    {scopes.map((scope) => {
                      const isSelected = selectedScopes.includes(scope.id);
                      const isInherited = inheritedScopes.includes(scope.id);

                      return (
                        <CommandItem
                          key={scope.id}
                          value={scope.id}
                          onSelect={() => handleSelect(scope.id)}
                          disabled={isInherited}
                          className={cn(
                            "flex items-center gap-3 px-3 py-2 cursor-pointer",
                            isInherited && "opacity-50 cursor-not-allowed"
                          )}
                        >
                          <div
                            className={cn(
                              "flex h-5 w-5 items-center justify-center rounded border",
                              isSelected || isInherited
                                ? "rgba(143,255,208,0.15) border-[rgba(143,255,208,0.3)]"
                                : "border-(--border-default)"
                            )}
                          >
                            {(isSelected || isInherited) && (
                              <Check className="h-3.5 w-3.5 text-white" />
                            )}
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium text-(--color-text-primary)">
                                {scope.name}
                              </span>
                              {scope.isSystem && (
                                <Badge variant="outline" className="text-[10px] px-1 py-0">
                                  System
                                </Badge>
                              )}
                              {isInherited && (
                                <Badge variant="secondary" className="text-[10px] px-1 py-0">
                                  Inherited
                                </Badge>
                              )}
                            </div>
                            {scope.description && (
                              <p className="text-xs text-(--color-text-muted) truncate">
                                {scope.description}
                              </p>
                            )}
                          </div>
                        </CommandItem>
                      );
                    })}
                  </CommandGroup>
                );
              })}
            </CommandList>
          </Command>

          {/* Footer actions */}
          {selectedScopes.length > 0 && (
            <>
              <Separator className="bg-(--border-subtle)" />
              <div className="flex items-center justify-between p-3">
                <span className="text-xs text-(--color-text-muted)">
                  {selectedScopes.length} selected
                  {maxScopes && ` / ${maxScopes} max`}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleClearAll}
                  disabled={disabled}
                  className="h-8 gap-1.5 text-xs text-error hover:text-error hover:bg-error/10"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  Clear all
                </Button>
              </div>
            </>
          )}
        </PopoverContent>
      </Popover>
    </div>
  );
}

export default SecretScopeSelector;
