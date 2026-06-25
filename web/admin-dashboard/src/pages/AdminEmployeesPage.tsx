import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { DataTable, type Column } from '@/components/ui/DataTable';
import { Dialog } from '@/components/ui/Dialog';
import { toast } from 'sonner';
import {
  Users,
  UserPlus,
  Shield,
  Copy,
  Check,
  Search,
  Building2,
  MapPin,
  Calendar,
  Key,
  ExternalLink,
} from 'lucide-react';

interface Employee {
  id: string;
  user_id: string;
  tenant_id: string;
  employee_number: string;
  ffid: string;
  department_id?: string;
  manager_id?: string;
  hire_date?: string;
  employment_type: string;
  clearance_level: string;
  work_location?: string;
  office_location?: string;
  timezone?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface CreateEmployeeForm {
  first_name: string;
  last_name: string;
  email: string;
  employment_type: string;
  clearance_level: string;
  work_location: string;
  department_id: string;
}

const CLEARANCE_LEVELS = [
  { value: 'standard', label: 'Standard' },
  { value: 'elevated', label: 'Elevated' },
  { value: 'confidential', label: 'Confidential' },
  { value: 'top_secret', label: 'Top Secret' },
];

const EMPLOYMENT_TYPES = [
  { value: 'full_time', label: 'Full Time' },
  { value: 'part_time', label: 'Part Time' },
  { value: 'contractor', label: 'Contractor' },
  { value: 'intern', label: 'Intern' },
];

const STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  on_leave: 'bg-yellow-500/20 text-yellow-400',
  terminated: 'bg-red-500/20 text-red-400',
  suspended: 'bg-orange-500/20 text-orange-400',
};

const CLEARANCE_COLORS: Record<string, string> = {
  standard: 'bg-gray-500/20 text-gray-400',
  elevated: 'bg-blue-500/20 text-blue-400',
  confidential: 'bg-yellow-500/20 text-yellow-400',
  top_secret: 'bg-red-500/20 text-red-400',
};

export function AdminEmployeesPage() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showAccess, setShowAccess] = useState<Employee | null>(null);
  const [accessToken, setAccessToken] = useState('');
  const [copied, setCopied] = useState(false);
  const [form, setForm] = useState<CreateEmployeeForm>({
    first_name: '',
    last_name: '',
    email: '',
    employment_type: 'full_time',
    clearance_level: 'standard',
    work_location: 'remote',
    department_id: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['admin-employees', search, statusFilter],
    queryFn: () =>
      adminApiClient.get<{ employees: Employee[]; total: number }>('/employees', {
        params: { search, status: statusFilter, limit: 100 },
      }),
  });

  const createMutation = useMutation({
    mutationFn: (form: CreateEmployeeForm) =>
      adminApiClient.post('/employees', form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-employees'] });
      toast.success('Employee created successfully');
      setShowCreate(false);
      resetForm();
    },
    onError: (err: any) => {
      toast.error(err?.message || 'Failed to create employee');
    },
  });

  const generateAccessMutation = useMutation({
    mutationFn: (employeeId: string) =>
      adminApiClient.post<{ token: string }>(`/employees/${employeeId}/generate-access`),
    onSuccess: (data) => {
      setAccessToken((data as any).token);
      toast.success('Access token generated');
    },
    onError: (err: any) => {
      toast.error(err?.message || 'Failed to generate access');
    },
  });

  const resetForm = () => {
    setForm({
      first_name: '',
      last_name: '',
      email: '',
      employment_type: 'full_time',
      clearance_level: 'standard',
      work_location: 'remote',
      department_id: '',
    });
  };

  const handleCreate = () => {
    if (!form.first_name || !form.last_name || !form.email) {
      toast.error('First name, last name, and email are required');
      return;
    }
    createMutation.mutate(form);
  };

  const handleGenerateAccess = (employee: Employee) => {
    setShowAccess(employee);
    setAccessToken('');
    generateAccessMutation.mutate(employee.id);
  };

  const handleCopyToken = () => {
    navigator.clipboard.writeText(accessToken);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
    toast.success('Token copied to clipboard');
  };

  const handleCopyEmployeePortalLink = () => {
    const link = `${window.location.origin.replace(':3002', ':3003')}/login`;
    navigator.clipboard.writeText(link);
    toast.success('Employee portal link copied');
  };

  const employees = (data as any)?.employees || [];
  const total = (data as any)?.total || 0;

  const columns: Column<Employee>[] = [
    {
      key: 'ffid',
      header: 'FFID',
      render: (emp) => (
        <div className="font-mono text-sm text-blue-400">{emp.ffid}</div>
      ),
    },
    {
      key: 'employee_number',
      header: 'Employee #',
      render: (emp) => (
        <span className="font-mono text-sm">{emp.employee_number}</span>
      ),
    },
    {
      key: 'employment_type',
      header: 'Type',
      render: (emp) => (
        <span className="text-sm capitalize">{emp.employment_type.replace('_', ' ')}</span>
      ),
    },
    {
      key: 'clearance_level',
      header: 'Clearance',
      render: (emp) => (
        <span className={`rounded-full px-2 py-0.5 text-xs ${CLEARANCE_COLORS[emp.clearance_level] || ''}`}>
          {emp.clearance_level}
        </span>
      ),
    },
    {
      key: 'work_location',
      header: 'Location',
      render: (emp) => (
        <div className="flex items-center gap-1 text-sm text-gray-400">
          <MapPin className="h-3 w-3" />
          {emp.work_location || '—'}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (emp) => (
        <span className={`rounded-full px-2 py-0.5 text-xs ${STATUS_COLORS[emp.status] || ''}`}>
          {emp.status}
        </span>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      render: (emp) => (
        <span className="text-sm text-gray-500">
          {new Date(emp.created_at).toLocaleDateString()}
        </span>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (emp) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => handleGenerateAccess(emp)}
            title="Generate portal access"
          >
            <Key className="h-3 w-3" />
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Employees</h1>
          <p className="text-sm text-gray-400">
            Manage employee records and generate portal access
          </p>
        </div>
        <div className="flex gap-3">
          <Button variant="outline" onClick={handleCopyEmployeePortalLink}>
            <ExternalLink className="mr-2 h-4 w-4" />
            Portal Link
          </Button>
          <Button onClick={() => setShowCreate(true)}>
            <UserPlus className="mr-2 h-4 w-4" />
            Add Employee
          </Button>
        </div>
      </div>

      <Card className="p-4">
        <div className="flex gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search by FFID, name, email..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-10 pr-4 text-sm text-gray-100 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
            />
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
          >
            <option value="">All Statuses</option>
            <option value="active">Active</option>
            <option value="on_leave">On Leave</option>
            <option value="terminated">Terminated</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>
      </Card>

      <Card className="p-0">
        <div className="p-4 border-b border-gray-800">
          <div className="flex items-center gap-2 text-sm text-gray-400">
            <Users className="h-4 w-4" />
            <span>{total} employees</span>
          </div>
        </div>
        <DataTable
          data={employees}
          columns={columns}
          loading={isLoading}
          emptyMessage="No employees found"
        />
      </Card>

      {/* Create Employee Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <div className="space-y-4 p-6">
          <h2 className="text-lg font-semibold text-gray-100">Add Employee</h2>
          <p className="text-sm text-gray-400">
            Create a new employee record and generate portal access
          </p>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm text-gray-400">First Name *</label>
              <Input
                value={form.first_name}
                onChange={(e) => setForm({ ...form, first_name: e.target.value })}
                placeholder="Alex"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm text-gray-400">Last Name *</label>
              <Input
                value={form.last_name}
                onChange={(e) => setForm({ ...form, last_name: e.target.value })}
                placeholder="Smith"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-400">Email *</label>
            <Input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="alex.smith@functionfly.com"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm text-gray-400">Employment Type</label>
              <Select
                value={form.employment_type}
                onValueChange={(v) => setForm({ ...form, employment_type: v })}
                options={EMPLOYMENT_TYPES}
              />
            </div>
            <div>
              <label className="mb-1 block text-sm text-gray-400">Clearance Level</label>
              <Select
                value={form.clearance_level}
                onValueChange={(v) => setForm({ ...form, clearance_level: v })}
                options={CLEARANCE_LEVELS}
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-400">Work Location</label>
            <Select
              value={form.work_location}
              onValueChange={(v) => setForm({ ...form, work_location: v })}
              options={[
                { value: 'remote', label: 'Remote' },
                { value: 'hybrid', label: 'Hybrid' },
                { value: 'onsite', label: 'Onsite' },
              ]}
            />
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <Button variant="outline" onClick={() => setShowCreate(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={createMutation.isPending}>
              {createMutation.isPending ? 'Creating...' : 'Create Employee'}
            </Button>
          </div>
        </div>
      </Dialog>

      {/* Access Token Dialog */}
      <Dialog open={!!showAccess} onOpenChange={() => setShowAccess(null)}>
        <div className="space-y-4 p-6">
          <h2 className="text-lg font-semibold text-gray-100">Portal Access</h2>
          <p className="text-sm text-gray-400">
            Generate a login token for{' '}
            <span className="font-medium text-gray-200">
              {showAccess?.ffid}
            </span>
          </p>

          {generateAccessMutation.isPending ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
            </div>
          ) : accessToken ? (
            <div className="space-y-3">
              <div>
                <label className="mb-1 block text-sm text-gray-400">Access Token</label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={accessToken}
                    readOnly
                    className="flex-1 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 font-mono text-sm text-gray-100"
                  />
                  <Button size="sm" onClick={handleCopyToken}>
                    {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                  </Button>
                </div>
              </div>

              <div className="rounded-lg bg-blue-500/10 p-3 text-sm text-blue-400">
                <p className="font-medium">How to use:</p>
                <ol className="mt-1 list-inside list-decimal space-y-1">
                  <li>Share this token with the employee</li>
                  <li>Employee visits the FWOS portal</li>
                  <li>Employee enters the token on the login page</li>
                  <li>Token grants temporary access to set their password</li>
                </ol>
              </div>

              <div className="flex items-center gap-2 text-sm text-gray-400">
                <ExternalLink className="h-4 w-4" />
                <span>Portal: </span>
                <code className="rounded bg-gray-800 px-2 py-0.5 text-xs">
                  {window.location.origin.replace(':3002', ':3003')}
                </code>
              </div>
            </div>
          ) : null}

          <div className="flex justify-end pt-4">
            <Button variant="outline" onClick={() => setShowAccess(null)}>
              Close
            </Button>
          </div>
        </div>
      </Dialog>
    </div>
  );
}
