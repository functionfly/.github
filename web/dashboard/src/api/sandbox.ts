import { apiClient } from './client';

// ==================== Types ====================

export type SandboxTier = 'wasm' | 'gvisor' | 'docker' | 'microvm';

export interface TierInfo {
  id: string;
  name: string;
  description: string;
  available: boolean;
  isolation_level: string;
  status: string;
}

export interface SystemInfo {
  kernel: string;
  arch: string;
  cgroup_version: string;
  seccomp_enabled: boolean;
  user_namespace_enabled: boolean;
}

export interface SandboxStatus {
  gvisor_available: boolean;
  gvisor_path?: string;
  gvisor_version?: string;
  docker_available: boolean;
  docker_version?: string;
  active_tier: SandboxTier;
  supported_tiers: TierInfo[];
  last_checked: string;
  system_info: SystemInfo;
}

export interface SandboxTierConfig {
  sandbox_tier: SandboxTier;
  memory_mb?: number;
  cpu_limit?: number;
  timeout_ms?: number;
  network_enabled?: boolean;
}

export interface SandboxTiersResponse {
  tiers: TierInfo[];
  active_tier: SandboxTier;
}

// ==================== API Functions ====================

export async function getSandboxStatus(): Promise<SandboxStatus> {
  return apiClient.get<SandboxStatus>('/v1/sandbox/status');
}

export async function getSandboxTiers(): Promise<SandboxTiersResponse> {
  return apiClient.get<SandboxTiersResponse>('/v1/sandbox/tiers');
}

export async function updateFunctionSandboxTier(
  functionId: string,
  config: SandboxTierConfig
): Promise<{ status: string; function_id: string; sandbox_tier: string; message: string }> {
  return apiClient.put(`/v1/functions/${functionId}/sandbox`, config);
}

// ==================== Helpers ====================

export function getTierColor(tier: SandboxTier): string {
  switch (tier) {
    case 'wasm':
      return 'bg-green-500/20 text-green-400 border-green-500/30';
    case 'gvisor':
      return 'bg-cyan-500/20 text-cyan-400 border-cyan-500/30';
    case 'docker':
      return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
    case 'microvm':
      return 'bg-purple-500/20 text-purple-400 border-purple-500/30';
    default:
      return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
  }
}

export function getTierLabel(tier: SandboxTier): string {
  switch (tier) {
    case 'wasm': return 'WebAssembly';
    case 'gvisor': return 'gVisor';
    case 'docker': return 'Docker';
    case 'microvm': return 'MicroVM';
    default: return tier;
  }
}

export function getIsolationDescription(level: string): string {
  switch (level) {
    case 'memory_safe': return 'Memory-safe linear memory with fuel metering';
    case 'kernel_namespace': return 'User-space kernel with syscall interception';
    case 'container': return 'Container isolation with security hardening';
    case 'hardware_virtualization': return 'Full hardware virtualization (KVM)';
    default: return level;
  }
}
