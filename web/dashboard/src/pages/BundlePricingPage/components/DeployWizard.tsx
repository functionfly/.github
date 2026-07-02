'use client';

import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { deployBundle, getDeploymentStatus, type Bundle, type DeploymentStep } from '@/api/billing';
import { providersApi } from '@/api/providers';
import type { ConnectedProvider } from '@/types';
import { Modal } from '@/components/containment/Modal';
import { SealedButton } from '@/components/containment/SealedButton';
import { FrameButton } from '@/components/containment/FrameButton';
import { StatusPill } from '@/components/containment/StatusPill';
import {
  CheckCircle2,
  XCircle,
  Loader2,
  Rocket,
  Server,
  Shield,
  Sparkles,
  ArrowRight,
  Home,
  Settings,
  Mail,
  Package,
  Cloud,
  Globe,
} from 'lucide-react';

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
  const [connectedProviders, setConnectedProviders] = useState<ConnectedProvider[]>([]);
  const [selectedProvider, setSelectedProvider] = useState<ConnectedProvider | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (open) {
      providersApi.getConnectedProviders().then((providers) => {
        const deployable = providers.filter(p =>
          p.status === 'active' && ['workers', 'vercel', 'deno', 'fly'].includes(p.name)
        );
        setConnectedProviders(deployable);
        if (deployable.length > 0) {
          setSelectedProvider(deployable[0]);
        }
      }).catch(console.error);
    }
  }, [open]);

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
      const response = await deployBundle(bundle.slug, selectedRegion, selectedProvider?.id);
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

  const getTitle = () => {
    switch (currentStep) {
      case 'configure': return `Deploy ${bundle.display_name}`;
      case 'deploying': return 'Deploying...';
      case 'success': return 'Deployment Complete';
      case 'error': return 'Deployment Failed';
    }
  };

  const getDescription = () => {
    switch (currentStep) {
      case 'configure': return 'Configure your deployment settings';
      case 'deploying': return `Deploying ${bundle.display_name} to your account`;
      case 'success': return 'Your backend is ready to use';
      case 'error': return errorMessage || 'An unexpected error occurred';
    }
  };

  return (
    <Modal open={open} onClose={handleClose} title={getTitle()} className="sc-deploy-wizard__modal">
      <div className="sc-deploy-wizard">
        {/* Description */}
        <p className="sc-deploy-wizard__desc">{getDescription()}</p>

        {/* Configure Step */}
        {currentStep === 'configure' && (
          <div className="sc-deploy-wizard__body">
            {/* Bundle Info */}
            <div className="sc-deploy-wizard__bundle">
              <div className="sc-deploy-wizard__bundle-icon">
                <Rocket size={20} />
              </div>
              <div>
                <p className="sc-deploy-wizard__bundle-name">{bundle.display_name}</p>
                <p className="sc-deploy-wizard__bundle-desc">{bundle.short_description}</p>
              </div>
            </div>

            {/* Region Selector */}
            <div className="sc-deploy-wizard__field">
              <label className="sc-deploy-wizard__label">Deployment Region</label>
              <div className="sc-deploy-wizard__regions">
                {regionOptions.map((region) => (
                  <button
                    key={region.value}
                    onClick={() => setSelectedRegion(region.value)}
                    className={`sc-deploy-wizard__region ${selectedRegion === region.value ? 'active' : ''}`}
                  >
                    <span className="sc-deploy-wizard__region-flag">{region.flag}</span>
                    <span className="sc-deploy-wizard__region-label">{region.label}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Provider Selector */}
            <div className="sc-deploy-wizard__field">
              <label className="sc-deploy-wizard__label">Deploy To</label>
              {connectedProviders.length > 0 ? (
                <div className="sc-deploy-wizard__regions">
                  {connectedProviders.map((p) => (
                    <button
                      key={p.id}
                      onClick={() => setSelectedProvider(p)}
                      className={`sc-deploy-wizard__region ${selectedProvider?.id === p.id ? 'active' : ''}`}
                    >
                      <span className="sc-deploy-wizard__region-flag">
                        {p.name === 'workers' ? '☁️' : p.name === 'vercel' ? '▲' : p.name === 'deno' ? '🦕' : '🚀'}
                      </span>
                      <span className="sc-deploy-wizard__region-label">
                        {p.name === 'workers' ? 'Cloudflare Workers' : p.name === 'vercel' ? 'Vercel' : p.name === 'deno' ? 'Deno Deploy' : 'Fly.io'}
                      </span>
                    </button>
                  ))}
                </div>
              ) : (
                <div className="sc-deploy-wizard__info">
                  <p className="sc-deploy-wizard__info-text">
                    <Cloud size={16} style={{ marginRight: 8, verticalAlign: 'middle' }} />
                    No provider connected. <a href="/dashboard/providers" style={{ color: 'var(--color-primary)' }}>Connect a provider</a> to deploy to your own infrastructure.
                  </p>
                </div>
              )}
            </div>

            {/* What's Included */}
            <div className="sc-deploy-wizard__info">
              <p className="sc-deploy-wizard__info-text">
                <strong>What's included:</strong> App, Backend, Functions, Auth setup, Email workflows, and more.
                {pricingMode === 'deferred' && ' You can start for free and pay later.'}
              </p>
            </div>
          </div>
        )}

        {/* Deploying Step */}
        {currentStep === 'deploying' && (
          <div className="sc-deploy-wizard__body">
            {/* Progress Bar */}
            <div className="sc-deploy-wizard__progress">
              <div className="sc-deploy-wizard__progress-header">
                <span className="sc-deploy-wizard__progress-label">Overall Progress</span>
                <span className="sc-deploy-wizard__progress-value">{Math.round(progress)}%</span>
              </div>
              <div className="sc-deploy-wizard__progress-bar">
                <div
                  className="sc-deploy-wizard__progress-fill"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>

            {/* Step List */}
            <div className="sc-deploy-wizard__steps">
              {steps.map((step) => {
                const stepData = deploymentSteps.find(s => s.id === step.id);
                const status = stepData?.status || 'pending';
                const isActive = currentStepId === step.id;
                const isComplete = status === 'completed';
                const isFailed = status === 'failed';

                return (
                  <div
                    key={step.id}
                    className={`sc-deploy-wizard__step ${isActive ? 'active' : ''} ${isComplete ? 'complete' : ''} ${isFailed ? 'failed' : ''}`}
                  >
                    <div className={`sc-deploy-wizard__step-icon ${isActive ? 'active' : ''} ${isComplete ? 'complete' : ''} ${isFailed ? 'failed' : ''}`}>
                      {isActive && <Loader2 size={14} className="sc-community-spinner" />}
                      {isComplete && <CheckCircle2 size={14} />}
                      {isFailed && <XCircle size={14} />}
                      {!isActive && !isComplete && !isFailed && <step.icon size={14} />}
                    </div>
                    <div className="sc-deploy-wizard__step-content">
                      <p className={`sc-deploy-wizard__step-label ${isActive ? 'active' : ''} ${isComplete ? 'complete' : ''} ${isFailed ? 'failed' : ''}`}>
                        {step.label}
                      </p>
                      {isActive && stepData?.description && (
                        <p className="sc-deploy-wizard__step-desc">{stepData.description}</p>
                      )}
                      {isFailed && stepData?.error && (
                        <p className="sc-deploy-wizard__step-error">{stepData.error}</p>
                      )}
                    </div>
                    {isActive && (
                      <span className="sc-deploy-wizard__step-status">
                        <Loader2 size={10} className="sc-community-spinner" />
                        Running
                      </span>
                    )}
                    {isComplete && (
                      <StatusPill status="live" label="Done" />
                    )}
                    {isFailed && (
                      <StatusPill status="revoked" label="Failed" />
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Success Step */}
        {currentStep === 'success' && (
          <div className="sc-deploy-wizard__body">
            <div className="sc-deploy-wizard__result">
              <div className="sc-deploy-wizard__result-icon success">
                <CheckCircle2 size={32} />
              </div>
              <h3 className="sc-deploy-wizard__result-title">
                {bundle.display_name} Deployed!
              </h3>
              <p className="sc-deploy-wizard__result-desc">
                Your backend is ready to use. All functions and auth have been configured.
              </p>
            </div>

            {/* IDs */}
            {(appId || backendId) && (
              <div className="sc-deploy-wizard__ids">
                {appId && (
                  <div className="sc-deploy-wizard__id">
                    <span className="sc-deploy-wizard__id-label">App ID</span>
                    <span className="sc-deploy-wizard__id-value">{appId}</span>
                  </div>
                )}
                {backendId && (
                  <div className="sc-deploy-wizard__id">
                    <span className="sc-deploy-wizard__id-label">Backend ID</span>
                    <span className="sc-deploy-wizard__id-value">{backendId}</span>
                  </div>
                )}
              </div>
            )}

            {/* What's Ready */}
            <div className="sc-deploy-wizard__checklist">
              <p className="sc-deploy-wizard__checklist-title">What's ready:</p>
              <ul className="sc-deploy-wizard__checklist-list">
                <li><CheckCircle2 size={14} /> App with default configuration</li>
                <li><CheckCircle2 size={14} /> Backend infrastructure</li>
                <li><CheckCircle2 size={14} /> Pre-configured function templates</li>
                <li><CheckCircle2 size={14} /> Auth provider setup</li>
              </ul>
            </div>
          </div>
        )}

        {/* Error Step */}
        {currentStep === 'error' && (
          <div className="sc-deploy-wizard__body">
            <div className="sc-deploy-wizard__result">
              <div className="sc-deploy-wizard__result-icon error">
                <XCircle size={32} />
              </div>
              <h3 className="sc-deploy-wizard__result-title">
                Deployment Failed
              </h3>
              <p className="sc-deploy-wizard__result-desc">
                {errorMessage || 'An unexpected error occurred during deployment.'}
              </p>
            </div>

            <div className="sc-deploy-wizard__checklist error">
              <p className="sc-deploy-wizard__checklist-title">What you can do:</p>
              <ul className="sc-deploy-wizard__checklist-list">
                <li><ArrowRight size={14} /> Try again in a few minutes</li>
                <li><ArrowRight size={14} /> Contact support if the issue persists</li>
              </ul>
            </div>
          </div>
        )}

        {/* Footer Actions */}
        <div className="sc-deploy-wizard__footer">
          {currentStep === 'configure' && (
            <>
              <FrameButton onClick={handleClose}>Cancel</FrameButton>
              <SealedButton
                iconLeft={<Rocket size={14} />}
                loading={isDeploying}
                onClick={handleStartDeploy}
              >
                Start Deployment
              </SealedButton>
            </>
          )}

          {currentStep === 'deploying' && (
            <FrameButton disabled>
              <Loader2 size={14} className="sc-community-spinner" />
              Deploying...
            </FrameButton>
          )}

          {currentStep === 'success' && (
            <>
              <FrameButton onClick={handleDeployAnother}>Deploy Another</FrameButton>
              <FrameButton iconLeft={<Settings size={14} />} onClick={handleViewBackend}>
                View Backend
              </FrameButton>
              <FrameButton iconLeft={<Package size={14} />} onClick={() => { handleClose(); navigate(`/bundles/provisioning?bundle=${bundle.slug}`); }}>
                Provisioning
              </FrameButton>
              <SealedButton iconLeft={<Home size={14} />} onClick={handleGoToDashboard}>
                Go to Dashboard
              </SealedButton>
            </>
          )}

          {currentStep === 'error' && (
            <>
              <FrameButton onClick={handleClose}>Close</FrameButton>
              <SealedButton onClick={handleDeployAnother}>Try Again</SealedButton>
            </>
          )}
        </div>
      </div>
    </Modal>
  );
}
