/**
 * UI Components Index
 *
 * Central export file for all custom UI components used throughout the FunctionFly
 * dashboard. Import components from this file for consistent usage.
 *
 * @example
 * ```tsx
 * import { AnimatedGradientCard, ShinyButton, TextGradient } from "@/components/ui";
 * ```
 */

// ============================================================================
// Base UI Components (shadcn/ui)
// ============================================================================

export { Button, buttonVariants } from "./button";
export type { ButtonProps } from "./button";

export {
  Card,
  CardHeader,
  CardFooter,
  CardTitle,
  CardDescription,
  CardContent,
} from "./card";

export { Input } from "./input";
export { Label } from "./label";
export { Textarea } from "./textarea";
export { Checkbox } from "./checkbox";
export { Switch } from "./switch";
export { Select } from "./select";
export { Badge } from "./badge";
export { Avatar, AvatarImage, AvatarFallback } from "./avatar";
export { Separator } from "./separator";
export { Skeleton } from "./skeleton";
export { Progress } from "./progress";
export { Slider } from "./slider";
export { Tabs, TabsList, TabsTrigger, TabsContent } from "./tabs";
export { Dialog, DialogContent, DialogHeader, DialogTitle } from "./dialog";
export {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./dropdown-menu";
export { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "./tooltip";
export { Popover, PopoverContent, PopoverTrigger } from "./popover";
export { Calendar } from "./calendar";
export { Command, CommandInput, CommandList, CommandItem } from "./command";
export { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "./accordion";
export { AlertDialog } from "./alert-dialog";
export { Alert } from "./alert";
export { RadioGroup, RadioGroupItem } from "./radio-group";
export { ScrollArea } from "./scroll-area";
export { Sheet } from "./sheet";
export { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./table";
export { Pagination } from "./pagination";
export { FormField } from "./form-field";
export { FormError } from "./form-error";
export { LoadingSpinner } from "./loading-spinner";
export { HelpTooltip } from "./help-tooltip";
export { FileUpload } from "./file-upload";

// ============================================================================
// Custom "Wow" Effect Components
// ============================================================================

/**
 * AnimatedGradientCard - Card with animated gradient borders
 * Perfect for highlighting premium content and featured items.
 */
export { AnimatedGradientCard } from "./AnimatedGradientCard";
export type { AnimatedGradientCardProps } from "./AnimatedGradientCard";

/**
 * GlassmorphismCard - Glass-like card with backdrop blur
 * Ideal for modern, layered UI designs with depth.
 */
export { GlassmorphismCard } from "./GlassmorphismCard";
export type { GlassmorphismCardProps } from "./GlassmorphismCard";

/**
 * ShinyButton - Button with metallic shine animation on hover
 * Premium button component for CTAs and important actions.
 */
export { ShinyButton, shinyButtonVariants } from "./ShinyButton";
export type { ShinyButtonProps } from "./ShinyButton";

/**
 * ParticleBackground - Floating particles animation
 * Subtle ambient background effect for hero sections.
 */
export { ParticleBackground } from "./ParticleBackground";
export type { ParticleBackgroundProps } from "./ParticleBackground";

/**
 * SpotlightCard - Card with mouse-following spotlight effect
 * Interactive card perfect for blog posts and feature highlights.
 */
export { SpotlightCard } from "./SpotlightCard";
export type { SpotlightCardProps } from "./SpotlightCard";

/**
 * TextGradient - Animated gradient text component
 * Stunning text effect for headlines and attention-grabbing content.
 */
export { TextGradient, DEFAULT_COLORS } from "./TextGradient";
export type { TextGradientProps } from "./TextGradient";

/**
 * UserNotFoundView - Empty state when a profile username doesn't exist
 * Used on public and dashboard profile pages.
 */
export { UserNotFoundView } from "./UserNotFoundView";
export type { UserNotFoundViewProps } from "./UserNotFoundView";
