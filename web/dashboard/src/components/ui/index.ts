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
export { DataTable } from "./data-table";
export type { DataTableProps } from "./data-table";
export { FormField } from "./form-field";
export { FormError } from "./form-error";
export { LoadingSpinner } from "./loading-spinner";
export { HelpTooltip } from "./help-tooltip";
export { FileUpload } from "./file-upload";

// ============================================================================
// Toast Components
// ============================================================================

export {
  Toast,
  ToastAction,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from "./toast";

export { useToast, toast } from "./use-toast";

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

// ============================================================================
// Rich Text Editor Components
// ============================================================================

/**
 * RichTextEditor - WYSIWYG editor with secure HTML sanitization
 * Uses DOMPurify for XSS protection. Supports headings, lists, links, etc.
 */
export { RichTextEditor, RichTextViewer } from "./rich-text-editor";
export type { RichTextEditorProps, RichTextViewerProps } from "./rich-text-editor";

// ============================================================================
// Date/Time Picker Components
// ============================================================================

/**
 * DatePicker & DateRangePicker - Calendar selection components
 * Built on react-day-picker with Radix UI popover primitives.
 */
export { DatePicker, DateRangePicker } from "./date-picker";
export type { DatePickerProps, DateRangePickerProps, DateRangeSelection } from "./date-picker";

/**
 * TimePicker - Time selection component with step controls
 * Supports custom hour/minute steps and time constraints.
 */
export { TimePicker } from "./time-picker";
export type { TimePickerProps } from "./time-picker";

// ============================================================================
// Chart Components (Standardized Recharts Wrappers)
// ============================================================================

/**
 * Chart Components - Consistent, styled chart wrappers
 * Pre-configured recharts components matching FunctionFly design system.
 */
export { LineChart, Sparkline } from "./chart-line";
export type { LineChartProps, LineSeries, LineChartData } from "./chart-line";

export { BarChart, SimpleBarChart } from "./chart-bar";
export type { BarChartProps, BarSeries, BarChartData } from "./chart-bar";

export { AreaChart, SparkAreaChart } from "./chart-area";
export type { AreaChartProps, AreaSeries, AreaChartData } from "./chart-area";

export { PieChart } from "./chart-pie";
export type { PieChartProps, PieChartData } from "./chart-pie";

// ============================================================================
// Empty State Components
// ============================================================================

/**
 * EmptyState - Illustrated empty states with actions
 * Replaces generic "No data" text with visual context.
 */
export { EmptyState, emptyStateVariants } from "./empty-state";
export type { EmptyStateProps } from "./empty-state";

// ============================================================================
// Toggle Button Components
// ============================================================================

/**
 * ToggleButtonGroup - Button group for view switching
 * Perfect for list/grid/calendar toggles and filter groups.
 */
export { ToggleButtonGroup, toggleGroupVariants, toggleGroupItemVariants } from "./toggle-button-group";
export type { ToggleButtonGroupProps, ToggleButtonOption } from "./toggle-button-group";

// ============================================================================
// Stepper/Wizard Components
// ============================================================================

/**
 * Stepper - Multi-step progress indicator
 * Supports horizontal and vertical orientations with clickable steps.
 */
export { Stepper, StepperContent, StepItem, StepIndicator, StepConnector, useStepper } from "./stepper";
export type { StepperProps, Step, StepItemProps } from "./stepper";

// ============================================================================
// Color Picker Components
// ============================================================================

/**
 * ColorPicker - Color selection with presets and manual input
 * Supports hex color format with validation.
 */
export { ColorPicker } from "./color-picker";
export type { ColorPickerProps } from "./color-picker";

// ============================================================================
// Domain-Specific Skeleton Components
// ============================================================================

/**
 * Skeleton Components - Loading placeholders for specific content types
 * More contextual than generic Skeleton - shows structure of content loading.
 */
export {
  SkeletonCard,
  SkeletonList,
  SkeletonListItem,
  SkeletonTable,
  SkeletonStats,
  SkeletonForm,
  SkeletonChart,
  SkeletonAvatar,
  skeletonCardVariants,
  SkeletonListVariants,
  SkeletonTableVariants,
} from "./skeleton-domain";
export type {
  SkeletonCardProps,
  SkeletonListProps,
  SkeletonListItemProps,
  SkeletonTableProps,
  SkeletonStatsProps,
  SkeletonFormProps,
  SkeletonChartProps,
  SkeletonAvatarProps,
} from "./skeleton-domain";
