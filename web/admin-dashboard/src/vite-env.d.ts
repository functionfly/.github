// Vite environment variables type definitions

/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_ADMIN_API_BASE_URL: string;
  readonly VITE_SESSION_TIMEOUT: string;
  readonly VITE_IDLE_TIMEOUT: string;
  readonly VITE_MFA_REVERIFY_INTERVAL: string;
  readonly VITE_ENABLE_IP_WHITELIST: string;
  readonly VITE_ENABLE_DEVICE_FINGERPRINT: string;
  readonly VITE_ENABLE_AUDIT_LOGGING: string;
  readonly VITE_ENABLE_SESSION_RECORDING: string;
  readonly VITE_EXPECT_ZT_HEADERS: string;
  readonly VITE_DEVELOPMENT: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
