import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import {
  Loader2,
  Cloud,
  Globe,
  Code,
  Settings,
  AlertCircle,
  CheckCircle2,
  Info
} from "lucide-react";
import { ProviderIcon } from "./ProviderIcon";

const deploySchema = z.object({
  functionName: z.string().min(1, "Function name is required").regex(/^[a-zA-Z0-9_-]+$/, "Function name can only contain letters, numbers, hyphens, and underscores"),
  version: z.string().min(1, "Version is required").regex(/^\d+\.\d+\.\d+$/, "Version must be in format x.y.z"),
  description: z.string().optional(),
  runtime: z.enum(["node", "python", "deno", "go", "rust"], {
    required_error: "Please select a runtime",
  }),
  provider: z.enum(["cloudflare", "vercel", "fly", "deno", "functionfly-edge"], {
    required_error: "Please select a provider",
  }),
  environment: z.enum(["production", "staging", "development"], {
    required_error: "Please select an environment",
  }),
  domain: z.string().optional(),
  enableCORS: z.boolean().default(false),
  enableAuth: z.boolean().default(false),
  enableCaching: z.boolean().default(false),
  memoryLimit: z.number().min(128).max(10240).optional(),
  timeout: z.number().min(5).max(900).optional(),
  environmentVariables: z.record(z.string()).optional(),
});

type DeployFormData = z.infer<typeof deploySchema>;

interface DeployModalProps {
  isOpen: boolean;
  onClose: () => void;
  onDeploy: (data: DeployFormData) => Promise<void>;
  initialData?: Partial<DeployFormData>;
  functionInfo?: {
    name: string;
    version: string;
    description?: string;
    runtime?: string;
  };
}

const runtimeOptions = [
  { value: "node", label: "Node.js", icon: Code },
  { value: "python", label: "Python", icon: Code },
  { value: "deno", label: "Deno", icon: Globe },
  { value: "go", label: "Go", icon: Settings },
  { value: "rust", label: "Rust", icon: Settings },
] as const;

const providerOptions = [
  { value: "cloudflare", label: "Cloudflare Workers", color: "#f48120" },
  { value: "vercel", label: "Vercel", color: "#ffffff" },
  { value: "fly", label: "Fly.io", color: "#7b68ee" },
  { value: "deno", label: "Deno Deploy", color: "#ffffff" },
  { value: "functionfly-edge", label: "FunctionFly Edge", color: "#6366f1" },
] as const;

const environmentOptions = [
  { value: "development", label: "Development", color: "#10b981" },
  { value: "staging", label: "Staging", color: "#f59e0b" },
  { value: "production", label: "Production", color: "#ef4444" },
] as const;

export function DeployModal({
  isOpen,
  onClose,
  onDeploy,
  initialData,
  functionInfo,
}: DeployModalProps) {
  const [isDeploying, setIsDeploying] = useState(false);
  const [deployStatus, setDeployStatus] = useState<"idle" | "deploying" | "success" | "error">("idle");
  const [deployError, setDeployError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors, isValid },
  } = useForm<DeployFormData>({
    resolver: zodResolver(deploySchema),
    defaultValues: {
      functionName: functionInfo?.name || initialData?.functionName || "",
      version: functionInfo?.version || initialData?.version || "1.0.0",
      description: functionInfo?.description || initialData?.description || "",
      runtime: (functionInfo?.runtime as any) || initialData?.runtime || undefined,
      provider: initialData?.provider || undefined,
      environment: initialData?.environment || "development",
      domain: initialData?.domain || "",
      enableCORS: initialData?.enableCORS || false,
      enableAuth: initialData?.enableAuth || false,
      enableCaching: initialData?.enableCaching || false,
      memoryLimit: initialData?.memoryLimit || 256,
      timeout: initialData?.timeout || 30,
      environmentVariables: initialData?.environmentVariables || {},
    },
  });

  const watchedProvider = watch("provider");
  const watchedEnvironment = watch("environment");
  const watchedRuntime = watch("runtime");

  const handleDeploy = async (data: DeployFormData) => {
    setIsDeploying(true);
    setDeployStatus("deploying");
    setDeployError(null);

    try {
      await onDeploy(data);
      setDeployStatus("success");

      // Auto-close after success
      setTimeout(() => {
        onClose();
        reset();
        setDeployStatus("idle");
      }, 2000);
    } catch (error) {
      setDeployStatus("error");
      setDeployError(error instanceof Error ? error.message : "Deployment failed");
    } finally {
      setIsDeploying(false);
    }
  };

  const handleClose = () => {
    if (!isDeploying) {
      onClose();
      reset();
      setDeployStatus("idle");
      setDeployError(null);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Cloud className="w-5 h-5" />
            Deploy Function
          </DialogTitle>
          <DialogDescription>
            Configure and deploy your function to the selected provider and environment.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(handleDeploy)} className="space-y-6">
          {/* Basic Information */}
          <Card>
            <CardContent className="pt-6">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="functionName">
                    Function Name <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="functionName"
                    {...register("functionName")}
                    placeholder="my-function"
                    className={cn(errors.functionName && "border-red-500")}
                  />
                  {errors.functionName && (
                    <p className="text-xs text-red-500">{errors.functionName.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="version">
                    Version <span className="text-red-500">*</span>
                  </Label>
                  <Input
                    id="version"
                    {...register("version")}
                    placeholder="1.0.0"
                    className={cn(errors.version && "border-red-500")}
                  />
                  {errors.version && (
                    <p className="text-xs text-red-500">{errors.version.message}</p>
                  )}
                </div>
              </div>

              <div className="space-y-2 mt-4">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  {...register("description")}
                  placeholder="Brief description of your function..."
                  rows={2}
                />
              </div>
            </CardContent>
          </Card>

          {/* Runtime & Provider */}
          <Card>
            <CardContent className="pt-6">
              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>
                    Runtime <span className="text-red-500">*</span>
                  </Label>
                  <Select
                    value={watchedRuntime}
                    onValueChange={(value) => setValue("runtime", value as any)}
                  >
                    <SelectTrigger className={cn(errors.runtime && "border-red-500")}>
                      <SelectValue placeholder="Select runtime" />
                    </SelectTrigger>
                    <SelectContent>
                      {runtimeOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          <div className="flex items-center gap-2">
                            <option.icon className="w-4 h-4" />
                            {option.label}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {errors.runtime && (
                    <p className="text-xs text-red-500">{errors.runtime.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label>
                    Provider <span className="text-red-500">*</span>
                  </Label>
                  <Select
                    value={watchedProvider}
                    onValueChange={(value) => setValue("provider", value as any)}
                  >
                    <SelectTrigger className={cn(errors.provider && "border-red-500")}>
                      <SelectValue placeholder="Select provider" />
                    </SelectTrigger>
                    <SelectContent>
                      {providerOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          <div className="flex items-center gap-2">
                            <ProviderIcon provider={option.value} size="sm" />
                            {option.label}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {errors.provider && (
                    <p className="text-xs text-red-500">{errors.provider.message}</p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label>
                    Environment <span className="text-red-500">*</span>
                  </Label>
                  <Select
                    value={watchedEnvironment}
                    onValueChange={(value) => setValue("environment", value as any)}
                  >
                    <SelectTrigger className={cn(errors.environment && "border-red-500")}>
                      <SelectValue placeholder="Select environment" />
                    </SelectTrigger>
                    <SelectContent>
                      {environmentOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          <div className="flex items-center gap-2">
                            <div
                              className="w-2 h-2 rounded-full"
                              style={{ backgroundColor: option.color }}
                            />
                            {option.label}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {errors.environment && (
                    <p className="text-xs text-red-500">{errors.environment.message}</p>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Advanced Settings */}
          <Card>
            <CardContent className="pt-6">
              <h3 className="text-sm font-medium mb-4">Advanced Settings</h3>

              <div className="grid grid-cols-2 gap-4 mb-4">
                <div className="space-y-2">
                  <Label htmlFor="domain">Custom Domain</Label>
                  <Input
                    id="domain"
                    {...register("domain")}
                    placeholder="api.example.com"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="memoryLimit">Memory Limit (MB)</Label>
                  <Input
                    id="memoryLimit"
                    type="number"
                    {...register("memoryLimit", { valueAsNumber: true })}
                    placeholder="256"
                    min={128}
                    max={10240}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 mb-4">
                <div className="space-y-2">
                  <Label htmlFor="timeout">Timeout (seconds)</Label>
                  <Input
                    id="timeout"
                    type="number"
                    {...register("timeout", { valueAsNumber: true })}
                    placeholder="30"
                    min={5}
                    max={900}
                  />
                </div>
              </div>

              <Separator className="my-4" />

              <div className="space-y-3">
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="enableCORS"
                    {...register("enableCORS")}
                  />
                  <Label htmlFor="enableCORS" className="text-sm">
                    Enable CORS
                  </Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="enableAuth"
                    {...register("enableAuth")}
                  />
                  <Label htmlFor="enableAuth" className="text-sm">
                    Enable Authentication
                  </Label>
                </div>

                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="enableCaching"
                    {...register("enableCaching")}
                  />
                  <Label htmlFor="enableCaching" className="text-sm">
                    Enable Response Caching
                  </Label>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Status Messages */}
          {deployStatus === "error" && deployError && (
            <Alert className="border-red-500/20 bg-red-500/10">
              <AlertCircle className="h-4 w-4 text-red-500" />
              <AlertDescription className="text-red-600 dark:text-red-400">
                {deployError}
              </AlertDescription>
            </Alert>
          )}

          {deployStatus === "success" && (
            <Alert className="border-green-500/20 bg-green-500/10">
              <CheckCircle2 className="h-4 w-4 text-green-500" />
              <AlertDescription className="text-green-600 dark:text-green-400">
                Function deployed successfully!
              </AlertDescription>
            </Alert>
          )}

          {/* Deployment Summary */}
          {watchedProvider && watchedRuntime && (
            <Card className="bg-indigo-500/5 border-indigo-500/20">
              <CardContent className="pt-6">
                <div className="flex items-center gap-2 mb-2">
                  <Info className="w-4 h-4 text-indigo-400" />
                  <span className="text-sm font-medium text-indigo-400">Deployment Summary</span>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Badge variant="secondary" className="bg-indigo-500/10 text-indigo-400">
                    <ProviderIcon provider={watchedProvider} size="sm" className="mr-1" />
                    {providerOptions.find(p => p.value === watchedProvider)?.label}
                  </Badge>
                  <Badge variant="secondary" className="bg-indigo-500/10 text-indigo-400">
                    {runtimeOptions.find(r => r.value === watchedRuntime)?.label}
                  </Badge>
                  <Badge
                    variant="secondary"
                    className="bg-indigo-500/10 text-indigo-400"
                    style={{
                      backgroundColor: `${environmentOptions.find(e => e.value === watchedEnvironment)?.color}20`,
                      color: environmentOptions.find(e => e.value === watchedEnvironment)?.color
                    }}
                  >
                    {environmentOptions.find(e => e.value === watchedEnvironment)?.label}
                  </Badge>
                </div>
              </CardContent>
            </Card>
          )}
        </form>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={isDeploying}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit(handleDeploy)}
            disabled={!isValid || isDeploying}
            className="min-w-[120px]"
          >
            {isDeploying ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Deploying...
              </>
            ) : deployStatus === "success" ? (
              <>
                <CheckCircle2 className="w-4 h-4 mr-2" />
                Deployed
              </>
            ) : (
              <>
                <Cloud className="w-4 h-4 mr-2" />
                Deploy Function
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
