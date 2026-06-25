import { useState, useRef, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { signaturesApi, type DocumentSignature } from '@/api/signatures';
import { PenLine, Plus, Check, X, Clock, FileSignature, Users } from 'lucide-react';

const statusColors: Record<string, string> = {
  pending: 'bg-yellow-500/20 text-yellow-400',
  signed: 'bg-green-500/20 text-green-400',
  declined: 'bg-red-500/20 text-red-400',
  expired: 'bg-gray-500/20 text-gray-400',
};

function SignaturePad({ onSave, onCancel }: { onSave: (data: string) => void; onCancel: () => void }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [drawing, setDrawing] = useState(false);

  const getPos = useCallback((e: React.MouseEvent<HTMLCanvasElement> | React.TouchEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current!;
    const rect = canvas.getBoundingClientRect();
    if ('touches' in e) {
      return { x: e.touches[0].clientX - rect.left, y: e.touches[0].clientY - rect.top };
    }
    return { x: e.clientX - rect.left, y: e.clientY - rect.top };
  }, []);

  const startDraw = useCallback((e: React.MouseEvent<HTMLCanvasElement> | React.TouchEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current!;
    const ctx = canvas.getContext('2d')!;
    const pos = getPos(e);
    ctx.beginPath();
    ctx.moveTo(pos.x, pos.y);
    setDrawing(true);
  }, [getPos]);

  const draw = useCallback((e: React.MouseEvent<HTMLCanvasElement> | React.TouchEvent<HTMLCanvasElement>) => {
    if (!drawing) return;
    const canvas = canvasRef.current!;
    const ctx = canvas.getContext('2d')!;
    const pos = getPos(e);
    ctx.lineTo(pos.x, pos.y);
    ctx.strokeStyle = '#e5e7eb';
    ctx.lineWidth = 2;
    ctx.lineCap = 'round';
    ctx.stroke();
  }, [drawing, getPos]);

  const stopDraw = useCallback(() => setDrawing(false), []);

  const clear = useCallback(() => {
    const canvas = canvasRef.current!;
    const ctx = canvas.getContext('2d')!;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
  }, []);

  const save = useCallback(() => {
    const canvas = canvasRef.current!;
    onSave(canvas.toDataURL('image/png'));
  }, [onSave]);

  return (
    <div className="space-y-3">
      <canvas
        ref={canvasRef}
        width={400}
        height={150}
        className="w-full cursor-crosshair rounded-lg border border-gray-700 bg-gray-800"
        onMouseDown={startDraw}
        onMouseMove={draw}
        onMouseUp={stopDraw}
        onMouseLeave={stopDraw}
        onTouchStart={startDraw}
        onTouchMove={draw}
        onTouchEnd={stopDraw}
      />
      <div className="flex justify-between">
        <div className="flex gap-2">
          <button onClick={clear} className="rounded-lg px-3 py-1.5 text-xs text-gray-400 hover:text-gray-200">Clear</button>
          <button onClick={onCancel} className="rounded-lg px-3 py-1.5 text-xs text-gray-400 hover:text-gray-200">Cancel</button>
        </div>
        <button
          onClick={save}
          className="flex items-center gap-1 rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-700"
        >
          <Check className="h-3 w-3" />
          Save Signature
        </button>
      </div>
    </div>
  );
}

export function SignaturesPage() {
  const queryClient = useQueryClient();
  const [showRequest, setShowRequest] = useState(false);
  const [showDecline, setShowDecline] = useState<string | null>(null);
  const [signingId, setSigningId] = useState<string | null>(null);
  const [declineReason, setDeclineReason] = useState('');
  const [docId, setDocId] = useState('');
  const [signers, setSigners] = useState<{ signer_id: string; signer_name: string }[]>([{ signer_id: '', signer_name: '' }]);
  const [lookupDocId, setLookupDocId] = useState('');

  const { data: statusData } = useQuery({
    queryKey: ['signature-status', lookupDocId],
    queryFn: () => signaturesApi.getStatus(lookupDocId),
    enabled: !!lookupDocId,
  });

  const signMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: string }) => signaturesApi.sign(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['signature-status'] });
      setSigningId(null);
    },
  });

  const declineMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => signaturesApi.decline(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['signature-status'] });
      setShowDecline(null);
      setDeclineReason('');
    },
  });

  const requestMutation = useMutation({
    mutationFn: () => signaturesApi.request(docId, signers.filter((s) => s.signer_id.trim())),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['signature-status'] });
      setShowRequest(false);
      setDocId('');
      setSigners([{ signer_id: '', signer_name: '' }]);
    },
  });

  const signatures: DocumentSignature[] = statusData?.data?.signatures || [];

  function addSigner() {
    setSigners([...signers, { signer_id: '', signer_name: '' }]);
  }

  function updateSigner(index: number, field: string, value: string) {
    const updated = [...signers];
    (updated[index] as Record<string, string>)[field] = value;
    setSigners(updated);
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <PenLine className="h-6 w-6 text-orange-400" />
          <h1 className="text-2xl font-bold">E-Signatures</h1>
        </div>
        <button
          onClick={() => setShowRequest(true)}
          className="flex items-center gap-2 rounded-lg bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700"
        >
          <Plus className="h-4 w-4" />
          Request Signature
        </button>
      </div>

      <div className="flex items-center gap-3">
        <FileSignature className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Enter document ID to view signature status..."
          value={lookupDocId}
          onChange={(e) => setLookupDocId(e.target.value)}
          className="w-80 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
      </div>

      {!lookupDocId ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <FileSignature className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">Enter a document ID to view signatures</p>
        </div>
      ) : signatures.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Users className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No signatures found for this document</p>
        </div>
      ) : (
        <div className="space-y-3">
          {signatures.map((sig) => (
            <div key={sig.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium text-gray-100">{sig.signer_name}</h3>
                  <div className="mt-1 flex items-center gap-2 text-xs text-gray-500">
                    <span className={`rounded-full px-2 py-0.5 ${statusColors[sig.status] || ''}`}>{sig.status}</span>
                    <span>Signer: {sig.signer_id}</span>
                    {sig.signed_at && <span>Signed: {new Date(sig.signed_at).toLocaleString()}</span>}
                    {sig.expires_at && <span>Expires: {new Date(sig.expires_at).toLocaleDateString()}</span>}
                  </div>
                  {sig.decline_reason && (
                    <p className="mt-1 text-xs text-red-400">Declined: {sig.decline_reason}</p>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {sig.status === 'pending' && (
                    <>
                      <button
                        onClick={() => setSigningId(sig.id)}
                        className="flex items-center gap-2 rounded-lg bg-green-600/10 px-3 py-2 text-sm text-green-400 hover:bg-green-600/20"
                      >
                        <PenLine className="h-4 w-4" />
                        Sign
                      </button>
                      <button
                        onClick={() => setShowDecline(sig.id)}
                        className="flex items-center gap-2 rounded-lg bg-red-600/10 px-3 py-2 text-sm text-red-400 hover:bg-red-600/20"
                      >
                        <X className="h-4 w-4" />
                        Decline
                      </button>
                    </>
                  )}
                  {sig.status === 'signed' && (
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-green-500/20">
                      <Check className="h-4 w-4 text-green-400" />
                    </div>
                  )}
                  {sig.status === 'pending' && (
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-yellow-500/20">
                      <Clock className="h-4 w-4 text-yellow-400" />
                    </div>
                  )}
                </div>
              </div>

              {signingId === sig.id && (
                <div className="mt-4 border-t border-gray-800 pt-4">
                  <p className="mb-2 text-sm text-gray-400">Draw your signature below:</p>
                  <SignaturePad
                    onSave={(data) => signMutation.mutate({ id: sig.id, data })}
                    onCancel={() => setSigningId(null)}
                  />
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showDecline && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Decline Signature</h2>
            <textarea
              placeholder="Reason for declining"
              value={declineReason}
              onChange={(e) => setDeclineReason(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
              autoFocus
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => { setShowDecline(null); setDeclineReason(''); }} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => declineMutation.mutate({ id: showDecline, reason: declineReason })}
                disabled={!declineReason.trim()}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                Decline
              </button>
            </div>
          </div>
        </div>
      )}

      {showRequest && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="max-h-[80vh] w-full max-w-md overflow-y-auto rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Request Signature</h2>
            <input
              type="text"
              placeholder="Document ID"
              value={docId}
              onChange={(e) => setDocId(e.target.value)}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <div className="mb-4">
              <div className="mb-2 flex items-center justify-between">
                <label className="text-sm font-medium text-gray-300">Signers</label>
                <button onClick={addSigner} className="text-xs text-blue-400 hover:text-blue-300">+ Add Signer</button>
              </div>
              {signers.map((s, i) => (
                <div key={i} className="mb-2 flex gap-2">
                  <input
                    type="text"
                    placeholder="Signer ID"
                    value={s.signer_id}
                    onChange={(e) => updateSigner(i, 'signer_id', e.target.value)}
                    className="flex-1 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                  <input
                    type="text"
                    placeholder="Signer Name"
                    value={s.signer_name}
                    onChange={(e) => updateSigner(i, 'signer_name', e.target.value)}
                    className="flex-1 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                  />
                </div>
              ))}
            </div>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowRequest(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => requestMutation.mutate()}
                disabled={!docId.trim() || signers.every((s) => !s.signer_id.trim())}
                className="rounded-lg bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
              >
                Request
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
