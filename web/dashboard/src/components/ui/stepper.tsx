'use client';

import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { Check, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

const stepperVariants = cva('flex', {
  variants: {
    orientation: {
      horizontal: 'flex-row items-center gap-2',
      vertical: 'flex-col gap-4',
    },
    size: {
      default: '',
      sm: '',
      lg: '',
    },
  },
  defaultVariants: {
    orientation: 'horizontal',
    size: 'default',
  },
});

export interface Step {
  id: string;
  title: string;
  description?: string;
  icon?: React.ReactNode;
  optional?: boolean;
  disabled?: boolean;
}

interface StepperContextValue {
  steps: Step[];
  currentStep: number;
  completedSteps: string[];
  orientation: 'horizontal' | 'vertical';
  size: 'default' | 'sm' | 'lg';
  clickable?: boolean;
  onStepClick?: (index: number) => void;
}

const StepperContext = React.createContext<StepperContextValue | undefined>(undefined);

function useStepper() {
  const context = React.useContext(StepperContext);
  if (!context) {
    throw new Error('Stepper components must be used within a Stepper');
  }
  return context;
}

export interface StepperProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof stepperVariants> {
  steps: Step[];
  currentStep: number;
  completedSteps?: string[];
  clickable?: boolean;
  onStepClick?: (index: number) => void;
  asChild?: boolean;
}

const Stepper = React.forwardRef<HTMLDivElement, StepperProps>(
  (
    {
      className,
      orientation = 'horizontal',
      size = 'default',
      steps,
      currentStep,
      completedSteps = [],
      clickable = false,
      onStepClick,
      asChild = false,
      ...props
    },
    ref
  ) => {
    const Comp = asChild ? Slot : 'div';
    const contextValue = React.useMemo(
      () => ({
        steps,
        currentStep,
        completedSteps,
        orientation: orientation || 'horizontal',
        size: size || 'default',
        clickable,
        onStepClick,
      }),
      [steps, currentStep, completedSteps, orientation, size, clickable, onStepClick]
    );

    return (
      <StepperContext.Provider value={contextValue}>
        <Comp
          ref={ref}
          className={cn(stepperVariants({ orientation, size, className }))}
          {...props}
        />
      </StepperContext.Provider>
    );
  }
);
Stepper.displayName = 'Stepper';

const stepIndicatorVariants = cva(
  'flex items-center justify-center rounded-full border-2 transition-all duration-200',
  {
    variants: {
      status: {
        pending: 'border-border-subtle bg-transparent text-text-muted',
        current: 'border-brand-500 bg-brand-500/10 text-brand-500',
        completed: 'border-brand-500 bg-brand-500 text-white',
        error: 'border-error bg-error/10 text-error',
      },
      size: {
        default: 'h-10 w-10 text-sm font-medium',
        sm: 'h-8 w-8 text-xs font-medium',
        lg: 'h-12 w-12 text-base font-semibold',
      },
    },
    defaultVariants: {
      status: 'pending',
      size: 'default',
    },
  }
);

interface StepIndicatorProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof stepIndicatorVariants> {
  asChild?: boolean;
  stepNumber: number;
  isCompleted?: boolean;
  isCurrent?: boolean;
  isError?: boolean;
  icon?: React.ReactNode;
}

const StepIndicator = React.forwardRef<HTMLDivElement, StepIndicatorProps>(
  ({ className, size, stepNumber, isCompleted, isCurrent, isError, icon, ...props }, ref) => {
    const { size: contextSize } = useStepper();
    const finalSize = size || contextSize;

    let status: 'pending' | 'current' | 'completed' | 'error' = 'pending';
    if (isError) status = 'error';
    else if (isCompleted) status = 'completed';
    else if (isCurrent) status = 'current';

    return (
      <div
        ref={ref}
        className={cn(stepIndicatorVariants({ status, size: finalSize, className }))}
        {...props}
      >
        {isCompleted ? <Check className="h-5 w-5" /> : icon || stepNumber}
      </div>
    );
  }
);
StepIndicator.displayName = 'StepIndicator';

const stepContentVariants = cva('', {
  variants: {
    orientation: {
      horizontal: 'flex flex-col items-center text-center',
      vertical: 'flex flex-row items-center gap-4',
    },
  },
  defaultVariants: {
    orientation: 'horizontal',
  },
});

export interface StepItemProps extends React.HTMLAttributes<HTMLDivElement> {
  step: Step;
  index: number;
  asChild?: boolean;
}

const StepItem = React.forwardRef<HTMLDivElement, StepItemProps>(
  ({ className, step, index, asChild = false, ...props }, ref) => {
    const { orientation, currentStep, completedSteps, clickable, onStepClick, size } = useStepper();
    const Comp = asChild ? Slot : 'div';

    const isCompleted = completedSteps.includes(step.id);
    const isCurrent = index === currentStep;
    const isPending = index > currentStep && !isCompleted;
    const isError = false;

    const canClick = clickable && !step.disabled && (isCompleted || index <= currentStep + 1);

    return (
      <Comp
        ref={ref}
        className={cn(
          stepContentVariants({ orientation }),
          canClick && 'cursor-pointer',
          step.disabled && 'opacity-50',
          className
        )}
        onClick={() => canClick && onStepClick?.(index)}
        {...props}
      >
        <StepIndicator
          stepNumber={index + 1}
          isCompleted={isCompleted}
          isCurrent={isCurrent}
          isError={isError}
          icon={step.icon}
          size={size}
        />
        <div className={cn(orientation === 'vertical' && 'flex-1')}>
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'text-sm font-medium',
                isCurrent ? 'text-text-primary' : 'text-text-secondary',
                isCompleted && 'text-brand-500'
              )}
            >
              {step.title}
            </span>
            {step.optional && (
              <span className="text-xs text-text-muted">(Optional)</span>
            )}
          </div>
          {step.description && (
            <p className={cn('mt-0.5 text-xs text-text-muted', orientation === 'horizontal' && 'max-w-[150px]')}>
              {step.description}
            </p>
          )}
        </div>
      </Comp>
    );
  }
);
StepItem.displayName = 'StepItem';

const stepConnectorVariants = cva('transition-all duration-200', {
  variants: {
    orientation: {
      horizontal: 'h-0.5 flex-1',
      vertical: 'w-0.5 flex-1',
    },
    status: {
      pending: 'bg-border-subtle',
      completed: 'bg-brand-500',
    },
  },
  defaultVariants: {
    orientation: 'horizontal',
    status: 'pending',
  },
});

interface StepConnectorProps extends React.HTMLAttributes<HTMLDivElement>, VariantProps<typeof stepConnectorVariants> {}

const StepConnector = React.forwardRef<HTMLDivElement, StepConnectorProps>(
  ({ className, orientation, status, ...props }, ref) => {
    const { orientation: contextOrientation } = useStepper();
    const finalOrientation = orientation || contextOrientation;

    return (
      <div
        ref={ref}
        className={cn(stepConnectorVariants({ orientation: finalOrientation, status, className }))}
        {...props}
      />
    );
  }
);
StepConnector.displayName = 'StepConnector';

const StepperContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    const { steps, currentStep, completedSteps, orientation } = useStepper();

    return (
      <div
        ref={ref}
        className={cn(
          'flex',
          orientation === 'horizontal' ? 'flex-row items-start' : 'flex-col',
          className
        )}
        {...props}
      >
        {steps.map((step, index) => (
          <React.Fragment key={step.id}>
            <StepItem step={step} index={index} />
            {index < steps.length - 1 && (
              <StepConnector
                status={index < currentStep || completedSteps.includes(step.id) ? 'completed' : 'pending'}
              />
            )}
          </React.Fragment>
        ))}
      </div>
    );
  }
);
StepperContent.displayName = 'StepperContent';

export { Stepper, StepperContent, StepItem, StepIndicator, StepConnector, useStepper };
