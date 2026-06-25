import { useState } from 'react';
import { User, Briefcase, FunctionSquare, ShieldCheck } from 'lucide-react';
import type { Employee } from '@/api/employees';

interface IdentityLayersProps {
  employee: Employee;
  className?: string;
}

const tabs = [
  { id: 'human', label: 'Human', icon: User },
  { id: 'professional', label: 'Professional', icon: Briefcase },
  { id: 'function', label: 'Function', icon: FunctionSquare },
  { id: 'trust', label: 'Trust', icon: ShieldCheck },
] as const;

type TabId = (typeof tabs)[number]['id'];

function HumanLayer({ employee }: { employee: Employee }) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        {[
          { label: 'Pronouns', value: employee.pronouns },
          { label: 'Timezone', value: employee.timezone },
          { label: 'Location', value: employee.work_location },
          { label: 'Office', value: employee.office_location },
        ].map((item) => (
          <div key={item.label} className="rounded-lg bg-gray-800/50 p-3">
            <span className="text-xs text-gray-500">{item.label}</span>
            <p className="mt-1 text-sm text-gray-200">{item.value || '—'}</p>
          </div>
        ))}
      </div>
      {employee.bio && (
        <div className="rounded-lg bg-gray-800/50 p-4">
          <span className="text-xs text-gray-500">Bio</span>
          <p className="mt-1 text-sm text-gray-300">{employee.bio}</p>
        </div>
      )}
    </div>
  );
}

function ProfessionalLayer({ employee }: { employee: Employee }) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        {[
          { label: 'Employment Type', value: employee.employment_type.replace('_', ' ') },
          { label: 'Employee Number', value: employee.employee_number },
          { label: 'Department ID', value: employee.department_id?.toString() },
          { label: 'Manager ID', value: employee.manager_id },
          { label: 'Hire Date', value: employee.hire_date ? new Date(employee.hire_date).toLocaleDateString() : undefined },
        ].map((item) => (
          <div key={item.label} className="rounded-lg bg-gray-800/50 p-3">
            <span className="text-xs text-gray-500">{item.label}</span>
            <p className="mt-1 text-sm text-gray-200">{item.value || '—'}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function FunctionLayer({ employee }: { employee: Employee }) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        {[
          { label: 'FFID', value: employee.ffid },
          { label: 'Tenant ID', value: employee.tenant_id },
          { label: 'User ID', value: employee.user_id },
        ].map((item) => (
          <div key={item.label} className="rounded-lg bg-gray-800/50 p-3">
            <span className="text-xs text-gray-500">{item.label}</span>
            <p className="mt-1 font-mono text-sm text-gray-200">{item.value || '—'}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function TrustLayer({ employee }: { employee: Employee }) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
        {[
          { label: 'Clearance Level', value: employee.clearance_level },
          { label: 'Status', value: employee.status },
          { label: 'Created', value: employee.created_at ? new Date(employee.created_at).toLocaleDateString() : undefined },
          { label: 'Last Updated', value: employee.updated_at ? new Date(employee.updated_at).toLocaleDateString() : undefined },
        ].map((item) => (
          <div key={item.label} className="rounded-lg bg-gray-800/50 p-3">
            <span className="text-xs text-gray-500">{item.label}</span>
            <p className="mt-1 text-sm text-gray-200">{item.value || '—'}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

const layerComponents: Record<TabId, React.ComponentType<{ employee: Employee }>> = {
  human: HumanLayer,
  professional: ProfessionalLayer,
  function: FunctionLayer,
  trust: TrustLayer,
};

export function IdentityLayers({ employee, className = '' }: IdentityLayersProps) {
  const [activeTab, setActiveTab] = useState<TabId>('human');
  const LayerComponent = layerComponents[activeTab];

  return (
    <div className={`rounded-xl border border-gray-800 bg-gray-900 ${className}`}>
      <div className="flex border-b border-gray-800">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'border-b-2 border-blue-500 text-blue-400'
                : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            <tab.icon className="h-4 w-4" />
            {tab.label}
          </button>
        ))}
      </div>
      <div className="p-5">
        <LayerComponent employee={employee} />
      </div>
    </div>
  );
}
