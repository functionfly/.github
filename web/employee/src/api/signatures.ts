import apiClient from './client';

export interface DocumentSignature {
  id: string;
  document_id: string;
  signer_id: string;
  signer_name: string;
  signature_data?: string;
  signed_at?: string;
  status: string;
  decline_reason?: string;
  expires_at?: string;
  created_at: string;
}

export const signaturesApi = {
  request: (documentId: string, signers: { signer_id: string; signer_name: string }[]) => apiClient.post('/v1/signatures/request', { document_id: documentId, signers }),
  sign: (id: string, signatureData: string) => apiClient.post(`/v1/signatures/${id}/sign`, { signature_data: signatureData }),
  decline: (id: string, reason: string) => apiClient.post(`/v1/signatures/${id}/decline`, { reason }),
  getStatus: (documentId: string) => apiClient.get<{ signatures: DocumentSignature[] }>(`/v1/signatures/document/${documentId}`),
};
