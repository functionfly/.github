import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { employeesApi, type Employee, type ListEmployeesOpts } from '@/api/employees';

export function useEmployees(opts?: ListEmployeesOpts) {
  return useQuery({
    queryKey: ['employees', opts],
    queryFn: () => employeesApi.list(opts),
  });
}

export function useEmployee(id: string) {
  return useQuery({
    queryKey: ['employee', id],
    queryFn: () => employeesApi.get(id),
    enabled: !!id,
  });
}

export function useEmployeeByFFID(ffid: string) {
  return useQuery({
    queryKey: ['employee', 'ffid', ffid],
    queryFn: () => employeesApi.getByFFID(ffid),
    enabled: !!ffid,
  });
}

export function useCreateEmployee() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Employee>) => employeesApi.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['employees'] }),
  });
}

export function useUpdateEmployee() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Employee> }) =>
      employeesApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      queryClient.invalidateQueries({ queryKey: ['employee', id] });
    },
  });
}
