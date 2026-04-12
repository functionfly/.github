import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ToastPosition = 
  | 'top-left' 
  | 'top-center' 
  | 'top-right' 
  | 'bottom-left' 
  | 'bottom-center' 
  | 'bottom-right';

interface ToastSettingsState {
  position: ToastPosition;
  duration: number;
  richColors: boolean;
  closeButton: boolean;
  setPosition: (position: ToastPosition) => void;
  setDuration: (duration: number) => void;
  setRichColors: (enabled: boolean) => void;
  setCloseButton: (enabled: boolean) => void;
  resetToDefaults: () => void;
}

const DEFAULTS = {
  position: 'bottom-right' as ToastPosition,
  duration: 5000,
  richColors: true,
  closeButton: true,
};

export const useToastSettingsStore = create<ToastSettingsState>()(
  persist(
    (set) => ({
      position: DEFAULTS.position,
      duration: DEFAULTS.duration,
      richColors: DEFAULTS.richColors,
      closeButton: DEFAULTS.closeButton,
      setPosition: (position) => set({ position }),
      setDuration: (duration) => set({ duration }),
      setRichColors: (richColors) => set({ richColors }),
      setCloseButton: (closeButton) => set({ closeButton }),
      resetToDefaults: () => set({ ...DEFAULTS }),
    }),
    {
      name: 'toast-settings-storage',
    }
  )
);
