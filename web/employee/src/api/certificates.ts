import apiClient from './client';

export interface EmployeeCertificate {
  id: string;
  employee_id: string;
  certificate_serial: string;
  certificate_type: string;
  subject: string;
  issuer: string;
  device_id?: string;
  device_name?: string;
  issued_at: string;
  expires_at: string;
  revoked_at?: string;
  status: string;
}

export const certificatesApi = {
  list: () => apiClient.get<{ certificates: EmployeeCertificate[] }>('/v1/certificates'),
  get: (id: string) => apiClient.get<{ certificate: EmployeeCertificate }>(`/v1/certificates/${id}`),
  issue: (data: { employee_id: string; certificate_type?: string; device_id?: string; device_name?: string }) => apiClient.post<{ certificate: EmployeeCertificate }>('/v1/certificates', data),
  revoke: (id: string, reason: string) => apiClient.post(`/v1/certificates/${id}/revoke`, { reason }),
};
