import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '@/api/client';
import { walletApi, type WalletPass } from '@/api/wallet';
import { Wallet, QrCode, Apple, CheckCircle, Clock, Shield, XCircle } from 'lucide-react';
import { toast } from 'sonner';

const platformIcons: Record<string, typeof Apple> = {
  apple_wallet: Apple,
  google_wallet: Shield,
};

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  revoked: 'bg-red-500/20 text-red-400',
  expired: 'bg-gray-500/20 text-gray-400',
  pending: 'bg-yellow-500/20 text-yellow-400',
};

export function WalletBadgePage() {
  const queryClient = useQueryClient();
  const [qrToken, setQrToken] = useState<string | null>(null);
  const [qrExpiresAt, setQrExpiresAt] = useState<string | null>(null);
  const [timeLeft, setTimeLeft] = useState<string>('');

  const { data: passesData, isLoading } = useQuery({
    queryKey: ['wallet-passes'],
    queryFn: () => apiClient.get<{ passes: WalletPass[] }>('/v1/wallet/passes'),
  });

  const generateMutation = useMutation({
    mutationFn: (platform: 'apple_wallet' | 'google_wallet') => walletApi.generate(platform),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['wallet-passes'] });
      const pass = data.data.pass;
      setQrToken(pass.qr_token);
      setQrExpiresAt(pass.qr_expires_at);
      toast.success(`${pass.platform === 'apple_wallet' ? 'Apple' : 'Google'} Wallet pass generated`);
    },
    onError: () => toast.error('Failed to generate pass'),
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => walletApi.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wallet-passes'] });
      toast.success('Pass revoked');
    },
    onError: () => toast.error('Failed to revoke pass'),
  });

  useEffect(() => {
    if (!qrExpiresAt) return;
    const interval = setInterval(() => {
      const diff = new Date(qrExpiresAt).getTime() - Date.now();
      if (diff <= 0) {
        setTimeLeft('Expired');
        clearInterval(interval);
        return;
      }
      const mins = Math.floor(diff / 60000);
      const secs = Math.floor((diff % 60000) / 1000);
      setTimeLeft(`${mins}:${secs.toString().padStart(2, '0')}`);
    }, 1000);
    return () => clearInterval(interval);
  }, [qrExpiresAt]);

  const passes = passesData?.data?.passes || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Wallet className="h-6 w-6 text-emerald-400" />
        <h1 className="text-2xl font-bold">Wallet Badge</h1>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
          <h2 className="mb-4 text-lg font-semibold">Badge Preview</h2>
          <div className="rounded-xl border-2 border-blue-500/30 bg-gradient-to-br from-gray-800 to-gray-900 p-6">
            <div className="flex items-center gap-4">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-blue-600 text-2xl font-bold text-white">
                F
              </div>
              <div>
                <p className="text-xs uppercase tracking-wider text-blue-400">FunctionFly ID</p>
                <p className="text-xl font-bold">FWOS Badge</p>
              </div>
            </div>
            <div className="mt-6 grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-xs text-gray-500">FFID</span>
                <p className="font-mono font-medium">FF-20260001</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Name</span>
                <p className="font-medium">Employee</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Department</span>
                <p className="font-medium">Engineering</p>
              </div>
              <div>
                <span className="text-xs text-gray-500">Clearance</span>
                <p className="font-medium">Level 3</p>
              </div>
            </div>
            <div className="mt-4 flex items-center gap-2 text-xs text-gray-500">
              <Shield className="h-3 w-3" />
              <span>Verified Digital Credential</span>
            </div>
          </div>

          <div className="mt-6 flex gap-3">
            <button
              onClick={() => generateMutation.mutate('apple_wallet')}
              disabled={generateMutation.isPending}
              className="flex flex-1 items-center justify-center gap-2 rounded-lg bg-gray-800 px-4 py-3 text-sm font-medium text-white hover:bg-gray-700 disabled:opacity-50"
            >
              <Apple className="h-4 w-4" />
              Apple Wallet
            </button>
            <button
              onClick={() => generateMutation.mutate('google_wallet')}
              disabled={generateMutation.isPending}
              className="flex flex-1 items-center justify-center gap-2 rounded-lg bg-gray-800 px-4 py-3 text-sm font-medium text-white hover:bg-gray-700 disabled:opacity-50"
            >
              <Shield className="h-4 w-4" />
              Google Wallet
            </button>
          </div>
        </div>

        <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
          <h2 className="mb-4 text-lg font-semibold">QR Code</h2>
          {qrToken ? (
            <div className="flex flex-col items-center">
              <div className="flex h-48 w-48 items-center justify-center rounded-xl border-2 border-dashed border-gray-700 bg-gray-800">
                <div className="text-center">
                  <QrCode className="mx-auto h-16 w-16 text-blue-400" />
                  <p className="mt-2 font-mono text-xs text-gray-400">{qrToken.slice(0, 16)}...</p>
                </div>
              </div>
              <div className="mt-4 flex items-center gap-2 text-sm">
                <Clock className="h-4 w-4 text-yellow-400" />
                <span className={timeLeft === 'Expired' ? 'text-red-400' : 'text-gray-300'}>
                  {timeLeft === 'Expired' ? 'QR Code Expired' : `Expires in ${timeLeft}`}
                </span>
              </div>
              <button
                onClick={() => {
                  setQrToken(null);
                  setQrExpiresAt(null);
                }}
                className="mt-3 text-xs text-gray-500 hover:text-gray-300"
              >
                Dismiss
              </button>
            </div>
          ) : (
            <div className="flex h-48 items-center justify-center rounded-xl border border-dashed border-gray-700 text-center text-sm text-gray-500">
              Generate a wallet pass to display QR code
            </div>
          )}
        </div>
      </div>

      <div>
        <h2 className="mb-3 text-lg font-semibold">Active Passes</h2>
        {isLoading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
          </div>
        ) : passes.length === 0 ? (
          <div className="rounded-xl border border-gray-800 bg-gray-900 py-12 text-center text-gray-500">
            No wallet passes generated yet
          </div>
        ) : (
          <div className="space-y-3">
            {passes.map((pass) => {
              const PlatformIcon = platformIcons[pass.platform] || Wallet;
              return (
                <div
                  key={pass.id}
                  className="flex items-center justify-between rounded-xl border border-gray-800 bg-gray-900 p-4"
                >
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg bg-gray-800 p-2">
                      <PlatformIcon className="h-5 w-5 text-gray-400" />
                    </div>
                    <div>
                      <p className="font-medium capitalize">{pass.platform.replace('_', ' ')}</p>
                      <p className="text-xs text-gray-500">
                        {pass.pass_type} &middot; Created {new Date(pass.installed_at || pass.qr_expires_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColors[pass.status] || 'bg-gray-500/20 text-gray-400'}`}
                    >
                      {pass.status === 'active' ? <CheckCircle className="h-3 w-3" /> : <XCircle className="h-3 w-3" />}
                      {pass.status}
                    </span>
                    {pass.status === 'active' && (
                      <button
                        onClick={() => revokeMutation.mutate(pass.id)}
                        disabled={revokeMutation.isPending}
                        className="rounded-lg border border-red-800 px-3 py-1.5 text-xs text-red-400 hover:bg-red-900/20 disabled:opacity-50"
                      >
                        Revoke
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
