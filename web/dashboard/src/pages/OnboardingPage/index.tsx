import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { AnimatePresence, motion } from 'framer-motion';
import {
  AlertTriangle,
  ArrowLeft,
  ArrowRight,
  CheckCircle,
  Cloud,
  HelpCircle,
  Lightbulb,
  Loader2,
  Rocket,
  Shield,
  Users,
  Zap,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import Confetti from 'react-confetti';
import { useNavigate, useParams } from 'react-router-dom';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { trackEvent } from '@/lib/analytics';
import { useOnboardingStore, type OnboardingStep } from '@/stores/onboardingStore';
import { Footer } from '../LandingPage/components/Footer';
import { ConnectProviderStep } from './ConnectProviderStep';
import { DeployFunctionStep } from './DeployFunctionStep';
import { TeamSetupStep } from './TeamSetupStep';
import { TestFailoverStep } from './TestFailoverStep';
import { WelcomeStep } from './WelcomeStep';

const baseSteps: Array<{ id: OnboardingStep; title: string; description: string; icon: typeof Zap }> = [
  {
    id: 'welcome',
    title: 'Welcome to FunctionFly',
    description: 'Learn about multi-provider deployment and automatic failover',
    icon: Zap,
  },
  {
    id: 'connect-provider',
    title: 'Connect Your First Provider',
    description: 'Link a cloud provider to start deploying your functions',
    icon: Cloud,
  },
  {
    id: 'deploy-function',
    title: 'Deploy Your First Function',
    description: 'Deploy a simple function to see FunctionFly in action',
    icon: Rocket,
  },
  {
    id: 'test-failover',
    title: 'Test Failover',
    description: 'Verify your setup by testing automatic failover',
    icon: Shield,
  },
];

const adminSteps: Array<{ id: OnboardingStep; title: string; description: string; icon: typeof Users }> = [
  {
    id: 'team-setup',
    title: 'Setup Your Team',
    description: 'Invite team members and configure collaboration settings',
    icon: Users,
  },
];

export function OnboardingPage() {
  const navigate = useNavigate();
  const params = useParams();

  const { currentStep, completedSteps, completeStep, skipOnboarding, userRole, setCurrentStep } =
    useOnboardingStore();
  const [isCompleting, setIsCompleting] = useState(false);
  const [showSkipDialog, setShowSkipDialog] = useState(false);
  const [showConfetti, setShowConfetti] = useState(false);

  const steps = userRole === 'admin' ? [...baseSteps, ...adminSteps] : baseSteps;

  const currentStepFromUrl = params.step as OnboardingStep | undefined;
  const currentStepIndex = steps.findIndex((s) => s.id === (currentStepFromUrl || currentStep));
  const progress = ((currentStepIndex + 1) / steps.length) * 100;

  useEffect(() => {
    if (currentStepFromUrl && currentStepFromUrl !== currentStep) {
      setCurrentStep(currentStepFromUrl);
    }
  }, [currentStepFromUrl]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (showSkipDialog || isCompleting) return;

      if (e.key === 'ArrowRight' && currentStepIndex < steps.length - 1) {
        handleNext();
      } else if (e.key === 'ArrowLeft' && currentStepIndex > 0) {
        handleBack();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentStepIndex, steps.length, showSkipDialog, isCompleting]);

  const updateUrl = (step: OnboardingStep) => {
    const targetUrl = `/onboarding/${step}`;
    if (window.location.pathname !== targetUrl) {
      setTimeout(() => {
        navigate(targetUrl, { replace: true });
      }, 50);
    }
  };

  const handleNext = async () => {
    if (isCompleting) return;
    trackEvent('onboarding_step_completed', { step: currentStep });
    if (currentStep === 'team-setup') {
      completeStep('team-setup' as OnboardingStep);
      setIsCompleting(true);
      setShowConfetti(true);

      setTimeout(() => {
        setShowConfetti(false);
        trackEvent('onboarding_completed');
        navigate('/overview');
      }, 3000);
      return;
    }

    if (currentStepIndex < steps.length - 1) {
      const nextStep = steps[currentStepIndex + 1].id;
      setCurrentStep(nextStep);
      updateUrl(nextStep);
    } else {
      setIsCompleting(true);
      setShowConfetti(true);
      completeStep(currentStep as OnboardingStep);

      setTimeout(() => {
        setShowConfetti(false);
        trackEvent('onboarding_completed');
        navigate('/overview');
      }, 3000);
    }
  };

  const handleSkip = () => {
    trackEvent('onboarding_skip_viewed');
    setShowSkipDialog(true);
  };

  const confirmSkip = () => {
    trackEvent('onboarding_skipped');
    skipOnboarding();
    setShowSkipDialog(false);
    navigate('/overview');
  };

  const handleBack = () => {
    trackEvent('onboarding_step_back', { step: currentStep });
    if (currentStepIndex > 0) {
      const prevStep = steps[currentStepIndex - 1].id;
      setCurrentStep(prevStep);
      updateUrl(prevStep);
    }
  };

  const CurrentStepComponent = {
    welcome: WelcomeStep,
    'connect-provider': ConnectProviderStep,
    'deploy-function': DeployFunctionStep,
    'test-failover': TestFailoverStep,
    'team-setup': TeamSetupStep,
  }[currentStep];

  return (
    <div className="min-h-screen bg-aviation-bg-primary flex flex-col">
      <header className="border-b border-aviation-border-panel bg-aviation-bg-primary/95 backdrop-blur-sm">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button
              onClick={() => {
                navigate('/onboarding');
                setCurrentStep('welcome');
              }}
              className="flex items-center gap-3 hover:opacity-80 transition-opacity"
            >
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-aviation-amber to-aviation-amber-glow flex items-center justify-center shadow-lg shadow-aviation-amber-dim">
                <Zap className="w-5 h-5 text-aviation-bg-primary" fill="currentColor" />
              </div>
              <span className="text-xl font-bold font-mono text-aviation-text-primary">
                Function<span className="text-aviation-amber">Fly</span>
              </span>
            </button>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm font-mono text-aviation-text-muted">
              Step {currentStepIndex + 1} of {steps.length}
            </span>
            <HelpTooltip content="Need help? FunctionFly deploys your functions across multiple cloud providers for high availability. Each step connects a new provider and tests failover automatically.">
              <Button variant="ghost" size="sm" className="text-aviation-text-muted hover:text-aviation-amber hover:bg-aviation-bg-instrument" aria-label="Help">
                <HelpCircle className="w-4 h-4" />
              </Button>
            </HelpTooltip>
            <Button variant="ghost" size="sm" onClick={handleSkip} className="text-aviation-text-muted hover:text-aviation-amber font-mono text-sm">
              Skip for now
            </Button>
          </div>
        </div>
      </header>

      <div className="w-full h-1 bg-aviation-bg-secondary">
        <motion.div
          className="h-full bg-gradient-to-r from-aviation-amber to-aviation-cyan"
          initial={{ width: 0 }}
          animate={{ width: `${progress}%` }}
          transition={{ duration: 0.5, ease: 'easeInOut' }}
        />
      </div>

      <main className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-2xl">
          <div
            role="status"
            aria-live="polite"
            aria-atomic="true"
            className="sr-only"
          >
            {`Step ${currentStepIndex + 1} of ${steps.length}: ${steps[currentStepIndex]?.title || ''}`}
          </div>

          <AnimatePresence mode="wait">
            <motion.div
              key={currentStep}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.3 }}
            >
              <Card className="aviation-panel">
                <CardHeader className="text-center pb-2">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-aviation-amber/20 to-aviation-cyan/20 rounded-2xl flex items-center justify-center mb-4 border border-aviation-amber-dim">
                    {(() => {
                      const step = steps[currentStepIndex];
                      if (!step?.icon) return null;
                      const Icon = step.icon;
                      return <Icon className="w-8 h-8 text-aviation-amber" />;
                    })()}
                  </div>
                  {(() => {
                    const step = steps[currentStepIndex];
                    return (
                      <>
                        <CardTitle className="text-2xl text-aviation-text-primary font-mono font-bold">
                          {step?.title || ''}
                        </CardTitle>
                        <CardDescription className="text-aviation-text-secondary text-base font-mono">
                          {step?.description || ''}
                        </CardDescription>
                      </>
                    );
                  })()}
                </CardHeader>

                <CardContent className="space-y-6">
                  <CurrentStepComponent />

                  <div className="flex items-center justify-between pt-4 border-t border-aviation-border-panel">
                    <Button
                      variant="ghost"
                      onClick={handleBack}
                      disabled={currentStepIndex <= 0}
                      className="gap-2 font-mono text-aviation-text-secondary hover:text-aviation-amber hover:bg-aviation-bg-instrument"
                    >
                      <ArrowLeft className="w-4 h-4" />
                      Back
                    </Button>

                    <Button
                      onClick={handleNext}
                      disabled={isCompleting}
                      className="aviation-button-primary gap-2 font-mono"
                    >
                      {isCompleting ? (
                        <>
                          <Loader2 className="w-4 h-4 animate-spin" />
                          Completing...
                        </>
                      ) : currentStepIndex === steps.length - 1 ? (
                        <>
                          <CheckCircle className="w-4 h-4" />
                          Complete Setup
                        </>
                      ) : (
                        <>
                          Next Step
                          <ArrowRight className="w-4 h-4" />
                        </>
                      )}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          </AnimatePresence>

          {completedSteps.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-6"
            >
              <Card className="aviation-panel-glow p-4">
                <div className="flex items-start gap-3">
                  <Lightbulb className="w-5 h-5 text-aviation-amber shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <h4 className="font-mono font-semibold text-aviation-text-primary mb-2">
                      {currentStepIndex === 0 && 'Welcome to FunctionFly'}
                      {completedSteps.includes('welcome') &&
                        currentStepIndex === 1 &&
                        'Ready to Connect'}
                      {completedSteps.includes('connect-provider') &&
                        currentStepIndex === 2 &&
                        'Great! Provider Connected'}
                      {completedSteps.includes('deploy-function') &&
                        currentStepIndex === 3 &&
                        'Function Deployed Successfully'}
                      {completedSteps.length === 4 && 'Setup Complete!'}
                    </h4>
                    <div className="text-sm text-aviation-text-secondary space-y-2 font-mono">
                      {currentStepIndex === 0 && (
                        <div className="space-y-2">
                          <p>
                            Take a moment to learn about FunctionFly's multi-provider deployment and
                            automatic failover features.
                          </p>
                          {userRole && (
                            <div
                              className={`p-2 rounded text-xs font-mono ${
                                userRole === 'admin'
                                  ? 'bg-aviation-stratosphere/20 border border-aviation-stratosphere/40 text-aviation-stratosphere'
                                  : userRole === 'member'
                                    ? 'bg-aviation-cyan-dim border border-aviation-cyan/40 text-aviation-cyan'
                                    : 'bg-aviation-green-dim border border-aviation-green/40 text-aviation-green'
                              }`}
                            >
                              Welcome{' '}
                              {userRole === 'admin'
                                ? 'Team Administrator'
                                : userRole === 'member'
                                  ? 'Team Member'
                                  : 'Viewer'}
                              !
                              {userRole === 'admin' &&
                                ' You can manage providers, deploy functions, and invite team members.'}
                              {userRole === 'member' &&
                                ' You can deploy functions and access shared resources.'}
                              {userRole === 'viewer' &&
                                ' You have read-only access to team functions and metrics.'}
                            </div>
                          )}
                        </div>
                      )}
                      {completedSteps.includes('welcome') && currentStepIndex === 1 && (
                        <p>
                          You're about to connect your first cloud provider. This allows FunctionFly
                          to deploy functions across multiple providers for high availability.
                        </p>
                      )}
                      {completedSteps.includes('connect-provider') && currentStepIndex === 2 && (
                        <div className="space-y-2">
                          <p>
                            Excellent! Your provider is connected. Now let's deploy your first
                            function to see FunctionFly in action.
                          </p>
                          <div className="bg-aviation-green-dim border border-aviation-green/30 rounded p-2">
                            <p className="text-aviation-green text-xs font-mono font-medium">
                              API token securely stored and encrypted
                            </p>
                          </div>
                        </div>
                      )}
                      {completedSteps.includes('deploy-function') && currentStepIndex === 3 && (
                        <div className="space-y-2">
                          <p>
                            Your function is live! The final step tests your failover setup to
                            ensure high availability.
                          </p>
                          <div className="bg-aviation-cyan-dim border border-aviation-cyan/40 rounded p-2">
                            <p className="text-aviation-cyan text-xs font-mono font-medium">
                              FunctionFly automatically routes traffic to healthy providers if
                              one fails
                            </p>
                          </div>
                        </div>
                      )}
                      {completedSteps.includes('test-failover') &&
                        userRole === 'admin' &&
                        currentStepIndex === 4 && (
                          <div className="space-y-2">
                            <p>
                              Great! Your failover setup is working. Now let's set up your team for
                              collaboration.
                            </p>
                            <div className="bg-aviation-cyan-dim border border-aviation-cyan/40 rounded p-2">
                              <p className="text-aviation-cyan text-xs font-mono font-medium">
                                Invite team members to collaborate on functions and share
                                provider access
                              </p>
                            </div>
                          </div>
                        )}
                      {completedSteps.includes('test-failover') &&
                        userRole !== 'admin' &&
                        currentStepIndex === 3 && (
                          <div className="space-y-2">
                            <p>
                              Excellent! Your FunctionFly setup is complete and ready for
                              production.
                            </p>
                            <div className="bg-aviation-green-dim border border-aviation-green/40 rounded p-2">
                              <p className="text-aviation-green text-xs font-mono font-medium">
                                You're all set to start deploying functions with high
                                availability!
                              </p>
                            </div>
                          </div>
                        )}
                      {completedSteps.length === steps.length && (
                        <div className="space-y-2">
                          <p>
                            Congratulations! Your FunctionFly setup is complete and
                            production-ready.
                          </p>
                          <div className="bg-aviation-stratosphere/10 border border-aviation-stratosphere/20 rounded p-2">
                            <p className="text-aviation-stratosphere text-xs font-mono">
                              Your functions are now deployed across multiple providers with
                              automatic failover
                            </p>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </Card>
            </motion.div>
          )}

          <div className="mt-8 flex justify-center">
            <div className="flex items-center gap-4">
              {steps.map((step, index) => {
                if (!step) return null;
                const isActive = index === currentStepIndex;
                const isCompleted = completedSteps.includes(step.id);
                const Icon = step.icon;
                if (!Icon) return null;

                return (
                  <div key={step.id} className="flex items-center">
                    <div
                      role="listitem"
                      aria-current={isActive ? 'step' : undefined}
                      className={`flex flex-col items-center gap-2 p-4 rounded-xl transition-all duration-300 ${
                        isActive
                          ? 'bg-aviation-bg-instrument border-2 border-aviation-amber shadow-lg shadow-aviation-amber-dim'
                          : isCompleted
                            ? 'bg-aviation-bg-instrument border border-aviation-green/50'
                            : 'bg-aviation-bg-secondary border border-aviation-border-panel'
                      }`}
                    >
                      <div
                        className={`w-12 h-12 rounded-full flex items-center justify-center font-mono font-bold text-lg transition-all duration-300 ${
                          isActive
                            ? 'bg-gradient-to-br from-aviation-amber to-aviation-amber-glow text-aviation-bg-primary shadow-lg shadow-aviation-amber-glow scale-110'
                            : isCompleted
                              ? 'bg-aviation-green text-aviation-bg-primary'
                              : 'bg-aviation-bg-tertiary text-aviation-text-muted'
                        }`}
                      >
                        {isCompleted ? (
                          <CheckCircle className="w-6 h-6" />
                        ) : (
                          Icon && <Icon className="w-6 h-6" />
                        )}
                      </div>
                      <div className="flex flex-col items-center gap-1 text-center">
                        <span
                          className={`font-mono text-sm font-semibold ${
                            isActive
                              ? 'text-aviation-amber'
                              : isCompleted
                                ? 'text-aviation-green'
                                : 'text-aviation-text-muted'
                          }`}
                        >
                          {step.title}
                        </span>
                      </div>
                    </div>
                    {index < steps.length - 1 && (
                      <div
                        className={`w-8 h-1 mx-2 rounded-full transition-all duration-300 ${
                          isCompleted
                            ? 'bg-aviation-green'
                            : isActive
                              ? 'bg-gradient-to-r from-aviation-amber to-aviation-cyan'
                              : 'bg-aviation-border-panel'
                        }`}
                      />
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </main>

      <Dialog open={showSkipDialog} onOpenChange={setShowSkipDialog}>
        <DialogContent className="aviation-panel sm:max-w-md" aria-describedby="skip-dialog-desc">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 font-mono text-aviation-text-primary">
              <AlertTriangle className="w-5 h-5 text-aviation-amber" />
              Skip Onboarding?
            </DialogTitle>
            <DialogDescription id="skip-dialog-desc" className="text-aviation-text-secondary font-mono">
              You're about to skip the remaining FunctionFly onboarding steps. Here's what that means:
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            {currentStepIndex < 1 && (
              <div className="flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-aviation-amber shrink-0 mt-0.5" />
                <span className="text-aviation-text-secondary font-mono text-sm">You won't connect any cloud providers for deployment</span>
              </div>
            )}
            {currentStepIndex < 2 && (
              <div className="flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-aviation-amber shrink-0 mt-0.5" />
                <span className="text-aviation-text-secondary font-mono text-sm">You won't deploy your first function to test the setup</span>
              </div>
            )}
            {currentStepIndex < 3 && (
              <div className="flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-aviation-amber shrink-0 mt-0.5" />
                <span className="text-aviation-text-secondary font-mono text-sm">You won't test automatic failover capabilities</span>
              </div>
            )}
            {userRole === 'admin' && currentStepIndex < 4 && (
              <div className="flex items-start gap-2">
                <Lightbulb className="w-4 h-4 text-aviation-cyan shrink-0 mt-0.5" />
                <span className="text-aviation-text-secondary font-mono text-sm">You can still invite team members later from your dashboard</span>
              </div>
            )}
            <div className="flex items-start gap-2">
              <Lightbulb className="w-4 h-4 text-aviation-cyan shrink-0 mt-0.5" />
              <span className="text-aviation-text-secondary font-mono text-sm">You can resume onboarding anytime from your dashboard settings</span>
            </div>
            <div className="flex items-start gap-2">
              <CheckCircle className="w-4 h-4 text-aviation-green shrink-0 mt-0.5" />
              <span className="text-aviation-text-secondary font-mono text-sm">
                Basic functions can still be deployed, but without multi-provider benefits
              </span>
            </div>
          </div>
          <DialogFooter className="flex gap-3 pt-4">
            <button
              onClick={() => setShowSkipDialog(false)}
              className="flex-1 px-4 py-3 font-mono font-semibold rounded-lg border border-aviation-border-instrument bg-aviation-bg-instrument text-aviation-text-primary transition-all hover:border-aviation-amber hover:text-aviation-amber"
              aria-label="Continue onboarding"
            >
              Continue Onboarding
            </button>
            <button
              onClick={confirmSkip}
              className="flex-1 px-4 py-3 font-mono font-semibold rounded-lg bg-gradient-to-r from-aviation-amber to-aviation-amber-glow text-aviation-bg-primary transition-all hover:shadow-lg hover:shadow-aviation-amber-glow"
              aria-label="Skip onboarding"
            >
              Skip Anyway
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {isCompleting && (
        <>
          {showConfetti && (
            <Confetti
              width={window.innerWidth}
              height={window.innerHeight}
              recycle={false}
              numberOfPieces={200}
              gravity={0.3}
              colors={['#f59e0b', '#ffb800', '#06b6d4', '#5b7cf5', '#10b981', '#ff4f5e']}
            />
          )}
          <motion.div
            className="fixed inset-0 bg-aviation-bg-primary/80 backdrop-blur-sm z-50 flex items-center justify-center"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            <motion.div
              className="text-center relative"
              initial={{ scale: 0.5, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ type: 'spring', bounce: 0.4 }}
            >
              <div className="absolute inset-0 pointer-events-none">
                {[...Array(20)].map((_, i) => (
                  <motion.div
                    key={i}
                    className="absolute w-3 h-3 rounded-full"
                    style={{
                      background: ['#f59e0b', '#06b6d4', '#10b981', '#5b7cf5', '#ff4f5e'][i % 5],
                      left: `${Math.random() * 100}%`,
                      top: `${Math.random() * 100}%`,
                    }}
                    initial={{
                      x: 0,
                      y: 0,
                      scale: 0,
                      opacity: 1,
                    }}
                    animate={{
                      x: (Math.random() - 0.5) * 400,
                      y: (Math.random() - 0.5) * 400,
                      scale: [0, 1, 0],
                      opacity: [1, 1, 0],
                    }}
                    transition={{
                      duration: 2,
                      delay: Math.random() * 0.5,
                      ease: 'easeOut',
                    }}
                  />
                ))}
              </div>

              <motion.div
                className="w-24 h-24 bg-gradient-to-r from-aviation-amber to-aviation-amber-glow rounded-full flex items-center justify-center mx-auto mb-6 shadow-2xl shadow-aviation-amber-glow"
                animate={{
                  rotate: [0, 10, -10, 0],
                  scale: [1, 1.1, 1],
                }}
                transition={{
                  rotate: { duration: 0.6, repeat: Infinity, ease: 'easeInOut' },
                  scale: { duration: 1, repeat: Infinity, ease: 'easeInOut' },
                }}
              >
                <CheckCircle className="w-12 h-12 text-aviation-bg-primary" />
              </motion.div>

              <motion.h2
                className="text-3xl font-bold text-aviation-text-primary mb-4 font-mono"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
              >
                Welcome to FunctionFly!
              </motion.h2>

              <motion.p
                className="text-aviation-text-secondary text-lg font-mono"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.5 }}
              >
                Your setup is complete and production-ready
              </motion.p>
            </motion.div>
          </motion.div>
        </>
      )}

      <Footer />
    </div>
  );
}