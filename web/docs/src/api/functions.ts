import type { FunctionDocs, FunctionDocSummary, FunctionVersion } from '../types/function'

const API_BASE = '/docs'

export async function getFunctions(): Promise<FunctionDocSummary[]> {
  const response = await fetch(API_BASE)
  if (!response.ok) {
    throw new Error('Failed to fetch functions')
  }
  return response.json()
}

export async function getFunctionDocs(
  author: string,
  name: string,
  version?: string
): Promise<FunctionDocs> {
  const url = version
    ? `${API_BASE}/${author}/${name}/api?version=${version}`
    : `${API_BASE}/${author}/${name}/api`
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error('Failed to fetch function docs')
  }
  return response.json()
}

export async function getFunctionVersions(
  author: string,
  name: string
): Promise<FunctionVersion[]> {
  const response = await fetch(`${API_BASE}/${author}/${name}/versions`)
  if (!response.ok) {
    throw new Error('Failed to fetch function versions')
  }
  return response.json()
}

export async function getCategories(): Promise<Record<string, FunctionDocSummary[]>> {
  const functions = await getFunctions()
  const categories: Record<string, FunctionDocSummary[]> = {}

  for (const fn of functions) {
    const cat = fn.category || 'Uncategorized'
    if (!categories[cat]) {
      categories[cat] = []
    }
    categories[cat].push(fn)
  }

  return categories
}
