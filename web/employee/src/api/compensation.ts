import apiClient from './client';

export interface CompensationRecord {
  id: string;
  employee_id: string;
  base_salary_cents: number;
  currency: string;
  pay_frequency: string;
  effective_date: string;
  end_date?: string;
  review_date?: string;
}

export interface EquityGrant {
  id: string;
  employee_id: string;
  grant_type: string;
  shares: number;
  strike_price_cents?: number;
  vesting_start: string;
  vesting_end: string;
  cliff_date?: string;
  vested_shares: number;
  status: string;
  grant_date: string;
}

export const compensationApi = {
  get: (employeeId: string) =>
    apiClient.get<{ compensation: CompensationRecord }>(`/v1/compensation/${employeeId}`),

  update: (employeeId: string, data: Partial<CompensationRecord>) =>
    apiClient.patch(`/v1/compensation/${employeeId}`, data),

  listEquity: (employeeId: string) =>
    apiClient.get<{ grants: EquityGrant[] }>(`/v1/compensation/${employeeId}/equity`),
};
