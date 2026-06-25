import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { certificatesApi, type EmployeeCertificate } from '@/api/certificates';
import { KeyRound, Plus, Ban, Shield, Search } from 'lucide-react';

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  expired: 'bg-gray-500/20 text-gray-400',
  revoked: 'bg-red-500/20 text-red-400',
};

const typeColors: Record<string, string> = {
  device: 'bg-blue-500/20 text-blue-400',
  vpn: 'bg-purple-500/20 text-purple-400',
  email: 'bg-cyan-500/20 text-cyan-400',
  code_signing: 'bg-yellow-500/20 text-yellow-400',
};

export function CertificatesPage() {
  const queryClient = useQueryClient();
  const [showIssue, setShowIssue] = useState(false);
  const [showRevoke, setShowRevoke] = useState<string | null>(null);
  const [revokeReason, setRevokeReason] = useState('');
  const [search, setSearch] = useState('');
  const [form, setForm] = useState({ employee_id: '', certificate_type: 'device', device_id: '', device_name: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['certificates'],
    queryFn: () => certificatesApi.list(),
  });

  const issueMutation = useMutation({
    mutationFn: (data: { employee_id: string; certificate_type?: string; device_id?: string; device_name?: string }) => certificatesApi.issue(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificates'] });
      setShowIssue(false);
      setForm({ employee_id: '', certificate_type: 'device', device_id: '', device_name: '' });
    },
  });

  const revokeMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => certificatesApi.revoke(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['certificates'] });
      setShowRevoke(null);
      setRevokeReason('');
    },
  });

  const certs = (data?.data?.certificates || []).filter((c) =>
    !search ||
    c.subject.toLowerCase().includes(search.toLowerCase()) ||
    c.certificate_serial.toLowerCase().includes(search.toLowerCase()) ||
    (c.device_name && c.device_name.toLowerCase().includes(search.toLowerCase()))
  );

  const activeCount = (data?.data?.certificates || []).filter((c) => c.status === 'active').length;
  const expiredCount = (data?.data?.certificates || []).filter((c) => c.status === 'expired').length;
  const revokedCount = (data?.data?.certificates || []).filter((c) => c.status === 'revoked').length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <KeyRound className="h-6 w-6 text-green-400" />
          <h1 className="text-2xl font-bold">FF-CERT Management</h1>
        </div>
        <button
          onClick={() => setShowIssue(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Issue Certificate
        </button>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Shield className="h-4 w-4 text-green-400" />
            <span className="text-sm">Active</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-green-400">{activeCount}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Shield className="h-4 w-4 text-gray-400" />
            <span className="text-sm">Expired</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-gray-400">{expiredCount}</p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <div className="flex items-center gap-2 text-gray-400">
            <Ban className="h-4 w-4 text-red-400" />
            <span className="text-sm">Revoked</span>
          </div>
          <p className="mt-1 text-2xl font-bold text-red-400">{revokedCount}</p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search by subject, serial, or device..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-80 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : certs.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <KeyRound className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">{search ? 'No certificates match your search' : 'No certificates issued'}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {certs.map((cert) => (
            <div key={cert.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${cert.status === 'active' ? 'bg-green-500/10' : cert.status === 'revoked' ? 'bg-red-500/10' : 'bg-gray-800'}`}>
                    <KeyRound className={`h-5 w-5 ${cert.status === 'active' ? 'text-green-400' : cert.status === 'revoked' ? 'text-red-400' : 'text-gray-500'}`} />
                  </div>
                  <div>
                    <h3 className="font-medium text-gray-100">{cert.subject}</h3>
                    <div className="mt-1 flex items-center gap-2 text-xs text-gray-500">
                      <span className={`rounded-full px-2 py-0.5 ${statusColors[cert.status] || ''}`}>{cert.status}</span>
                      <span className={`rounded-full px-2 py-0.5 ${typeColors[cert.certificate_type] || 'bg-gray-500/20 text-gray-400'}`}>{cert.certificate_type}</span>
                      <code className="rounded bg-gray-800 px-1.5 py-0.5 text-gray-400">{cert.certificate_serial}</code>
                    </div>
                    <div className="mt-1 flex items-center gap-3 text-xs text-gray-500">
                      <span>Issuer: {cert.issuer}</span>
                      {cert.device_name && <span>Device: {cert.device_name}</span>}
                      <span>Issued: {new Date(cert.issued_at).toLocaleDateString()}</span>
                      <span>Expires: {new Date(cert.expires_at).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
                {cert.status === 'active' && (
                  <button
                    onClick={() => setShowRevoke(cert.id)}
                    className="flex items-center gap-2 rounded-lg bg-red-600/10 px-3 py-2 text-sm text-red-400 hover:bg-red-600/20"
                  >
                    <Ban className="h-4 w-4" />
                    Revoke
                  </button>
                )}
              </div>
              {cert.revoked_at && (
                <div className="mt-2 rounded-lg bg-red-500/5 px-3 py-2 text-xs text-red-400">
                  Revoked on {new Date(cert.revoked_at).toLocaleString()}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showIssue && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Issue Certificate</h2>
            <input
              type="text"
              placeholder="Employee ID"
              value={form.employee_id}
              onChange={(e) => setForm({ ...form, employee_id: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <select
              value={form.certificate_type}
              onChange={(e) => setForm({ ...form, certificate_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="device">Device</option>
              <option value="vpn">VPN</option>
              <option value="email">Email</option>
              <option value="code_signing">Code Signing</option>
            </select>
            <input
              type="text"
              placeholder="Device ID (optional)"
              value={form.device_id}
              onChange={(e) => setForm({ ...form, device_id: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <input
              type="text"
              placeholder="Device name (optional)"
              value={form.device_name}
              onChange={(e) => setForm({ ...form, device_name: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowIssue(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => issueMutation.mutate({
                  employee_id: form.employee_id,
                  certificate_type: form.certificate_type,
                  device_id: form.device_id || undefined,
                  device_name: form.device_name || undefined,
                })}
                disabled={!form.employee_id.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Issue
              </button>
            </div>
          </div>
        </div>
      )}

      {showRevoke && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Revoke Certificate</h2>
            <p className="mb-3 text-sm text-gray-400">This action cannot be undone. The certificate will be immediately invalidated.</p>
            <textarea
              placeholder="Reason for revocation"
              value={revokeReason}
              onChange={(e) => setRevokeReason(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
              autoFocus
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => { setShowRevoke(null); setRevokeReason(''); }} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => revokeMutation.mutate({ id: showRevoke, reason: revokeReason })}
                disabled={!revokeReason.trim()}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                Revoke
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
