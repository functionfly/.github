import { Moon, Sun, Palette } from 'lucide-react';
import { useThemeStore, type Theme } from '@/stores/themeStore';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface ThemeToggleProps {
  className?: string;
  variant?: 'button' | 'switch' | 'selector';
}

const themeOptions = [
  { value: 'dark' as Theme, label: 'Dark', icon: Moon, color: 'text-slate-400' },
  { value: 'light' as Theme, label: 'Light', icon: Sun, color: 'text-yellow-500' },
];

export function ThemeToggle({ className, variant = 'button' }: ThemeToggleProps) {
  const { theme, setTheme } = useThemeStore();

  if (variant === 'selector') {
    return (
      <div className={cn("flex flex-wrap gap-2", className)}>
        {themeOptions.map((option) => {
          const Icon = option.icon;
          return (
            <Button
              key={option.value}
              variant={theme === option.value ? "default" : "outline"}
              size="sm"
              onClick={() => setTheme(option.value)}
              className={cn(
                "flex items-center gap-2",
                theme === option.value && "bg-brand-500 hover:bg-brand-600 text-white"
              )}
            >
              <Icon className={cn("w-4 h-4", theme === option.value ? "text-white" : option.color)} />
              <span className="hidden sm:inline">{option.label}</span>
            </Button>
          );
        })}
      </div>
    );
  }

  if (variant === 'switch') {
    return (
      <div className={cn("flex items-center gap-2", className)}>
        <Sun className="h-4 w-4" style={{ color: 'var(--text-muted)' }} />
        <button
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          className={cn(
            "relative inline-flex h-6 w-11 items-center rounded-full transition-colors",
            theme === 'dark' ? 'bg-brand-500' : 'bg-[var(--text-muted)]'
          )}
        >
          <span
            className={cn(
              "inline-block h-4 w-4 transform rounded-full bg-white transition-transform",
              theme === 'dark' ? 'translate-x-6' : 'translate-x-1'
            )}
          />
        </button>
        <Moon className="h-4 w-4" style={{ color: 'var(--text-muted)' }} />
      </div>
    );
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
      className={cn(
        "transition-colors",
        className
      )}
      style={{
        color: 'var(--text-secondary)',
        backgroundColor: 'transparent',
      }}
      title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
    >
      <span style={{ color: theme === 'dark' ? 'var(--text-secondary)' : 'var(--text-accent)' }}>
        {theme === 'dark' ? (
          <Sun className="h-5 w-5" />
        ) : (
          <Moon className="h-5 w-5" />
        )}
      </span>
      <span className="sr-only">
        Switch to {theme === 'dark' ? 'light' : 'dark'} mode
      </span>
    </Button>
  );
}
