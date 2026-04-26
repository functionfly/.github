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
import { useState } from 'react';
import Confetti from 'react-confetti';
import { useNavigate } from 'react-router-dom';
// import { useAuthStore } from "@/stores/authStore";
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { Stepper, StepperContent, type Step } from '@/components/ui/stepper';
import { useOnboardingStore, type OnboardingStep } from '@/stores/onboardingStore';
import { Footer } from '../LandingPage/components/Footer';
import { ConnectProviderStep } from './ConnectProviderStep';
import { DeployFunctionStep } from './DeployFunctionStep';
import { TeamSetupStep } from './TeamSetupStep';
import { TestFailoverStep } from './TestFailoverStep';
import { WelcomeStep } from './WelcomeStep';

// Define base steps
const baseSteps = [
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

// Admin-only step
const adminSteps = [
  {
    id: 'team-setup',
    title: 'Setup Your Team',
    description: 'Invite team members and configure collaboration settings',
    icon: Users,
  },
];

export function OnboardingPage() {
  const navigate = useNavigate();

  const { currentStep, completedSteps, completeStep, skipOnboarding, userRole } =
    useOnboardingStore();
  const [isCompleting, setIsCompleting] = useState(false);
  const [showSkipDialog, setShowSkipDialog] = useState(false);
  const [showConfetti, setShowConfetti] = useState(false);

  // Determine steps based on user role
  const steps = userRole === 'admin' ? [...baseSteps, ...adminSteps] : baseSteps;

  const currentStepIndex = steps.findIndex((s) => s.id === currentStep);
  const progress = ((currentStepIndex + 1) / steps.length) * 100;

  const handleNext = async () => {
    // Special handling for team-setup step - this might change the steps shown
    if (currentStep === 'team-setup') {
      completeStep('team-setup' as OnboardingStep);
      // After team setup, onboarding is complete
      setIsCompleting(true);
      setShowConfetti(true);

      // Show celebration for 3 seconds before navigating
      setTimeout(() => {
        setShowConfetti(false);
        navigate('/overview');
      }, 3000);
      return;
    }

    if (currentStepIndex < steps.length - 1) {
      completeStep(currentStep as OnboardingStep);
    } else {
      // Complete onboarding with celebration
      setIsCompleting(true);
      setShowConfetti(true);
      completeStep(currentStep as OnboardingStep);

      // Show celebration for 3 seconds before navigating
      setTimeout(() => {
        setShowConfetti(false);
        navigate('/overview');
      }, 3000);
    }
  };

  const handleSkip = () => {
    setShowSkipDialog(true);
  };

  const confirmSkip = () => {
    skipOnboarding();
    setShowSkipDialog(false);
    navigate('/overview');
  };

  const handleBack = () => {
    if (currentStepIndex > 0) {
      // Allow going back to welcome step
      const prevStep = steps[currentStepIndex - 1].id;
      useOnboardingStore.setState({ currentStep: prevStep as OnboardingStep });
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
    <div className="min-h-screen bg-bg-primary flex flex-col">
      {/* Header */}
      <header className="border-b border-border-subtle bg-bg-glass backdrop-blur-sm">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-ff-flame to-ff-afterburner flex items-center justify-center">
              <Zap className="w-5 h-5 text-white" fill="currentColor" />
            </div>
            <span className="text-xl font-bold gradient-text ff-brand-flame">FunctionFly</span>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-text-secondary">
              Step {currentStepIndex + 1} of {steps.length}
            </span>
            <HelpTooltip content="Need help? FunctionFly deploys your functions across multiple cloud providers for high availability. Each step connects a new provider and tests failover automatically.">
              <Button variant="ghost" size="sm" className="text-text-muted" aria-label="Help">
                <HelpCircle className="w-4 h-4" />
              </Button>
            </HelpTooltip>
            <Button variant="ghost" size="sm" onClick={handleSkip} className="text-text-muted hover:text-ff-flame">
              Skip for now
            </Button>
          </div>
        </div>
      </header>

      {/* Progress Bar */}
      <div className="w-full h-1 bg-bg-secondary">
        <motion.div
          className="h-full bg-gradient-to-r from-ff-flame to-ff-afterburner"
          initial={{ width: 0 }}
          animate={{ width: `${progress}%` }}
          transition={{ duration: 0.5, ease: 'easeInOut' }}
        />
      </div>

      {/* Main Content */}
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-2xl">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentStep}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.3 }}
            >
              <Card className="card">
                <CardHeader className="text-center pb-2">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-ff-flame/20 to-ff-afterburner/20 rounded-full flex items-center justify-center mb-4">
                    {(() => {
                      const step = steps[currentStepIndex];
                      if (!step?.icon) return null;
                      const Icon = step.icon;
                      return <Icon className="w-8 h-8 text-ff-flame" />;
                    })()}
                  </div>
                  {(() => {
                    const step = steps[currentStepIndex];
                    return (
                      <>
                        <CardTitle className="text-2xl text-text-primary font-display">
                          {step?.title || ''}
                        </CardTitle>
                        <CardDescription className="text-text-secondary text-base">
                          {step?.description || ''}
                        </CardDescription>
                      </>
                    );
                  })()}
                </CardHeader>

                <CardContent className="space-y-6">
                  {/* Step Content */}
                  <CurrentStepComponent />

                  {/* Navigation */}
                  <div className="flex items-center justify-between pt-4 border-t border-border-subtle">
                    <Button
                      variant="ghost"
                      onClick={handleBack}
                      disabled={currentStepIndex <= 0}
                      className="gap-2"
                    >
                      <ArrowLeft className="w-4 h-4" />
                      Back
                    </Button>

                    <Button
                      onClick={handleNext}
                      disabled={isCompleting}
                      className="btn-primary gap-2"
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

          {/* Contextual Help Panel */}
          {completedSteps.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-6"
            >
              <Card className="card p-4 bg-bg-glass backdrop-blur-sm border-ff-flame/20">
                <div className="flex items-start gap-3">
                  <Lightbulb className="w-5 h-5 text-ff-flame shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <h4 className="font-medium text-text-primary mb-2">
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
                    <div className="text-sm text-text-secondary space-y-2">
                      {currentStepIndex === 0 && (
                        <div className="space-y-2">
                          <p>
                            Take a moment to learn about FunctionFly's multi-provider deployment and
                            automatic failover features.
                          </p>
                          {userRole && (
                            <div
                              className={`p-2 rounded text-xs ${
                                userRole === 'admin'
                                  ? 'bg-ff-stratosphere/20 border border-ff-stratosphere/40 text-ff-stratosphere'
                                  : userRole === 'member'
                                    ? 'bg-ff-cyan/20 border border-ff-cyan/40 text-ff-cyan'
                                    : 'bg-ff-taxiway/20 border border-ff-taxiway/40 text-ff-taxiway'
                              }`}
                            >
                              👋 Welcome{' '}
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
                          <div className="bg-ff-taxiway border border-ff-taxiway/30 rounded p-2">
                            <p className="text-ff-pitch text-xs font-medium">
                              ✅ API token securely stored and encrypted
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
                          <div className="bg-ff-cyan/20 border border-ff-cyan/40 rounded p-2">
                            <p className="text-ff-cyan text-xs font-medium">
                              💡 FunctionFly automatically routes traffic to healthy providers if
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
                            <div className="bg-ff-cyan/20 border border-ff-cyan/40 rounded p-2">
                              <p className="text-ff-cyan text-xs font-medium">
                                👥 Invite team members to collaborate on functions and share
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
<div className="bg-ff-taxiway/20 border border-ff-taxiway/40 rounded p-2">
                            <p className="text-ff-taxiway text-xs font-medium">
                              🎉 You're all set to start deploying functions with high
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
                          <div className="bg-ff-stratosphere/10 border border-ff-stratosphere/20 rounded p-2">
                            <p className="text-ff-stratosphere text-xs">
                              🚀 Your functions are now deployed across multiple providers with
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

          {/* Step Indicators - Custom branded stepper */}
          <div className="mt-8 flex justify-center">
            <div className="onboarding-stepper">
              {steps.map((step, index) => {
                if (!step) return null;
                const isActive = index === currentStepIndex;
                const isCompleted = completedSteps.includes(step.id);
                const Icon = step.icon;
                if (!Icon) return null;

                return (
                  <div
                    key={step.id}
                    className={`onboarding-stepper-item ${isActive ? 'active' : ''} ${isCompleted ? 'completed' : ''}`}
                  >
                    <div className="onboarding-stepper-icon">
                      {isCompleted ? (
                        <CheckCircle className="w-5 h-5" />
                      ) : (
                        Icon ? <Icon className="w-5 h-5" /> : null
                      )}
                    </div>
                    <div className="onboarding-stepper-content">
                      <span className="onboarding-stepper-title">{step.title}</span>
                      <span className="onboarding-stepper-description">{step.description}</span>
                    </div>
                    {index < steps.length - 1 && (
                      <div className={`onboarding-stepper-connector ${isActive ? 'active' : ''} ${isCompleted ? 'completed' : ''}`} />
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </main>

      {/* Skip Confirmation Dialog */}
      <Dialog open={showSkipDialog} onOpenChange={setShowSkipDialog} className="onboarding-skip-dialog">
        <DialogContent className="sm:max-w-md" aria-describedby="skip-dialog-desc">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 onboarding-skip-warning-icon" />
              Skip Onboarding?
            </DialogTitle>
            <DialogDescription id="skip-dialog-desc" className="onboarding-skip-text">
              You're about to skip the remaining FunctionFly onboarding steps. Here's what that means:
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            {currentStepIndex < 1 && (
              <div className="onboarding-skip-alert-item">
                <AlertTriangle className="w-4 h-4 onboarding-skip-alert-icon onboarding-skip-alert-icon--warning" />
                <span className="onboarding-skip-text">You won't connect any cloud providers for deployment</span>
              </div>
            )}
            {currentStepIndex < 2 && (
              <div className="onboarding-skip-alert-item">
                <AlertTriangle className="w-4 h-4 onboarding-skip-alert-icon onboarding-skip-alert-icon--warning" />
                <span className="onboarding-skip-text">You won't deploy your first function to test the setup</span>
              </div>
            )}
            {currentStepIndex < 3 && (
              <div className="onboarding-skip-alert-item">
                <AlertTriangle className="w-4 h-4 onboarding-skip-alert-icon onboarding-skip-alert-icon--warning" />
                <span className="onboarding-skip-text">You won't test automatic failover capabilities</span>
              </div>
            )}
            {userRole === 'admin' && currentStepIndex < 4 && (
              <div className="onboarding-skip-alert-item">
                <Lightbulb className="w-4 h-4 onboarding-skip-alert-icon onboarding-skip-alert-icon--info" />
                <span className="onboarding-skip-text">You can still invite team members later from your dashboard</span>
              </div>
            )}
            <div className="onboarding-skip-alert-item">
              <Lightbulb className="w-4 h-4 onboarding-skip-alert-icon onboarding-skip-alert-icon--info" />
              <span className="onboarding-skip-text">You can resume onboarding anytime from your dashboard settings</span>
            </div>
            <div className="onboarding-skip-alert-item">
              <CheckCircle className="w-4 h-4 onboarding-skip-alert-icon onboarding-skip-alert-icon--success" />
              <span className="onboarding-skip-text">
                Basic functions can still be deployed, but without multi-provider benefits
              </span>
            </div>
          </div>
          <DialogFooter className="onboarding-skip-dialog-footer">
            <button onClick={() => setShowSkipDialog(false)} className="onboarding-skip-btn-continue">
              Continue Onboarding
            </button>
            <button onClick={confirmSkip} className="onboarding-skip-btn-skip">
              Skip Anyway
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Completion Celebration Overlay */}
      {isCompleting && (
        <>
          {showConfetti && (
            <Confetti
              width={window.innerWidth}
              height={window.innerHeight}
              recycle={false}
              numberOfPieces={200}
              gravity={0.3}
              colors={['#FF6B35', '#FFB800', '#00D4FF', '#5B7CF5', '#10b981', '#FF4F5E']}
            />
          )}
          <motion.div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            <motion.div
              className="text-center"
              initial={{ scale: 0.5, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ type: 'spring', bounce: 0.4 }}
            >
              {/* Celebration particles */}
              <div className="absolute inset-0 pointer-events-none">
                {[...Array(50)].map((_, i) => (
                  <motion.div
                    key={i}
                    className="absolute w-3 h-3 rounded-full"
                    style={{
                      background: `linear-gradient(45deg, ${['#FF6B35', '#FF4F5E', '#00D4FF', '#FFB800', '#5B7CF5'][Math.floor(Math.random() * 5)]}, transparent)`,
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
                className="w-24 h-24 bg-gradient-to-r from-ff-flame to-ff-afterburner rounded-full flex items-center justify-center mx-auto mb-6 shadow-2xl"
                animate={{
                  rotate: [0, 10, -10, 0],
                  scale: [1, 1.1, 1],
                }}
                transition={{
                  rotate: { duration: 0.6, repeat: Infinity, ease: 'easeInOut' },
                  scale: { duration: 1, repeat: Infinity, ease: 'easeInOut' },
                }}
              >
                <CheckCircle className="w-12 h-12 text-white" />
              </motion.div>

              <motion.h2
                className="text-3xl font-bold text-white mb-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
              >
                Welcome to FunctionFly! 🎉
              </motion.h2>

              <motion.p
                className="text-white/80 text-lg"
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

      {/* Footer */}
      <Footer />
    </div>
  );
}
