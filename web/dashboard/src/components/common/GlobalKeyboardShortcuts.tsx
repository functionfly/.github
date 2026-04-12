import { ROUTES } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';
import { useKeyboardShortcutsStore } from '@/stores/keyboardShortcutsStore';
import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

export function GlobalKeyboardShortcuts() {
  const navigate = useNavigate();
  const location = useLocation();
  const user = useAuthStore((state) => state.user);
  const { toggleHelp, setHelpOpen } = useKeyboardShortcutsStore();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger shortcuts when typing in inputs
      const target = e.target as HTMLElement;
      const isInputElement = 
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.contentEditable === 'true' ||
        target.closest('[role="dialog"]');

      // Handle ? key for help (allow even in some contexts, but not in text inputs)
      if (e.key === '?' && e.shiftKey) {
        if (!isInputElement) {
          e.preventDefault();
          toggleHelp();
        }
        return;
      }

      // Don't handle other shortcuts in input elements
      if (isInputElement) return;

      // Handle navigation shortcuts with 'g' prefix
      if (e.key === 'g' && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey) {
        // Start listening for the next key
        const handleNextKey = (nextE: KeyboardEvent) => {
          document.removeEventListener('keydown', handleNextKey);

          // Don't trigger if user is typing in an input now
          const nextTarget = nextE.target as HTMLElement;
          if (
            nextTarget.tagName === 'INPUT' ||
            nextTarget.tagName === 'TEXTAREA' ||
            nextTarget.contentEditable === 'true'
          ) {
            return;
          }

          switch (nextE.key) {
            case 'd':
              if (location.pathname !== ROUTES.DASHBOARD) {
                navigate(ROUTES.DASHBOARD);
              }
              break;
            case 'o':
              if (location.pathname !== ROUTES.OVERVIEW) {
                navigate(ROUTES.OVERVIEW);
              }
              break;
            case 'f':
              if (location.pathname !== ROUTES.FUNCTIONS) {
                navigate(ROUTES.FUNCTIONS);
              }
              break;
            case 'r':
              if (location.pathname !== ROUTES.FRG) {
                navigate(ROUTES.FRG);
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
              navigate('/');
              break;
            case 'n':
              if (location.pathname !== '/notifications') {
                navigate('/notifications');
              }
              break;
            case 'm':
              navigate('/marketplace/functions');
              break;
          }
        };

        document.addEventListener('keydown', handleNextKey, { once: true });
        
        // Clear the listener after a short timeout to avoid stuck state
        setTimeout(() => {
          document.removeEventListener('keydown', handleNextKey);
        }, 1000);
        
        return;
      }

      // Handle direct shortcuts
      switch (e.key) {
        case 'b':
          if (e.ctrlKey || e.metaKey) {
            e.preventDefault();
            // Could toggle sidebar if we have access to sidebar state
          }
          break;
        case 'k':
          if ((e.ctrlKey || e.metaKey) && !e.shiftKey) {
            e.preventDefault();
            // Open command palette - can be implemented later
          }
          break;
        case 'n':
          if (e.ctrlKey || e.metaKey) {
            e.preventDefault();
            // Could trigger new item creation based on current page
          }
          break;
        case 'Escape':
          // Close the help modal if open
          setHelpOpen(false);
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [navigate, location.pathname, toggleHelp, setHelpOpen]);

  return null;
}
