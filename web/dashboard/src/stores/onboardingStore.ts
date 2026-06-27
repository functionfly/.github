import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type OnboardingStep =
  | 'welcome'
  | 'connect-provider'
  | 'deploy-function'
  | 'test-failover'
  | 'team-setup';

export type PlanTier = 'free' | 'starter' | 'professional' | 'enterprise' | 'agent_enterprise';

export type BillingCycle = 'monthly' | 'yearly';

/**
 * Environment can be either a string ID (for simple cases like 'development', 'production')
 * or a full object with additional metadata.
 */
export type Environment = string;

export interface FullEnvironment {
  id: string;
  name: string;
  apiKeyId?: string;
  apiKeyName?: string;
}

export interface IntegrationConfig {
  id: string;
  type: string;
  name: string;
  connected: boolean;
  configuredAt?: string;
  connectedAt?: string;
}

interface StepData {
  [key: string]: any; // Allow flexible step-specific data storage
}

interface CostEstimate {
  monthlyCost: number;
  currency: string;
  breakdown: Record<string, number>;
  providerData?: Record<string, any>;
}

interface TeamInvite {
  email: string;
  token: string;
  role: string;
  expires: number;
}

interface ProviderConfig {
  id: string;
  provider: string;
  providerName: string;
  connectedAt: string;
  isShared?: boolean;
  teamId?: string;
}

interface OnboardingState {
  currentStep: OnboardingStep;
  completedSteps: OnboardingStep[];
  stepData: StepData; // Store step-specific data like function names, URLs, etc.
  isOnboardingComplete: boolean;
  hasSkippedOnboarding: boolean;
  lastUpdated: number; // Timestamp for auto-save tracking

  // New onboarding features
  userRole?: 'admin' | 'member' | 'viewer';
  teamInvites?: TeamInvite[];
  costEstimates?: Record<string, CostEstimate>;
  providers?: ProviderConfig[];

  // API Key step
  createdApiKey?: string;
  apiKeyName?: string;
  apiKeyId?: string;
  setCreatedApiKey: (key: string, name: string, id: string) => void;

  // Region selection
  selectedRegions: string[];
  autoSelectRegions: boolean;
  setSelectedRegions: (regions: string[], autoSelect?: boolean) => void;

  // Custom domain step
  customDomain?: string;
  domainVerified?: boolean;
  setCustomDomain: (domain: string, verified: boolean) => void;

  // Environment setup
  environments: Environment[];
  activeEnvironment?: string;
  customEnvironments: Environment[];
  setEnvironments: (envs: Environment[], activeEnv?: string) => void;
  addCustomEnvironment: (env: Environment) => void;

  // Plan selection
  selectedPlan?: PlanTier;
  setSelectedPlan: (plan: PlanTier, billingCycle?: BillingCycle) => void;
  billingCycle: BillingCycle;

  // Integrations
  connectedIntegrations: IntegrationConfig[];
  addConnectedIntegration: (integration: IntegrationConfig) => void;
  removeConnectedIntegration: (id: string) => void;

  // Actions
  setCurrentStep: (step: OnboardingStep) => void;
  completeStep: (step: OnboardingStep) => void;
  updateStepData: (step: OnboardingStep, data: any) => void;
  skipOnboarding: () => void;
  resetOnboarding: () => void;
  goToNextStep: () => void;
  goToPrevStep: () => void;
  canResume: () => boolean;

  // New actions
  setUserRole: (role: 'admin' | 'member' | 'viewer') => void;
  setTeamInvites: (invites: TeamInvite[]) => void;
  setCostEstimate: (provider: string, estimate: CostEstimate) => void;
  addProvider: (provider: ProviderConfig) => void;
  updateProvider: (providerId: string, updates: Partial<ProviderConfig>) => void;
}

const steps: OnboardingStep[] = [
  'welcome',
  'connect-provider',
  'deploy-function',
  'test-failover',
  'team-setup',
];

export const useOnboardingStore = create<OnboardingState>()(
  persist(
    (set, get) => ({
      currentStep: 'welcome',
      completedSteps: [],
      stepData: {},
      isOnboardingComplete: false,
      hasSkippedOnboarding: false,
      lastUpdated: Date.now(),
      userRole: undefined,
      teamInvites: [],
      costEstimates: {},
      providers: [],

      // API Key step
      createdApiKey: undefined,
      apiKeyName: undefined,
      apiKeyId: undefined,
      setCreatedApiKey: (key, name, id) =>
        set({ createdApiKey: key, apiKeyName: name, apiKeyId: id, lastUpdated: Date.now() }),

      // Region selection
      selectedRegions: [],
      autoSelectRegions: false,
      setSelectedRegions: (regions, autoSelect) =>
        set({
          selectedRegions: regions,
          autoSelectRegions: autoSelect ?? false,
          lastUpdated: Date.now(),
        }),

      // Custom domain step
      customDomain: undefined,
      domainVerified: undefined,
      setCustomDomain: (domain, verified) =>
        set({ customDomain: domain, domainVerified: verified, lastUpdated: Date.now() }),

      // Environment setup
      environments: [],
      activeEnvironment: undefined,
      customEnvironments: [],
      setEnvironments: (envs, activeEnv) =>
        set({
          environments: envs,
          activeEnvironment: activeEnv ?? envs[0],
          lastUpdated: Date.now(),
        }),
      addCustomEnvironment: (env) =>
        set({ customEnvironments: [...get().customEnvironments, env], lastUpdated: Date.now() }),

      // Plan selection
      selectedPlan: undefined,
      setSelectedPlan: (plan, billingCycle) =>
        set({
          selectedPlan: plan,
          billingCycle: billingCycle ?? get().billingCycle,
          lastUpdated: Date.now(),
        }),
      billingCycle: 'monthly',

      // Integrations
      connectedIntegrations: [],
      addConnectedIntegration: (integration) =>
        set({
          connectedIntegrations: [...get().connectedIntegrations, integration],
          lastUpdated: Date.now(),
        }),
      removeConnectedIntegration: (id) =>
        set({
          connectedIntegrations: get().connectedIntegrations.filter((i) => i.id !== id),
          lastUpdated: Date.now(),
        }),

      setCurrentStep: (step) => set({ currentStep: step, lastUpdated: Date.now() }),

      completeStep: (step) => {
        const { completedSteps } = get();
        if (!completedSteps.includes(step)) {
          set({
            completedSteps: [...completedSteps, step],
            lastUpdated: Date.now(),
          });
        }

        // Auto-advance to next step
        const currentIndex = steps.indexOf(step);
        if (currentIndex < steps.length - 1) {
          set({ currentStep: steps[currentIndex + 1], lastUpdated: Date.now() });
        } else {
          // All steps completed
          set({
            isOnboardingComplete: true,
            lastUpdated: Date.now(),
          });
        }
      },

      updateStepData: (step, data) => {
        const { stepData } = get();
        set({
          stepData: { ...stepData, [step]: { ...stepData[step], ...data } },
          lastUpdated: Date.now(),
        });
      },

      skipOnboarding: () => {
        set({
          hasSkippedOnboarding: true,
          lastUpdated: Date.now(),
        });
      },

      resetOnboarding: () => {
        set({
          currentStep: 'welcome',
          completedSteps: [],
          stepData: {},
          isOnboardingComplete: false,
          hasSkippedOnboarding: false,
          lastUpdated: Date.now(),
        });
      },

      goToNextStep: () => {
        const { currentStep, completeStep } = get();
        completeStep(currentStep);
      },

      goToPrevStep: () => {
        const { currentStep } = get();
        const currentIndex = steps.indexOf(currentStep);
        if (currentIndex > 0) {
          set({ currentStep: steps[currentIndex - 1], lastUpdated: Date.now() });
        }
      },

      canResume: () => {
        const { completedSteps, isOnboardingComplete } = get();
        return completedSteps.length > 0 && !isOnboardingComplete;
      },

      // New actions for enhanced onboarding
      setUserRole: (role) => set({ userRole: role, lastUpdated: Date.now() }),

      setTeamInvites: (invites) => set({ teamInvites: invites, lastUpdated: Date.now() }),

      setCostEstimate: (provider, estimate) => {
        const { costEstimates } = get();
        set({
          costEstimates: { ...costEstimates, [provider]: estimate },
          lastUpdated: Date.now(),
        });
      },

      addProvider: (provider) => {
        const { providers } = get();
        set({
          providers: [...(providers || []), provider],
          lastUpdated: Date.now(),
        });
      },

      updateProvider: (providerId, updates) => {
        const { providers } = get();
        if (!providers) return;

        const updatedProviders = providers.map((p) =>
          p.id === providerId ? { ...p, ...updates } : p
        );
        set({ providers: updatedProviders, lastUpdated: Date.now() });
      },
    }),
    {
      name: 'onboarding-storage',
      version: 1,
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.lastUpdated = Date.now();
        }
      },
      partialize: (state) => ({
        currentStep: state.currentStep,
        completedSteps: state.completedSteps,
        stepData: state.stepData,
        isOnboardingComplete: state.isOnboardingComplete,
        hasSkippedOnboarding: state.hasSkippedOnboarding,
        lastUpdated: state.lastUpdated,
        userRole: state.userRole,
        teamInvites: state.teamInvites,
        costEstimates: state.costEstimates,
        providers: state.providers,
      }),
    }
  )
);
