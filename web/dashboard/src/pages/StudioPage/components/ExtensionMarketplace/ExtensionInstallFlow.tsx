import { useState } from "react";
import { Button, GlassCard, Badge, Spinner } from "@functionfly/ui-core";
import { Shield, AlertTriangle, Check, X, ChevronRight, FileCode, Lock } from "lucide-react";
import type { Extension, SandboxConfig } from "@/api/marketplace";

interface ExtensionInstallFlowProps {
  extension: Extension;
  onInstall: (extensionId: string, sandboxConfig?: SandboxConfig) => void;
  onCancel: () => void;
}

interface Permission {
  type: string;
  description: string;
  required: boolean;
  risk: "low" | "medium" | "high";
}

export function ExtensionInstallFlow({ extension, onInstall, onCancel }: ExtensionInstallFlowProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [sandboxTier, setSandboxTier] = useState<"wasm" | "worker" | "microvm" | "enterprise">("wasm");
  const [acceptedRisks, setAcceptedRisks] = useState<Set<string>>(new Set());
  const [isInstalling, setIsInstalling] = useState(false);

  const steps = [
    { id: 0, name: "Review", description: "Review extension details" },
    { id: 1, name: "Permissions", description: "Grant required permissions" },
    { id: 2, name: "Sandbox", description: "Configure sandbox settings" },
    { id: 3, name: "Install", description: "Confirm and install" },
  ];

  const permissions: Permission[] = [
    { type: "webhook", description: "Send webhooks to external URLs", required: true, risk: "medium" },
    { type: "notifications", description: "Send push notifications", required: false, risk: "low" },
    { type: "filesystem", description: "Read/write local files", required: false, risk: "high" },
    { type: "network", description: "Make network requests", required: true, risk: "medium" },
    { type: "environment", description: "Access environment variables", required: true, risk: "high" },
  ];

  const sandboxOptions = [
    { value: "wasm", label: "WebAssembly", description: "Lightweight sandbox with minimal capabilities", risk: "low" },
    { value: "worker", label: "Sandboxed Worker", description: "Isolated worker with controlled resources", risk: "medium" },
    { value: "microvm", label: "MicroVM", description: "Full isolation with dedicated resources", risk: "low" },
    { value: "enterprise", label: "Enterprise", description: "Maximum isolation for sensitive workloads", risk: "low" },
  ] as const;

  const requiredPermissions = permissions.filter((p) => p.required);
  const optionalPermissions = permissions.filter((p) => !p.required);
  const unacceptedRisks = permissions.filter((p) => p.risk === "high" && !acceptedRisks.has(p.type));

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleInstall = async () => {
    setIsInstalling(true);
    try {
      onInstall(extension.id, { tier: sandboxTier });
    } finally {
      setIsInstalling(false);
    }
  };

  const canProceed = () => {
    switch (currentStep) {
      case 0:
        return true;
      case 1:
        return requiredPermissions.every((p) => acceptedRisks.has(p.type));
      case 2:
        return true;
      case 3:
        return unacceptedRisks.length === 0;
      default:
        return false;
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-sm text-white/60">
        {steps.map((step, index) => (
          <div key={step.id} className="flex items-center">
            <button
              onClick={() => setCurrentStep(index)}
              className={`flex items-center gap-1 px-2 py-1 rounded ${
                currentStep === index
                  ? "bg-brand-500/20 text-brand-400"
                  : index < currentStep
                  ? "text-green-400"
                  : "text-white/40"
              }`}
            >
              <span className="w-5 h-5 rounded-full bg-white/10 flex items-center justify-center text-[10px]">
                {index < currentStep ? <Check className="w-3 h-3" /> : index + 1}
              </span>
              <span className="hidden sm:inline">{step.name}</span>
            </button>
            {index < steps.length - 1 && <ChevronRight className="w-4 h-4 mx-1" />}
          </div>
        ))}
      </div>

      <GlassCard className="p-4">
        {currentStep === 0 && (
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-lg bg-bg-secondary flex items-center justify-center">
                {extension.icon_url ? (
                  <img src={extension.icon_url} alt={extension.name} className="w-full h-full object-cover rounded" />
                ) : (
                  <span className="text-xl font-bold text-white/40">{extension.name[0]}</span>
                )}
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white">{extension.name}</h3>
                <p className="text-sm text-white/60">v{extension.version} • by {extension.creator_id}</p>
              </div>
            </div>
            <p className="text-sm text-white/80">{extension.description}</p>
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline" size="sm">{extension.category}</Badge>
              {extension.tags?.map((tag) => (
                <Badge key={tag} variant="ghost" size="sm">{tag}</Badge>
              ))}
            </div>
          </div>
        )}

        {currentStep === 1 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-blue-400" />
              <h3 className="text-lg font-semibold text-white">Permissions Required</h3>
            </div>
            <p className="text-sm text-white/60">This extension requires the following permissions to function:</p>

            <div className="space-y-2">
              {requiredPermissions.map((perm) => (
                <div
                  key={perm.type}
                  className="flex items-center gap-3 p-3 rounded-lg bg-white/5 border border-white/10"
                >
                  <button
                    onClick={() => {
                      const newSet = new Set(acceptedRisks);
                      if (newSet.has(perm.type)) {
                        newSet.delete(perm.type);
                      } else {
                        newSet.add(perm.type);
                      }
                      setAcceptedRisks(newSet);
                    }}
                    className={`w-5 h-5 rounded border flex items-center justify-center ${
                      acceptedRisks.has(perm.type)
                        ? "bg-green-500 border-green-500"
                        : "border-white/30"
                    }`}
                  >
                    {acceptedRisks.has(perm.type) && <Check className="w-3 h-3 text-white" />}
                  </button>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-white">{perm.type}</span>
                      {perm.risk === "high" && (
                        <Badge variant="destructive" size="sm">
                          <AlertTriangle className="w-3 h-3 mr-1" />
                          High Risk
                        </Badge>
                      )}
                    </div>
                    <p className="text-xs text-white/60">{perm.description}</p>
                  </div>
                </div>
              ))}
            </div>

            {optionalPermissions.length > 0 && (
              <>
                <h4 className="text-sm font-medium text-white/80 pt-2">Optional Permissions</h4>
                <div className="space-y-2">
                  {optionalPermissions.map((perm) => (
                    <div
                      key={perm.type}
                      className="flex items-center gap-3 p-3 rounded-lg bg-white/5 border border-white/10"
                    >
                      <button
                        onClick={() => {
                          const newSet = new Set(acceptedRisks);
                          if (newSet.has(perm.type)) {
                            newSet.delete(perm.type);
                          } else {
                            newSet.add(perm.type);
                          }
                          setAcceptedRisks(newSet);
                        }}
                        className={`w-5 h-5 rounded border flex items-center justify-center ${
                          acceptedRisks.has(perm.type)
                            ? "bg-brand-500 border-brand-500"
                            : "border-white/30"
                        }`}
                      >
                        {acceptedRisks.has(perm.type) && <Check className="w-3 h-3 text-white" />}
                      </button>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-white">{perm.type}</span>
                        </div>
                        <p className="text-xs text-white/60">{perm.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        )}

        {currentStep === 2 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Lock className="w-5 h-5 text-purple-400" />
              <h3 className="text-lg font-semibold text-white">Sandbox Configuration</h3>
            </div>
            <p className="text-sm text-white/60">Choose the isolation level for this extension:</p>

            <div className="space-y-2">
              {sandboxOptions.map((option) => (
                <button
                  key={option.value}
                  onClick={() => setSandboxTier(option.value)}
                  className={`w-full p-3 rounded-lg border text-left transition-colors ${
                    sandboxTier === option.value
                      ? "bg-brand-500/20 border-brand-500"
                      : "bg-white/5 border-white/10 hover:border-white/20"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-white">{option.label}</span>
                        <Badge
                          variant={option.risk === "low" ? "success" : "warning"}
                          size="sm"
                        >
                          {option.risk} risk
                        </Badge>
                      </div>
                      <p className="text-xs text-white/60 mt-1">{option.description}</p>
                    </div>
                    <div className={`w-4 h-4 rounded-full border-2 ${
                      sandboxTier === option.value
                        ? "border-brand-500 bg-brand-500"
                        : "border-white/30"
                    }`}>
                      {sandboxTier === option.value && (
                        <div className="w-full h-full rounded-full bg-white scale-[0.6]" />
                      )}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )}

        {currentStep === 3 && (
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-white">Confirm Installation</h3>

            <div className="p-4 rounded-lg bg-white/5 border border-white/10 space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-white/60">Extension</span>
                <span className="text-white">{extension.name}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-white/60">Version</span>
                <span className="text-white">v{extension.version}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-white/60">Sandbox</span>
                <span className="text-white">{sandboxTier}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-white/60">Permissions</span>
                <span className="text-white">{acceptedRisks.size}</span>
              </div>
            </div>

            {unacceptedRisks.length > 0 && (
              <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/30">
                <div className="flex items-center gap-2 text-red-400">
                  <AlertTriangle className="w-4 h-4" />
                  <span className="text-sm font-medium">High-risk permissions not accepted</span>
                </div>
                <p className="text-xs text-red-300/60 mt-1">
                  Please review and accept the following permissions: {unacceptedRisks.map((p) => p.type).join(", ")}
                </p>
              </div>
            )}
          </div>
        )}
      </GlassCard>

      <div className="flex items-center justify-between">
        <Button variant="ghost" onClick={currentStep === 0 ? onCancel : handleBack}>
          {currentStep === 0 ? "Cancel" : "Back"}
        </Button>
        <div className="flex items-center gap-2">
          {currentStep < steps.length - 1 ? (
            <Button onClick={handleNext} disabled={!canProceed()}>
              Next
              <ChevronRight className="w-4 h-4 ml-1" />
            </Button>
          ) : (
            <Button onClick={handleInstall} disabled={!canProceed() || isInstalling}>
              {isInstalling ? (
                <>
                  <Spinner className="w-4 h-4 mr-1" />
                  Installing...
                </>
              ) : (
                <>
                  <FileCode className="w-4 h-4 mr-1" />
                  Install Extension
                </>
              )}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}