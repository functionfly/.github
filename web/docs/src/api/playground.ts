import type { PlaygroundResponse } from '../types/function'

const PLAYGROUND_BASE = '/playground'

export async function executeFunction(
  author: string,
  name: string,
  input: unknown
): Promise<PlaygroundResponse> {
  const response = await fetch(`${PLAYGROUND_BASE}/${author}/${name}/execute`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  })

  if (!response.ok) {
    throw new Error('Failed to execute function')
  }

  return response.json()
}

export async function sharePlayground(
  author: string,
  name: string,
  input: unknown
): Promise<{ share_url: string }> {
  const response = await fetch(`${PLAYGROUND_BASE}/${author}/${name}/share`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  })

  if (!response.ok) {
    throw new Error('Failed to share playground')
  }

  return response.json()
}
