import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { certificatesApi } from '@/api/certificates';
import { Lock, Key, Copy, Check, Shield, ChevronRight } from 'lucide-react';

export function CertificatePKIPage() {
  const queryClient = useQueryClient();
  const [copied, setCopied] = useState(false);
  const [selectedCert, setSelectedCert] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['certificates'],
    queryFn: () => certificatesApi.list(),
  });

  const { data: certData } = useQuery({
    queryKey: ['certificate', selectedCert],
    queryFn: () => certificatesApi.get(selectedCert!),
    enabled: !!selectedCert,
  });

  const generateMutation = useMutation({
    mutationFn: () => certificatesApi.issue({ employee_id: 'self', certificate_type: 'pki' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['certificates'] }),
  });

  const certs = data?.data?.certificates || [];
  const selectedCertData = certData?.data?.certificate;

  function copyPublicKey(text: string) {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Lock className="h-6 w-6 text-cyan-400" />
          <h1 className="text-2xl font-bold">Certificate PKI</h1>
        </div>
        <button
          onClick={() => generateMutation.mutate()}
          className="flex items-center gap-2 rounded-lg bg-cyan-600 px-4 py-2 text-sm font-medium text-white hover:bg-cyan-700"
        >
          <Key className="h-4 w-4" />
          Generate Key Pair
        </button>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-1">
          <h2 className="mb-3 text-sm font-medium text-gray-400">Certificates</h2>
          {isLoading ? (
            <div className="flex justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
            </div>
          ) : certs.length === 0 ? (
            <div className="rounded-xl border border-gray-800 bg-gray-900 py-8 text-center">
              <Shield className="mx-auto mb-2 h-8 w-8 text-gray-600" />
              <p className="text-sm text-gray-500">No certificates</p>
            </div>
          ) : (
            <div className="space-y-2">
              {certs.map((cert) => (
                <button
                  key={cert.id}
                  onClick={() => setSelectedCert(cert.id === selectedCert ? null : cert.id)}
                  className={`flex w-full items-center justify-between rounded-lg border p-3 text-left transition-colors ${
                    cert.id === selectedCert
                      ? 'border-cyan-600 bg-gray-800'
                      : 'border-gray-800 bg-gray-900 hover:bg-gray-800'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${cert.status === 'active' ? 'bg-green-500/10' : 'bg-gray-800'}`}>
                      <Shield className={`h-4 w-4 ${cert.status === 'active' ? 'text-green-400' : 'text-gray-500'}`} />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-100">{cert.subject}</p>
                      <p className="text-xs text-gray-500">{cert.certificate_type}</p>
                    </div>
                  </div>
                  <ChevronRight className={`h-4 w-4 text-gray-500 transition-transform ${cert.id === selectedCert ? 'rotate-90' : ''}`} />
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="lg:col-span-2">
          {selectedCertData ? (
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
              <h2 className="mb-4 text-lg font-semibold text-gray-100">Certificate Details</h2>

              <div className="grid grid-cols-2 gap-4">
                <div className="rounded-lg bg-gray-800 p-3">
                  <span className="text-xs text-gray-500">Subject</span>
                  <p className="mt-1 text-sm font-medium text-gray-100">{selectedCertData.subject}</p>
                </div>
                <div className="rounded-lg bg-gray-800 p-3">
                  <span className="text-xs text-gray-500">Issuer</span>
                  <p className="mt-1 text-sm font-medium text-gray-100">{selectedCertData.issuer}</p>
                </div>
                <div className="rounded-lg bg-gray-800 p-3">
                  <span className="text-xs text-gray-500">Serial</span>
                  <code className="mt-1 block text-sm text-cyan-400">{selectedCertData.certificate_serial}</code>
                </div>
                <div className="rounded-lg bg-gray-800 p-3">
                  <span className="text-xs text-gray-500">Status</span>
                  <p className={`mt-1 text-sm font-medium ${selectedCertData.status === 'active' ? 'text-green-400' : selectedCertData.status === 'revoked' ? 'text-red-400' : 'text-gray-400'}`}>{selectedCertData.status}</p>
                </div>
                <div className="rounded-lg bg-gray-800 p-3">
                  <span className="text-xs text-gray-500">Issued</span>
                  <p className="mt-1 text-sm text-gray-100">{new Date(selectedCertData.issued_at).toLocaleString()}</p>
                </div>
                <div className="rounded-lg bg-gray-800 p-3">
                  <span className="text-xs text-gray-500">Expires</span>
                  <p className="mt-1 text-sm text-gray-100">{new Date(selectedCertData.expires_at).toLocaleString()}</p>
                </div>
              </div>

              <div className="mt-4">
                <div className="mb-2 flex items-center justify-between">
                  <span className="text-sm text-gray-400">Public Key</span>
                  <button
                    onClick={() => copyPublicKey(`-----BEGIN PUBLIC KEY-----\n${selectedCertData.certificate_serial}\n-----END PUBLIC KEY-----`)}
                    className="flex items-center gap-1 text-xs text-cyan-400 hover:text-cyan-300"
                  >
                    {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                    {copied ? 'Copied' : 'Copy'}
                  </button>
                </div>
                <pre className="overflow-x-auto rounded-lg bg-gray-800 p-3 text-xs text-gray-300">
{`-----BEGIN PUBLIC KEY-----
${selectedCertData.certificate_serial}
-----END PUBLIC KEY-----`}
                </pre>
              </div>

              <div className="mt-4">
                <h3 className="mb-2 text-sm font-medium text-gray-400">Certificate Chain</h3>
                <div className="space-y-2">
                  <div className="flex items-center gap-3 rounded-lg bg-gray-800 p-3">
                    <div className="flex h-6 w-6 items-center justify-center rounded bg-cyan-500/20">
                      <Lock className="h-3 w-3 text-cyan-400" />
                    </div>
                    <div>
                      <p className="text-sm text-gray-100">Root CA</p>
                      <p className="text-xs text-gray-500">FunctionFly Internal CA</p>
                    </div>
                  </div>
                  <div className="ml-4 border-l border-gray-700 pl-4">
                    <div className="flex items-center gap-3 rounded-lg bg-gray-800 p-3">
                      <div className="flex h-6 w-6 items-center justify-center rounded bg-green-500/20">
                        <Shield className="h-3 w-3 text-green-400" />
                      </div>
                      <div>
                        <p className="text-sm text-gray-100">{selectedCertData.subject}</p>
                        <p className="text-xs text-gray-500">{selectedCertData.certificate_serial}</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-20">
              <Lock className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">Select a certificate to view details</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
