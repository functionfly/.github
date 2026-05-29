declare global {
  interface Window {
    electronAPI?: {
      minimize: () => Promise<void>;
      maximize: () => Promise<void>;
      close: () => Promise<void>;
      isMaximized: () => Promise<boolean>;
      platform: () => Promise<string>;
      isTauri: boolean;
    };
    __TAURI__?: unknown;
    __TAURI_INTERNALS__?: unknown;
  }
}

/** True when running inside the FunctionFly Studio Tauri (or legacy Electron) shell. */
export function isDesktopApp(): boolean {
  if (typeof window === 'undefined') return false;
  return (
    window.electronAPI?.isTauri === true || '__TAURI__' in window || '__TAURI_INTERNALS__' in window
  );
}
