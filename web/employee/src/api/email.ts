import apiClient from './client';

export interface EmailAccount {
  id: string;
  employee_id: string;
  email: string;
  display_name?: string;
  provider: string;
  aliases: string[];
  groups: string[];
  status: string;
  provisioned_at?: string;
}

export const emailApi = {
  list: () => apiClient.get<{ accounts: EmailAccount[] }>('/v1/email'),
  provision: (employeeId: string, data?: { display_name?: string; aliases?: string[] }) =>
    apiClient.post<{ account: EmailAccount }>('/v1/email/provision', { employee_id: employeeId, ...data }),
  updateStatus: (id: string, status: string) => apiClient.patch(`/v1/email/${id}`, { status }),
};
