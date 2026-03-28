import { fetchSignupConfig } from '@/api/signup-config';
import { useQuery } from '@tanstack/react-query';

export function useSignupConfig() {
  return useQuery({
    queryKey: ['signup-config'],
    queryFn: fetchSignupConfig,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });
}
