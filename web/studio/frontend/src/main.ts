import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';

declare global {
  interface Window {
    electronAPI: {
      minimize: () => Promise<void>;
      maximize: () => Promise<void>;
      close: () => Promise<void>;
      isMaximized: () => Promise<boolean>;
      platform: () => Promise<string>;
      isTauri: boolean;
    };
  }
}

window.electronAPI = {
  minimize: () => invoke('minimize'),
  maximize: () => invoke('maximize'),
  close: () => invoke('close'),
  isMaximized: () => invoke('is_maximized'),
  platform: () => invoke('platform'),
  isTauri: true,
};

listen('deep-link', (event) => {
  console.log('Deep link received:', event.payload);
  const url = event.payload as string;
  window.dispatchEvent(new CustomEvent('studio-deep-link', { detail: url }));
});

console.log('FunctionFly Studio initialized');