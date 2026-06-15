/**
 * AI service client (FlyMind / ai-service).
 * Uses VITE_AI_SERVICE_URL when set; used by direct AI completion features in the dashboard.
 */

import { getAiServiceBaseUrl } from '@/lib/constants';
import { tokenVault } from '@/utils/token-vault';

export type ChatRole = 'system' | 'user' | 'assistant';

export interface ChatMessageInput {
  role: ChatRole;
  content: string;
}

export interface CompletionRequest {
  messages: ChatMessageInput[];
  provider?: string;
  model?: string;
  temperature?: number;
  max_tokens?: number;
  stream?: boolean;
}

export interface CompletionResponse {
  content: string;
  provider: string;
  model: string;
  usage?: { prompt_tokens?: number; completion_tokens?: number; total_tokens?: number };
  finish_reason?: string;
  latency_ms?: number;
}

/**
 * Call the ai-service completion endpoint. Returns the assistant reply text.
 * Throws on network error or non-2xx response.
 */
export async function complete(request: CompletionRequest): Promise<CompletionResponse> {
  const base = getAiServiceBaseUrl();
  if (!base) {
    throw new Error('AI service URL not configured (VITE_AI_SERVICE_URL)');
  }

  const url = `${base}/api/complete`;
  const body = {
    messages: request.messages.map((m) => ({ role: m.role, content: m.content })),
    ...(request.provider != null && { provider: request.provider }),
    ...(request.model != null && { model: request.model }),
    ...(request.temperature != null && { temperature: request.temperature }),
    ...(request.max_tokens != null && { max_tokens: request.max_tokens }),
    stream: false,
  };

  await tokenVault.initialize();
  const token = await tokenVault.getAccessToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(
      res.status === 502 || res.status === 503
        ? 'AI service is temporarily unavailable.'
        : res.status === 401
        ? 'Authentication required.'
        : `AI request failed: ${text}`
    );
  }

  const data = (await res.json()) as CompletionResponse;
  return data;
}
