'use client';

import { useState, useEffect, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { deployBundle, getDeploymentStatus, type Bundle, type DeploymentStep, type DeploymentStatusResponse } from '@/api/billing';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog';
import { cn } from '@/lib/utils';
import { CheckCircle2, XCircle, Loader2, Rocket, Server, Shield, Sparkles, ChevronRight, ArrowRight, Home, Settings, Mail } from 'lucide-react';

interface DeployWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  bundle: Bundle;
  pricingMode: 'immediate' | 'deferred';
  onDeployComplete?: (appId: string, backendId: string) => void;
  onDeployError?: (error: string) => void;
}

type WizardStep = 'configure' | 'deploying' | 'success' | 'error';

const steps = [
  { id: 'create_subscription', label: 'Subscription', icon: Sparkles },
  { id: 'provision_app', label: 'App', icon: Server },
  { id: 'provision_backend', label: 'Backend', icon: Server },
  { id: 'provision_functions', label: 'Functions', icon: Rocket },
  { id: 'provision_auth', label: 'Auth', icon: Shield },
  { id: 'provision_email_workflows', label: 'Email', icon: Mail },
  { id: 'finalize', label: 'Finalizing', icon: CheckCircle2 },
];

const regionOptions = [
  { value: 'eu-central-1', label: 'Germany (Frankfurt)', flag: '🇩🇪' },
  { value: 'us-east-1', label: 'USA (N. Virginia)', flag: '🇺🇸' },
];

export function DeployWizard({ open, onOpenChange, bundle, pricingMode, onDeployComplete, onDeployError }: DeployWizardProps) {
  const [currentStep, setCurrentStep] = useState<WizardStep>('configure');
  const [deploymentId, setDeploymentId] = useState<string | null>(null);
  const [deploymentSteps, setDeploymentSteps] = useState<DeploymentStep[]>([]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [appId, setAppId] = useState<string | null>(null);
  const [backendId, setBackendId] = useState<string | null>(null);
  const [selectedRegion, setSelectedRegion] = useState('us-east-1');
  const [isDeploying, setIsDeploying] = useState(false);

  const pollDeploymentStatus = useCallback(async (id: string) => {
    try {
      const status = await getDeploymentStatus(id);
      setDeploymentSteps(status.steps);

      if (status.status === 'completed') {
        setCurrentStep('success');
        setAppId(status.app_id || null);
        setBackendId(status.backend_id || null);
        if (onDeployComplete && status.app_id) {
          onDeployComplete(status.app_id, status.backend_id || '');
        }
      } else if (status.status === 'failed') {
        setCurrentStep('error');
        setErrorMessage(status.error || 'Deployment failed');
        if (onDeployError) {
          onDeployError(status.error || 'Deployment failed');
        }
      }
      return status.status;
    } catch (err) {
      console.error('Failed to poll deployment status:', err);
      return 'error';
    }
  }, [onDeployComplete, onDeployError]);

  useEffect(() => {
    if (!open) {
      setCurrentStep('configure');
      setDeploymentId(null);
      setDeploymentSteps([]);
      setErrorMessage(null);
      setAppId(null);
      setBackendId(null);
      setIsDeploying(false);
    }
  }, [open]);

  useEffect(() => {
    let pollInterval: NodeJS.Timeout;

    if (currentStep === 'deploying' && deploymentId) {
      pollInterval = setInterval(async () => {
        const status = await pollDeploymentStatus(deploymentId);
        if (status === 'completed' || status === 'failed') {
          clearInterval(pollInterval);
        }
      }, 2000);
    }

    return () => {
      if (pollInterval) {
        clearInterval(pollInterval);
      }
    };
  }, [currentStep, deploymentId, pollDeploymentStatus]);

  const handleStartDeploy = async () => {
    setIsDeploying(true);
    setCurrentStep('deploying');

    try {
      const response = await deployBundle(bundle.slug, selectedRegion);
      setDeploymentId(response.deployment_id);
      if (response.steps) {
        setDeploymentSteps(response.steps);
      }
    } catch (err: any) {
      console.error('Failed to start deployment:', err);
      setCurrentStep('error');
      setErrorMessage(err?.response?.data?.error || err?.message || 'Failed to start deployment');
      if (onDeployError) {
        onDeployError(err?.response?.data?.error || err?.message || 'Failed to start deployment');
      }
    } finally {
      setIsDeploying(false);
    }
  };

  const handleClose = () => {
    onOpenChange(false);
  };

  const handleGoToDashboard = () => {
    if (appId) {
      window.location.href = `/dashboard/apps/${appId}`;
    } else {
      window.location.href = '/dashboard';
    }
  };

  const handleViewBackend = () => {
    if (appId) {
      window.location.href = `/dashboard/apps/${appId}?tab=backends`;
    } else {
      window.location.href = '/dashboard';
    }
  };

  const handleDeployAnother = () => {
    setCurrentStep('configure');
    setDeploymentId(null);
    setDeploymentSteps([]);
    setErrorMessage(null);
    setAppId(null);
    setBackendId(null);
  };

  const progress = deploymentSteps.length > 0
    ? (deploymentSteps.filter(s => s.status === 'completed').length / deploymentSteps.length) * 100
    : 0;

  const currentStepId = deploymentSteps.find(s => s.status === 'in_progress')?.id;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {currentStep === 'configure' && (
              <>
                <Rocket className="w-5 h-5 text-brand-500" />
                Deploy {bundle.display_name}
              </>
            )}
            {currentStep === 'deploying' && (
              <>
                <Loader2 className="w-5 h-5 text-brand-500 animate-spin" />
                Deploying...
              </>
            )}
            {currentStep === 'success' && (
              <>
                <CheckCircle2 className="w-5 h-5 text-success" />
                Deployment Complete
              </>
            )}
            {currentStep === 'error' && (
              <>
                <XCircle className="w-5 h-5 text-error" />
                Deployment Failed
              </>
            )}
          </DialogTitle>
          <DialogDescription>
            {currentStep === 'configure' && 'Configure your deployment settings'}
            {currentStep === 'deploying' && `Deploying ${bundle.display_name} to your account`}
            {currentStep === 'success' && 'Your backend is ready to use'}
            {currentStep === 'error' && errorMessage}
          </DialogDescription>
        </DialogHeader>

        {/* Configure Step */}
        {currentStep === 'configure' && (
          <div className="space-y-6 py-4">
            <div className="bg-bg-tertiary rounded-lg p-4">
              <div className="flex items-center gap-3">
                <div className={`w-10 h-10 rounded-lg bg-gradient-to-r ${bundle.color} flex items-center justify-center`}>
                  <Rocket className="w-5 h-5 text-white" />
                </div>
                <div>
                  <p className="font-semibold text-text-primary">{bundle.display_name}</p>
                  <p className="text-sm text-text-secondary">{bundle.short_description}</p>
                </div>
              </div>
            </div>

            <div className="space-y-3">
              <label className="block text-sm font-medium text-text-primary">
                Deployment Region
              </label>
              <div className="grid grid-cols-2 gap-2">
                {regionOptions.map((region) => (
                  <button
                    key={region.value}
                    onClick={() => setSelectedRegion(region.value)}
                    className={cn(
                      'flex items-center gap-2 p-3 rounded-lg border transition-all text-left',
                      selectedRegion === region.value
                        ? 'border-brand-500 bg-brand-500/10 text-brand-500'
                        : 'border-border bg-bg-secondary hover:border-brand-500/50'
                    )}
                  >
                    <span className="text-lg">{region.flag}</span>
                    <span className="text-sm font-medium">{region.label}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="bg-info/10 dark:bg-info/20 rounded-lg p-4">
              <p className="text-sm text-info">
                <strong>What's included:</strong> App, Backend, Functions, Auth setup, Email workflows, and more.
                {pricingMode === 'deferred' && ' You can start for free and pay later.'}
              </p>
            </div>
          </div>
        )}

        {/* Deploying Step */}
        {currentStep === 'deploying' && (
          <div className="space-y-6 py-4">
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-text-secondary">Overall Progress</span>
                <span className="font-medium text-text-primary">{Math.round(progress)}%</span>
              </div>
              <div className="h-2 bg-bg-tertiary rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-brand-500 to-brand-400 transition-all duration-500"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>

            <div className="space-y-3">
              {steps.map((step, index) => {
                const stepData = deploymentSteps.find(s => s.id === step.id);
                const status = stepData?.status || 'pending';
                const isActive = currentStepId === step.id;
                const isComplete = status === 'completed';
                const isFailed = status === 'failed';

                return (
                  <div
                    key={step.id}
                    className={cn(
                      'flex items-center gap-3 p-3 rounded-lg border transition-all',
                      isActive && 'border-brand-500 bg-brand-500/10',
                      isComplete && 'border-success/50 bg-success/10',
                      isFailed && 'border-error/50 bg-error/10',
                      !isActive && !isComplete && !isFailed && 'border-border bg-bg-secondary'
                    )}
                  >
                    <div className={cn(
                      'w-8 h-8 rounded-full flex items-center justify-center',
                      isActive && 'bg-brand-500 text-white',
                      isComplete && 'bg-success text-white',
                      isFailed && 'bg-error text-white',
                      !isActive && !isComplete && !isFailed && 'bg-bg-tertiary text-text-muted'
                    )}>
                      {isActive && <Loader2 className="w-4 h-4 animate-spin" />}
                      {isComplete && <CheckCircle2 className="w-4 h-4" />}
                      {isFailed && <XCircle className="w-4 h-4" />}
                      {!isActive && !isComplete && !isFailed && (
                        <step.icon className="w-4 h-4" />
                      )}
                    </div>
                    <div className="flex-1">
                      <p className={cn(
                        'font-medium text-sm',
                        isActive && 'text-brand-500',
                        isComplete && 'text-success',
                        isFailed && 'text-error',
                        !isActive && !isComplete && !isFailed && 'text-text-muted'
                      )}>
                        {step.label}
                      </p>
                      {isActive && stepData?.description && (
                        <p className="text-xs text-text-secondary">{stepData.description}</p>
                      )}
                      {isFailed && stepData?.error && (
                        <p className="text-xs text-error">{stepData.error}</p>
                      )}
                    </div>
                    {isActive && (
                      <div className="flex items-center gap-1 text-xs text-brand-500">
                        <Loader2 className="w-3 h-3 animate-spin" />
                        Running
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Success Step */}
        {currentStep === 'success' && (
          <div className="space-y-6 py-4">
            <div className="text-center py-8">
              <div className="w-16 h-16 bg-success/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <CheckCircle2 className="w-8 h-8 text-success" />
              </div>
              <h3 className="text-xl font-bold text-text-primary mb-2">
                {bundle.display_name} Deployed!
              </h3>
              <p className="text-text-secondary">
                Your backend is ready to use. All functions and auth have been configured.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-3">
              {appId && (
                <div className="bg-bg-tertiary rounded-lg p-3">
                  <p className="text-xs text-text-muted mb-1">App ID</p>
                  <p className="font-mono text-sm text-text-primary truncate">{appId}</p>
                </div>
              )}
              {backendId && (
                <div className="bg-bg-tertiary rounded-lg p-3">
                  <p className="text-xs text-text-muted mb-1">Backend ID</p>
                  <p className="font-mono text-sm text-text-primary truncate">{backendId}</p>
                </div>
              )}
            </div>

            <div className="bg-success/10 dark:bg-success/20 rounded-lg p-4">
              <p className="text-sm text-success font-medium mb-2">What's ready:</p>
              <ul className="text-sm text-text-secondary space-y-1">
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-success" />
                  App with default configuration
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-success" />
                  Backend infrastructure
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-success" />
                  Pre-configured function templates
                </li>
                <li className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-success" />
                  Auth provider setup
                </li>
              </ul>
            </div>
          </div>
        )}

        {/* Error Step */}
        {currentStep === 'error' && (
          <div className="space-y-6 py-4">
            <div className="text-center py-8">
              <div className="w-16 h-16 bg-error/20 rounded-full flex items-center justify-center mx-auto mb-4">
                <XCircle className="w-8 h-8 text-error" />
              </div>
              <h3 className="text-xl font-bold text-text-primary mb-2">
                Deployment Failed
              </h3>
              <p className="text-text-secondary">
                {errorMessage || 'An unexpected error occurred during deployment.'}
              </p>
            </div>

            <div className="bg-error/10 dark:bg-error/20 rounded-lg p-4">
              <p className="text-sm text-error font-medium mb-2">What you can do:</p>
              <ul className="text-sm text-text-secondary space-y-1">
                <li className="flex items-center gap-2">
                  <ArrowRight className="w-4 h-4 text-error" />
                  Try again in a few minutes
                </li>
                <li className="flex items-center gap-2">
                  <ArrowRight className="w-4 h-4 text-error" />
                  Contact support if the issue persists
                </li>
              </ul>
            </div>
          </div>
        )}

        <DialogFooter>
          {currentStep === 'configure' && (
            <>
              <Button variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                onClick={handleStartDeploy}
                disabled={isDeploying}
                className="bg-gradient-to-r from-brand-500 to-brand-400 hover:from-ff-flame hover:to-ff-afterburner"
              >
                {isDeploying ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin mr-2" />
                    Starting...
                  </>
                ) : (
                  <>
                    <Rocket className="w-4 h-4 mr-2" />
                    Start Deployment
                  </>
                )}
              </Button>
            </>
          )}

          {currentStep === 'deploying' && (
            <Button variant="outline" onClick={handleClose} disabled>
              <Loader2 className="w-4 h-4 animate-spin mr-2" />
              Deploying...
            </Button>
          )}

          {currentStep === 'success' && (
            <>
              <Button variant="outline" onClick={handleDeployAnother}>
                Deploy Another
              </Button>
              <Button onClick={handleViewBackend} variant="outline">
                <Settings className="w-4 h-4 mr-2" />
                View Backend
              </Button>
              <Button
                onClick={handleGoToDashboard}
                className="bg-gradient-to-r from-brand-500 to-brand-400 hover:from-ff-flame hover:to-ff-afterburner"
              >
                <Home className="w-4 h-4 mr-2" />
                Go to Dashboard
              </Button>
            </>
          )}

          {currentStep === 'error' && (
            <>
              <Button variant="outline" onClick={handleClose}>
                Close
              </Button>
              <Button onClick={handleDeployAnother}>
                Try Again
              </Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}