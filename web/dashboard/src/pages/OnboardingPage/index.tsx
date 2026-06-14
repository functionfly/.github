import '@/styles/aviation-dashboard.css';
import '@/styles/onboarding.css';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
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
import { useEffect, useState, useCallback } from 'react';
import Confetti from 'react-confetti';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Helmet } from 'react-helmet-async';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { trackEvent, trackPageView } from '@/lib/analytics';
import { toast } from 'sonner';
import { useOnboardingStore, type OnboardingStep } from '@/stores/onboardingStore';
import { Footer } from '../LandingPage/components/Footer';
import { ConnectProviderStep } from './ConnectProviderStep';
import { DeployFunctionStep } from './DeployFunctionStep';
import { TeamSetupStep } from './TeamSetupStep';
import { TestFailoverStep } from './TestFailoverStep';
import { WelcomeStep } from './WelcomeStep';

const baseSteps: Array<{ id: OnboardingStep; titleKey: string; descriptionKey: string; icon: typeof Zap }> = [
  {
    id: 'welcome',
    titleKey: 'onboarding.steps.welcome.title',
    descriptionKey: 'onboarding.steps.welcome.description',
    icon: Zap,
  },
  {
    id: 'connect-provider',
    titleKey: 'onboarding.steps.connectProvider.title',
    descriptionKey: 'onboarding.steps.connectProvider.description',
    icon: Cloud,
  },
  {
    id: 'deploy-function',
    titleKey: 'onboarding.steps.deployFunction.title',
    descriptionKey: 'onboarding.steps.deployFunction.description',
    icon: Rocket,
  },
  {
    id: 'test-failover',
    titleKey: 'onboarding.steps.testFailover.title',
    descriptionKey: 'onboarding.steps.testFailover.description',
    icon: Shield,
  },
];

const adminSteps: Array<{ id: OnboardingStep; titleKey: string; descriptionKey: string; icon: typeof Users }> = [
  {
    id: 'team-setup',
    titleKey: 'onboarding.steps.teamSetup.title',
    descriptionKey: 'onboarding.steps.teamSetup.description',
    icon: Users,
  },
];

const ONBOARDING_STEPS_ORDER: OnboardingStep[] = ['welcome', 'connect-provider', 'deploy-function', 'test-failover', 'team-setup'];

export function OnboardingPage() {
  const navigate = useNavigate();
  const params = useParams();
  const { t } = useTranslation('onboarding');

  const { currentStep, completedSteps, completeStep, skipOnboarding, userRole, setCurrentStep } =
    useOnboardingStore();
  const [isCompleting, setIsCompleting] = useState(false);
  const [showConfetti, setShowConfetti] = useState(false);

  const steps = userRole === 'admin' ? [...baseSteps, ...adminSteps] : baseSteps;

  const currentStepFromUrl = params.step as OnboardingStep | undefined;
  const currentStepIndex = steps.findIndex((s) => s.id === (currentStepFromUrl || currentStep));
  const progress = ((currentStepIndex + 1) / steps.length) * 100;

  useEffect(() => {
    trackPageView('/onboarding');
    trackEvent('onboarding_page_viewed', { step: currentStep });
  }, []);

  useEffect(() => {
    if (currentStepFromUrl && currentStepFromUrl !== currentStep) {
      setCurrentStep(currentStepFromUrl);
    }
  }, [currentStepFromUrl]);

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (isCompleting) return;

    if (e.key === 'ArrowRight' && currentStepIndex < steps.length - 1) {
      handleNext();
    } else if (e.key === 'ArrowLeft' && currentStepIndex > 0) {
      handleBack();
    }
  }, [currentStepIndex, steps.length, isCompleting]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

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
    trackEvent('onboarding_step_completed', { step: currentStep, step_index: currentStepIndex });
    if (currentStep === 'team-setup') {
      completeStep('team-setup' as OnboardingStep);
      setIsCompleting(true);
      setShowConfetti(true);

      setTimeout(() => {
        setShowConfetti(false);
        trackEvent('onboarding_completed', { total_steps: steps.length });
        navigate('/overview');
      }, 3000);
      return;
    }

    if (currentStepIndex < steps.length - 1) {
      const nextStep = steps[currentStepIndex + 1].id;
      setCurrentStep(nextStep);
      updateUrl(nextStep);
      trackEvent('onboarding_step_started', { step: nextStep, step_index: currentStepIndex + 1 });
    } else {
      setIsCompleting(true);
      setShowConfetti(true);
      completeStep(currentStep as OnboardingStep);

      setTimeout(() => {
        setShowConfetti(false);
        trackEvent('onboarding_completed', { total_steps: steps.length });
        navigate('/overview');
      }, 3000);
    }
  };

  const handleSkip = () => {
    confirmSkip();
  };

  const confirmSkip = () => {
    trackEvent('onboarding_skipped', { completed_steps: completedSteps.length, skipped_at: currentStep });
    skipOnboarding();
    navigate('/overview', { replace: true });
    toast.info(t('onboarding.skipDialog.hints.canResume'));
  };

  const handleBack = () => {
    trackEvent('onboarding_step_back', { step: currentStep, step_index: currentStepIndex });
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

  const stepHints: Record<OnboardingStep, string> = {
    'welcome': t('onboarding.hints.welcome'),
    'connect-provider': t('onboarding.hints.connectProvider'),
    'deploy-function': t('onboarding.hints.deployFunction'),
    'test-failover': t('onboarding.hints.testFailover'),
    'team-setup': t('onboarding.hints.teamSetup'),
  };

  return (
    <div className="min-h-screen mesh-gradient-bg flex flex-col relative overflow-hidden">
      <Helmet>
        <title>{t('onboarding.seo.title')}</title>
        <meta name="description" content={t('onboarding.seo.description')} />
        <meta name="robots" content="noindex, nofollow" />
        <link rel="canonical" href="/onboarding" />
        <meta property="og:title" content={t('onboarding.seo.title')} />
        <meta property="og:description" content={t('onboarding.seo.description')} />
        <meta property="og:type" content="website" />
        <meta property="og:url" content="/onboarding" />
        <meta name="twitter:card" content="summary" />
        <meta name="twitter:title" content={t('onboarding.seo.title')} />
        <meta name="twitter:description" content={t('onboarding.seo.description')} />
      </Helmet>

      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-aviation-amber focus:text-aviation-bg-primary focus:rounded-lg focus:font-mono"
      >
        {t('onboarding.accessibility.skipToMain')}
      </a>

      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-500/10 rounded-full blur-[128px] animate-float" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[128px] animate-float-rotate" />
        <div className="spotlight-container">
          <div className="spotlight-bg animate-spotlight" />
        </div>
      </div>

      <header className="relative border-b border-aviation-border-panel bg-aviation-bg-primary/80 backdrop-blur-sm z-10">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button
              onClick={() => {
                navigate('/onboarding');
                setCurrentStep('welcome');
              }}
              className="flex items-center gap-3 hover:opacity-80 transition-opacity focus:outline-none focus:ring-2 focus:ring-aviation-amber focus:ring-offset-2 focus:ring-offset-aviation-bg-primary rounded-lg"
              aria-label={t('onboarding.accessibility.goHome')}
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
            <span className="text-sm font-mono text-aviation-text-muted" aria-live="polite">
              {t('onboarding.stepCounter', { current: currentStepIndex + 1, total: steps.length })}
            </span>
            <HelpTooltip content={t('onboarding.helpTooltip')}>
              <Button variant="ghost" size="sm" className="text-aviation-text-muted hover:text-aviation-amber hover:bg-aviation-bg-instrument" aria-label={t('onboarding.accessibility.help')}>
                <HelpCircle className="w-4 h-4" />
              </Button>
            </HelpTooltip>
            <Button variant="ghost" size="sm" onClick={handleSkip} className="text-aviation-text-muted hover:text-aviation-amber font-mono text-sm" aria-label={t('onboarding.accessibility.skip')}>
              {t('onboarding.skip')}
            </Button>
          </div>
        </div>
      </header>

      <div className="w-full h-1 bg-aviation-bg-secondary" role="progressbar" aria-valuenow={Math.round(progress)} aria-valuemin={0} aria-valuemax={100} aria-label={t('onboarding.accessibility.progressBar', { percent: Math.round(progress) })}>
        <motion.div
          className="h-full bg-gradient-to-r from-aviation-amber to-aviation-cyan"
          initial={{ width: 0 }}
          animate={{ width: `${progress}%` }}
          transition={{ duration: 0.5, ease: 'easeInOut' }}
        />
      </div>

      <main id="main-content" className="aviation-dashboard dashboard-main-bg flex-1 flex items-center justify-center p-4 relative z-10">
        <div className="w-full max-w-2xl">
          <div
            role="status"
            aria-live="polite"
            aria-atomic="true"
            className="sr-only"
          >
            {t('onboarding.accessibility.currentStep', { current: currentStepIndex + 1, total: steps.length, title: t(steps[currentStepIndex]?.titleKey) })}
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
                      return <Icon className="w-8 h-8 text-aviation-amber" aria-hidden="true" />;
                    })()}
                  </div>
                  {(() => {
                    const step = steps[currentStepIndex];
                    return (
                      <>
                        <CardTitle className="text-2xl text-aviation-text-primary font-mono font-bold">
                          {step ? t(step.titleKey) : ''}
                        </CardTitle>
                        <CardDescription className="text-aviation-text-secondary text-base font-mono">
                          {step ? t(step.descriptionKey) : ''}
                        </CardDescription>
                      </>
                    );
                  })()}
                </CardHeader>

                <CardContent className="space-y-6">
                  <ErrorBoundary
                    fallback={
                      <div className="text-center py-8">
                        <AlertTriangle className="w-12 h-12 text-aviation-red mx-auto mb-4" />
                        <p className="text-aviation-text-primary font-mono mb-4">{t('onboarding.errors.stepLoadFailed')}</p>
                        <Button onClick={() => window.location.reload()} variant="outline" className="font-mono">
                          {t('onboarding.actions.retry')}
                        </Button>
                      </div>
                    }
                  >
                    <CurrentStepComponent />
                  </ErrorBoundary>

                  <div className="flex items-center justify-between pt-4 border-t border-aviation-border-panel">
                    <Button
                      variant="ghost"
                      onClick={handleBack}
                      disabled={currentStepIndex <= 0}
                      className="gap-2 font-mono text-aviation-text-secondary hover:text-aviation-amber hover:bg-aviation-bg-instrument"
                      aria-label={t('onboarding.accessibility.backButton')}
                    >
                      <ArrowLeft className="w-4 h-4" aria-hidden="true" />
                      {t('onboarding.actions.back')}
                    </Button>

                    <Button
                      onClick={handleNext}
                      disabled={isCompleting}
                      className="aviation-button-primary gap-2 font-mono"
                      aria-label={currentStepIndex === steps.length - 1 ? t('onboarding.accessibility.completeButton') : t('onboarding.accessibility.nextButton')}
                    >
                      {isCompleting ? (
                        <>
                          <Loader2 className="w-4 h-4 animate-spin" aria-hidden="true" />
                          {t('onboarding.actions.completing')}
                        </>
                      ) : currentStepIndex === steps.length - 1 ? (
                        <>
                          <CheckCircle className="w-4 h-4" aria-hidden="true" />
                          {t('onboarding.actions.completeSetup')}
                        </>
                      ) : (
                        <>
                          {t('onboarding.actions.nextStep')}
                          <ArrowRight className="w-4 h-4" aria-hidden="true" />
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
                  <Lightbulb className="w-5 h-5 text-aviation-amber shrink-0 mt-0.5" aria-hidden="true" />
                  <div className="flex-1">
                    <h4 className="font-mono font-semibold text-aviation-text-primary mb-2">
                      {currentStepIndex === 0 && t('onboarding.tips.welcomeTitle')}
                      {completedSteps.includes('welcome') &&
                        currentStepIndex === 1 &&
                        t('onboarding.tips.readyToConnect')}
                      {completedSteps.includes('connect-provider') &&
                        currentStepIndex === 2 &&
                        t('onboarding.tips.providerConnected')}
                      {completedSteps.includes('deploy-function') &&
                        currentStepIndex === 3 &&
                        t('onboarding.tips.functionDeployed')}
                      {completedSteps.length === 4 && t('onboarding.tips.setupComplete')}
                    </h4>
                    <div className="text-sm text-aviation-text-secondary space-y-2 font-mono">
                      {currentStepIndex === 0 && (
                        <div className="space-y-2">
                          <p>{t('onboarding.tips.welcomeDesc')}</p>
                          {userRole && (
                            <div
                              className={`p-2 rounded text-xs font-mono ${
                                userRole === 'admin'
                                  ? 'bg-aviation-stratosphere/20 border border-aviation-stratosphere/40 text-aviation-stratosphere'
                                  : userRole === 'member'
                                    ? 'bg-aviation-cyan-dim border border-aviation-cyan/40 text-aviation-cyan'
                                    : 'bg-aviation-green-dim border border-aviation-green/40 text-aviation-green'
                              }`}
                              role="status"
                            >
                              {t('onboarding.tips.roleGreeting', { role: t(`onboarding.roles.${userRole}`) })}
                              {userRole === 'admin' && t('onboarding.tips.adminPrivileges')}
                              {userRole === 'member' && t('onboarding.tips.memberAccess')}
                              {userRole === 'viewer' && t('onboarding.tips.viewerAccess')}
                            </div>
                          )}
                        </div>
                      )}
                      {completedSteps.includes('welcome') && currentStepIndex === 1 && (
                        <p>{t('onboarding.tips.connectProviderDesc')}</p>
                      )}
                      {completedSteps.includes('connect-provider') && currentStepIndex === 2 && (
                        <div className="space-y-2">
                          <p>{t('onboarding.tips.deployFunctionDesc')}</p>
                          <div className="bg-aviation-green-dim border border-aviation-green/30 rounded p-2">
                            <p className="text-aviation-green text-xs font-mono font-medium">
                              {t('onboarding.tips.tokenSecure')}
                            </p>
                          </div>
                        </div>
                      )}
                      {completedSteps.includes('deploy-function') && currentStepIndex === 3 && (
                        <div className="space-y-2">
                          <p>{t('onboarding.tips.testFailoverDesc')}</p>
                          <div className="bg-aviation-cyan-dim border border-aviation-cyan/40 rounded p-2">
                            <p className="text-aviation-cyan text-xs font-mono font-medium">
                              {t('onboarding.tips.autoFailover')}
                            </p>
                          </div>
                        </div>
                      )}
                      {completedSteps.includes('test-failover') &&
                        userRole === 'admin' &&
                        currentStepIndex === 4 && (
                          <div className="space-y-2">
                            <p>{t('onboarding.tips.teamSetupDesc')}</p>
                            <div className="bg-aviation-cyan-dim border border-aviation-cyan/40 rounded p-2">
                              <p className="text-aviation-cyan text-xs font-mono font-medium">
                                {t('onboarding.tips.inviteTeam')}
                              </p>
                            </div>
                          </div>
                        )}
                      {completedSteps.includes('test-failover') &&
                        userRole !== 'admin' &&
                        currentStepIndex === 3 && (
                          <div className="space-y-2">
                            <p>{t('onboarding.tips.setupCompleteDesc')}</p>
                            <div className="bg-aviation-green-dim border border-aviation-green/40 rounded p-2">
                              <p className="text-aviation-green text-xs font-mono font-medium">
                                {t('onboarding.tips.productionReady')}
                              </p>
                            </div>
                          </div>
                        )}
                      {completedSteps.length === steps.length && (
                        <div className="space-y-2">
                          <p>{t('onboarding.tips.allCompleteDesc')}</p>
                          <div className="bg-aviation-stratosphere/10 border border-aviation-stratosphere/20 rounded p-2">
                            <p className="text-aviation-stratosphere text-xs font-mono">
                              {t('onboarding.tips.multiProviderDeployed')}
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

          <nav className="mt-8 flex justify-center" aria-label={t('onboarding.accessibility.stepNavigation')}>
            <div className="flex items-center gap-4" role="list">
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
                      aria-label={`${t('onboarding.accessibility.stepLabel', { number: index + 1, title: t(step.titleKey), status: isCompleted ? t('onboarding.accessibility.statusCompleted') : isActive ? t('onboarding.accessibility.statusActive') : t('onboarding.accessibility.statusPending') })}`}
                      tabIndex={0}
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
                          <CheckCircle className="w-6 h-6" aria-hidden="true" />
                        ) : (
                          Icon && <Icon className="w-6 h-6" aria-hidden="true" />
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
                          {t(step.titleKey)}
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
                        aria-hidden="true"
                      />
                    )}
                  </div>
                );
              })}
            </div>
          </nav>
        </div>
      </main>

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
            role="dialog"
            aria-modal="true"
            aria-labelledby="completion-title"
          >
            <motion.div
              className="text-center relative"
              initial={{ scale: 0.5, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ type: 'spring', bounce: 0.4 }}
            >
              <div className="absolute inset-0 pointer-events-none" aria-hidden="true">
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
                <CheckCircle className="w-12 h-12 text-aviation-bg-primary" aria-hidden="true" />
              </motion.div>

              <motion.h2
                id="completion-title"
                className="text-3xl font-bold text-aviation-text-primary mb-4 font-mono"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
              >
                {t('onboarding.completion.title')}
              </motion.h2>

              <motion.p
                className="text-aviation-text-secondary text-lg font-mono"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.5 }}
              >
                {t('onboarding.completion.subtitle')}
              </motion.p>
            </motion.div>
          </motion.div>
        </>
      )}

      <Footer showScrollToTop={false} />
    </div>
  );
}