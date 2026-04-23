import { useCallback, useSyncExternalStore } from 'react'
import { RTL_LANGUAGES } from '@/lib/i18n/languages'

// External store for dir attribute so multiple components subscribe without re-renders from context
let currentDir: 'ltr' | 'rtl' = 'ltr'
const dirListeners = new Set<() => void>()

function subscribeDir(listener: () => void) {
  dirListeners.add(listener)
  return () => dirListeners.delete(listener)
}

function getDirSnapshot() {
  return currentDir
}

function setDir(dir: 'ltr' | 'rtl') {
  if (currentDir !== dir) {
    currentDir = dir
    dirListeners.forEach((l) => l())
  }
}

/**
 * Hook that tracks and applies text direction based on language.
 * Listens to document.documentElement.dir changes via MutationObserver
 * and provides a reactive dir value.
 */
export function useDirection() {
  const dir = useSyncExternalStore(subscribeDir, getDirSnapshot)

  const isRtl = dir === 'rtl'

  const applyDir = useCallback((langCode: string) => {
    const newDir = RTL_LANGUAGES.has(langCode) ? 'rtl' : 'ltr'
    document.documentElement.dir = newDir
    document.documentElement.lang = langCode
    setDir(newDir)
  }, [])

  return { dir, isRtl, applyDir }
}
