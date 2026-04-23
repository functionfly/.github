import { evaluatePasswordStrength } from '@/lib/validation';
import { useTranslation } from 'react-i18next';
import { Check, X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface PasswordStrengthIndicatorProps {
  password: string;
  className?: string;
}

export function PasswordStrengthIndicator({ password, className }: PasswordStrengthIndicatorProps) {
  const { t } = useTranslation();
  const strength = evaluatePasswordStrength(password);

  if (!password) return null;

  return (
    <div className={cn('space-y-2', className)}>
      {/* Strength bar */}
      <div className="flex items-center gap-2">
        <div className="flex-1 bg-bg-secondary rounded-full h-2 overflow-hidden">
          <div
            className={cn(
              'h-full transition-all duration-300 ease-in-out rounded-full',
              strength.color
            )}
            style={{ width: `${(strength.score / strength.maxScore) * 100}%` }}
          />
        </div>
        <span className={cn(
          'text-xs font-medium px-2 py-1 rounded-full',
          {
            'bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-400': strength.strength === 'weak',
            'bg-orange-100 text-orange-700 dark:bg-orange-900/20 dark:text-orange-400': strength.strength === 'weak' && strength.score >= 2,
            'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400': strength.strength === 'medium',
            'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400': strength.strength === 'strong',
            'bg-green-200 text-green-800 dark:bg-green-900/30 dark:text-green-300': strength.strength === 'very-strong',
          }
        )}>
          {strength.label}
        </span>
      </div>

      {/* Requirements checklist */}
      <div className="grid grid-cols-1 gap-1 text-xs">
        {Object.entries(strength.checks).map(([key, passed]) => (
          <div key={key} className="flex items-center gap-2">
            {passed ? (
              <Check className="w-3 h-3 text-green-500" />
            ) : (
              <X className="w-3 h-3 text-red-400" />
            )}
            <span className={cn(
              passed ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'
            )}>
              {getRequirementLabel(key, t)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function getRequirementLabel(key: string, t: (key: string) => string): string {
  const labels = {
    length: t('passwordStrength.length'),
    uppercase: t('passwordStrength.uppercase'),
    lowercase: t('passwordStrength.lowercase'),
    number: t('passwordStrength.number'),
    special: t('passwordStrength.special'),
  };
  return labels[key as keyof typeof labels] || key;
}