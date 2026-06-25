import apiClient from './client';

export interface Device {
  id: string;
  employee_id?: string;
  device_name: string;
  device_type: string;
  serial_number?: string;
  os?: string;
  os_version?: string;
  manufacturer?: string;
  model?: string;
  compliance_status: string;
  status: string;
  last_seen_at?: string;
  enrolled_at?: string;
}

export const devicesApi = {
  list: () => apiClient.get<{ devices: Device[] }>('/v1/devices'),
  get: (id: string) => apiClient.get<{ device: Device }>(`/v1/devices/${id}`),
  register: (data: Partial<Device>) => apiClient.post<{ device: Device }>('/v1/devices', data),
  update: (id: string, data: Partial<Device>) => apiClient.patch(`/v1/devices/${id}`, data),
};
