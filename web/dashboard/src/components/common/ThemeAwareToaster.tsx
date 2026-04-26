import { Toaster } from 'sonner';
import { useTheme } from './ThemeProvider';

export function ThemeAwareToaster() {
  const { theme } = useTheme();

  return (
    <Toaster
      position="bottom-right"
      theme={theme}
      toastOptions={{
        style: {
          background: theme === 'dark' ? '#1a1a24' : '#ffffff',
          border: theme === 'dark' ? '1px solid #2a2a3a' : '1px solid #e5e5e5',
          color: theme === 'dark' ? '#ffffff' : '#1a1a24',
        },
      }}
    />
  );
}