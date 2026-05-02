import { useThemeStore } from '@/stores/themeStore';
import { Monitor, Moon, Sun } from 'lucide-react';

type Theme = 'light' | 'dark' | 'system';

export function DarkModeToggle() {
  const { theme, setTheme } = useThemeStore();

  const themes: { value: Theme; icon: React.ReactNode; label: string }[] = [
    { value: 'light', icon: <Sun className="w-5 h-5" />, label: 'Light' },
    { value: 'dark', icon: <Moon className="w-5 h-5" />, label: 'Dark' },
    { value: 'system', icon: <Monitor className="w-5 h-5" />, label: 'System' },
  ];

  return (
    <div className="flex items-center gap-2">
      {themes.map((t) => (
        <button
          key={t.value}
          onClick={() => setTheme(t.value)}
          className={`p-2.5 rounded-full transition-all duration-200 shadow-sm ${
            theme === t.value
              ? 'bg-blue-600 text-white shadow-md ring-2 ring-blue-400 ring-offset-2 dark:ring-offset-slate-900'
              : 'bg-white dark:bg-slate-700 text-gray-600 dark:text-slate-300 hover:bg-gray-100 dark:hover:bg-slate-600 border border-gray-200 dark:border-slate-600'
          }`}
          title={t.label}
        >
          {t.icon}
        </button>
      ))}
    </div>
  );
}

export default DarkModeToggle;