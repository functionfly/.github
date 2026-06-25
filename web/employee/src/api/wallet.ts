import apiClient from './client';

export interface WalletPass {
  id: string;
  employee_id: string;
  pass_type: string;
  platform: string;
  pass_id: string;
  qr_token: string;
  qr_expires_at: string;
  status: string;
  installed_at?: string;
}

export const walletApi = {
  generate: (platform: 'apple_wallet' | 'google_wallet') =>
    apiClient.post<{ pass: WalletPass }>('/v1/wallet/generate', { platform }),
  verify: (qrToken: string) =>
    apiClient.post<{ valid: boolean; employee?: { ffid: string; name: string; department: string; clearance: string; status: string } }>('/v1/wallet/verify', { qr_token: qrToken }),
  revoke: (id: string) => apiClient.post(`/v1/wallet/${id}/revoke`),
};
