import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { FormError } from '@/components/ui/form-error';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useSignupForm } from '@/hooks/useAuthForms';
import { getMarketingPageUrl } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { AnimatePresence, motion } from 'framer-motion';
import {
  ArrowLeft,
  ArrowRight,
  AtSign,
  Building2,
  Calendar,
  Check,
  Eye,
  EyeOff,
  Key,
  Loader2,
  Mail,
  Shield,
  User,
  X,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const todayISODate = () => new Date().toISOString().slice(0, 10);

// Password requirement checklist component
function PasswordRequirements({ password }: { password: string }) {
  const requirements = [
    { key: 'length', label: 'At least 8 characters', test: (p: string) => p.length >= 8 },
    { key: 'uppercase', label: 'One uppercase letter', test: (p: string) => /[A-Z]/.test(p) },
    { key: 'lowercase', label: 'One lowercase letter', test: (p: string) => /[a-z]/.test(p) },
    { key: 'number', label: 'One number', test: (p: string) => /[0-9]/.test(p) },
    { key: 'special', label: 'One special character', test: (p: string) => /[^A-Za-z0-9]/.test(p) },
  ];

  return (
    <div className="mt-2 space-y-1">
      {requirements.map((req) => {
        const passed = req.test(password);
        return (
          <div
            key={req.key}
            className={cn(
              'flex items-center gap-2 text-xs transition-colors',
              passed ? 'text-green-500' : 'text-text-muted'
            )}
          >
            {passed ? <Check className="w-3 h-3" /> : <X className="w-3 h-3" />}
            <span>{req.label}</span>
          </div>
        );
      })}
    </div>
  );
}

// Step indicator
function StepIndicator({ currentStep, totalSteps }: { currentStep: number; totalSteps: number }) {
  return (
    <div className="flex items-center justify-center gap-2 mb-8">
      {Array.from({ length: totalSteps }).map((_, index) => (
        <div key={index} className="flex items-center">
          <div
            className={cn(
              'w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors',
              index + 1 < currentStep
                ? 'bg-green-500 text-white'
                : index + 1 === currentStep
                  ? 'bg-[#6366f1] text-white'
                  : 'bg-bg-tertiary text-text-muted'
            )}
          >
            {index + 1 < currentStep ? <Check className="w-4 h-4" /> : index + 1}
          </div>
          {index < totalSteps - 1 && (
            <div
              className={cn(
                'w-12 h-0.5 mx-1',
                index + 1 < currentStep ? 'bg-green-500' : 'bg-bg-tertiary'
              )}
            />
          )}
        </div>
      ))}
    </div>
  );
}

// Step 1: Email & Password
function Step1EmailPassword({
  formData,
  updateFormData,
  errors,
  register,
  watch,
}: {
  formData: Record<string, string>;
  updateFormData: (data: Record<string, string>) => void;
  errors: Record<string, { message?: string }>;
  register: (name: string) => Record<string, unknown>;
  watch: (name: string) => string;
}) {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const password = watch('password');

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-center">Create your account</h3>
      <p className="text-sm text-text-muted text-center mb-4">
        Start with your email and a secure password
      </p>

      <div className="space-y-2">
        <Label htmlFor="email" className="flex items-center gap-2">
          <Mail className="w-4 h-4" />
          Email <span className="text-error">*</span>
        </Label>
        <Input id="email" type="email" placeholder="you@example.com" {...register('email')} />
        {errors.email && <div className="text-xs text-error">{errors.email.message}</div>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="password" className="flex items-center gap-2">
          <Key className="w-4 h-4" />
          Password <span className="text-error">*</span>
        </Label>
        <div className="relative">
          <Input
            id="password"
            type={showPassword ? 'text' : 'password'}
            placeholder="Create a strong password"
            {...register('password')}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary"
          >
            {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
        </div>
        {password && <PasswordRequirements password={password} />}
        {errors.password && <div className="text-xs text-error">{errors.password.message}</div>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="confirmPassword" className="flex items-center gap-2">
          <Key className="w-4 h-4" />
          Confirm Password <span className="text-error">*</span>
        </Label>
        <div className="relative">
          <Input
            id="confirmPassword"
            type={showConfirmPassword ? 'text' : 'password'}
            placeholder="Confirm your password"
            {...register('confirmPassword')}
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary"
          >
            {showConfirmPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
        </div>
        {errors.confirmPassword && (
          <div className="text-xs text-error">{errors.confirmPassword.message}</div>
        )}
      </div>
    </div>
  );
}

// Step 2: Profile Info
function Step2Profile({
  errors,
  register,
}: {
  errors: Record<string, { message?: string }>;
  register: (name: string) => Record<string, unknown>;
}) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-center">Tell us about yourself</h3>
      <p className="text-sm text-text-muted text-center mb-4">
        This helps us personalize your experience
      </p>

      <div className="space-y-2">
        <Label htmlFor="name" className="flex items-center gap-2">
          <User className="w-4 h-4" />
          Full Name <span className="text-error">*</span>
        </Label>
        <Input id="name" type="text" placeholder="John Doe" {...register('name')} />
        {errors.name && <div className="text-xs text-error">{errors.name.message}</div>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="username" className="flex items-center gap-2 text-text-secondary">
          <AtSign className="w-4 h-4" />
          Username <span className="text-text-muted text-xs">(optional)</span>
        </Label>
        <Input id="username" type="text" placeholder="johndoe" {...register('username')} />
        {errors.username && <div className="text-xs text-error">{errors.username.message}</div>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="dateOfBirth" className="flex items-center gap-2">
          <Calendar className="w-4 h-4" />
          Date of birth <span className="text-error">*</span>
        </Label>
        <Input
          id="dateOfBirth"
          type="date"
          autoComplete="bday"
          max={todayISODate()}
          {...register('dateOfBirth')}
        />
        {errors.dateOfBirth && (
          <div className="text-xs text-error">{errors.dateOfBirth.message}</div>
        )}
        <p className="text-xs text-text-muted">You must be at least 13 years old.</p>
      </div>
    </div>
  );
}

// Step 3: Company & Invite (Optional)
function Step3Optional({
  errors,
  register,
}: {
  errors: Record<string, { message?: string }>;
  register: (name: string) => Record<string, unknown>;
}) {
  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-center">Almost done!</h3>
      <p className="text-sm text-text-muted text-center mb-4">
        Add optional details or skip this step
      </p>

      <div className="space-y-2">
        <Label htmlFor="companyName" className="flex items-center gap-2 text-text-secondary">
          <Building2 className="w-4 h-4" />
          Company name <span className="text-text-muted text-xs">(optional)</span>
        </Label>
        <Input id="companyName" type="text" placeholder="Acme Inc" {...register('companyName')} />
        {errors.companyName && (
          <div className="text-xs text-error">{errors.companyName.message}</div>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="inviteCode" className="flex items-center gap-2 text-text-secondary">
          <Key className="w-4 h-4" />
          Invite Code <span className="text-text-muted text-xs">(optional)</span>
        </Label>
        <Input
          id="inviteCode"
          type="text"
          placeholder="Enter your invite code"
          {...register('inviteCode')}
        />
        {errors.inviteCode && <div className="text-xs text-error">{errors.inviteCode.message}</div>}
      </div>

      <div className="flex items-start gap-3 pt-2">
        <Checkbox id="terms" {...register('termsAccepted')} className="mt-1" />
        <div className="space-y-1">
          <label
            htmlFor="terms"
            className="text-sm text-text-secondary leading-tight cursor-pointer"
          >
            I agree to the{' '}
            <a
              href={getMarketingPageUrl('/terms')}
              className="text-brand-500 hover:underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              Terms of Service
            </a>{' '}
            and{' '}
            <a
              href={getMarketingPageUrl('/privacy')}
              className="text-brand-500 hover:underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              Privacy Policy
            </a>
          </label>
          {errors.termsAccepted && (
            <div className="text-xs text-error">{errors.termsAccepted.message}</div>
          )}
        </div>
      </div>
    </div>
  );
}

// Main SignupWizard component
export function SignupWizard() {
  const navigate = useNavigate();
  const { signup, isLoading, error, clearError } = useAuthStore();
  const [currentStep, setCurrentStep] = useState(1);
  const [formData, setFormData] = useState<Record<string, string>>({});
  const totalSteps = 3;

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isValid },
    trigger,
  } = useSignupForm();

  const handleNext = async () => {
    let fieldsToValidate: string[] = [];

    // Validate fields for current step
    if (currentStep === 1) {
      fieldsToValidate = ['email', 'password', 'confirmPassword'];
    } else if (currentStep === 2) {
      fieldsToValidate = ['name', 'dateOfBirth'];
    }

    const isStepValid = await trigger(fieldsToValidate);
    if (isStepValid) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    setCurrentStep(currentStep - 1);
  };

  const onSubmit = async (data: any) => {
    clearError();
    try {
      const response = await signup(data);

      if (response.requiresVerification) {
        navigate('/auth/verify-email', {
          state: {
            message: response.message,
            email: data.email,
            emailSent: response.emailSent,
          },
        });
      }
    } catch {
      // Error is handled by the store
    }
  };

  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      {/* Header */}
      <header className="border-b border-border-subtle bg-bg-secondary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link to="/" className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-linear-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                <Zap className="w-5 h-5 text-white" fill="currentColor" />
              </div>
              <span className="text-xl font-bold gradient-text">FunctionFly</span>
            </Link>
            <Link to="/login">
              <Button variant="ghost" size="sm">
                Already have an account? Sign in
              </Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          {/* Step Indicator */}
          <StepIndicator currentStep={currentStep} totalSteps={totalSteps} />

          {/* Error Display */}
          {error && <FormError error={error} className="mb-6" />}

          {/* Form */}
          <form onSubmit={handleSubmit(onSubmit)}>
            <AnimatePresence mode="wait">
              <motion.div
                key={currentStep}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.2 }}
              >
                {currentStep === 1 && (
                  <Step1EmailPassword
                    formData={formData}
                    updateFormData={setFormData}
                    errors={errors}
                    register={register}
                    watch={watch}
                  />
                )}
                {currentStep === 2 && <Step2Profile errors={errors} register={register} />}
                {currentStep === 3 && <Step3Optional errors={errors} register={register} />}
              </motion.div>
            </AnimatePresence>

            {/* Navigation Buttons */}
            <div className="flex justify-between mt-8">
              {currentStep > 1 ? (
                <Button type="button" variant="outline" onClick={handleBack}>
                  <ArrowLeft className="w-4 h-4 mr-2" />
                  Back
                </Button>
              ) : (
                <div />
              )}

              {currentStep < totalSteps ? (
                <Button type="button" onClick={handleNext} className="gap-2">
                  Continue
                  <ArrowRight className="w-4 h-4" />
                </Button>
              ) : (
                <Button type="submit" disabled={isLoading} className="gap-2">
                  {isLoading ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Creating account...
                    </>
                  ) : (
                    <>
                      Create Account
                      <ArrowRight className="w-4 h-4" />
                    </>
                  )}
                </Button>
              )}
            </div>
          </form>

          {/* Security Badge */}
          <div className="flex items-center justify-center gap-2 mt-6 text-xs text-text-muted">
            <Shield className="w-3 h-3" />
            <span>Secured by reCAPTCHA</span>
          </div>
        </div>
      </main>
    </div>
  );
}
