import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { ROUTES } from '@/lib/constants';

export function GlobalKeyboardShortcuts() {
  const navigate = useNavigate();
  const location = useLocation();
  const user = useAuthStore((state) => state.user);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Only handle shortcuts when not typing in an input
      const target = e.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.contentEditable === 'true' ||
        target.closest('[role="dialog"]')
      ) {
        return;
      }

      // Handle navigation shortcuts with 'g' prefix
      if (e.key === 'g' && !e.ctrlKey && !e.metaKey && !e.altKey) {
        // Start listening for the next key
        const handleNextKey = (nextE: KeyboardEvent) => {
          document.removeEventListener('keydown', handleNextKey);

          switch (nextE.key) {
            case 'd':
              if (location.pathname !== ROUTES.DASHBOARD) {
                navigate(ROUTES.DASHBOARD);
              }
              break;
            case 'f':
              if (location.pathname !== ROUTES.FUNCTIONS) {
                navigate(ROUTES.FUNCTIONS);
              }
              break;
            case 'p':
              if (location.pathname !== ROUTES.PROVIDERS) {
                navigate(ROUTES.PROVIDERS);
              }
              break;
            case 'a':
              if (location.pathname !== ROUTES.ANALYTICS) {
                navigate(ROUTES.ANALYTICS);
              }
              break;
            case 's':
              if (location.pathname !== ROUTES.SETTINGS) {
                navigate(ROUTES.SETTINGS);
              }
              break;
            case 'h':
              navigate('/'); // Go to landing page
              break;
          }
        };

        document.addEventListener('keydown', handleNextKey, { once: true });
        return;
      }

      // Handle direct shortcuts
      switch (e.key) {
        case 'b':
          if (e.ctrlKey || e.metaKey) {
            e.preventDefault();
            // Could toggle sidebar if we have access to sidebar state
            // For now, just focus on navigation shortcuts
          }
          break;
        case '?':
          if (e.shiftKey) {
            e.preventDefault();
            // Could show keyboard shortcuts help modal
            console.log('Show keyboard shortcuts help');
          }
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [navigate, location.pathname]);

  return null; // This component doesn't render anything
}