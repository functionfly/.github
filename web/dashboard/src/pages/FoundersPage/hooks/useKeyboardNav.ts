import { ROUTES } from '@/lib/constants';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const COMBO_TIMEOUT = 500;

export function useKeyboardNav() {
  const navigate = useNavigate();

  useEffect(() => {
    let pendingG = false;
    let timer: ReturnType<typeof setTimeout>;

    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable
      ) {
        return;
      }

      if (e.key === 'g' && !pendingG) {
        pendingG = true;
        timer = setTimeout(() => {
          pendingG = false;
        }, COMBO_TIMEOUT);
        return;
      }

      if (e.key === 'f' && pendingG) {
        pendingG = false;
        clearTimeout(timer);
        navigate(ROUTES.FOUNDERS);
        return;
      }

      pendingG = false;
      clearTimeout(timer);
    };

    window.addEventListener('keydown', handler);
    return () => {
      window.removeEventListener('keydown', handler);
      clearTimeout(timer);
    };
  }, [navigate]);
}
