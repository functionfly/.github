/**
 * Runs axe-core accessibility checks in development (logs to console).
 * Only runs when import.meta.env.DEV is true.
 * Optional: re-enable in App.tsx and run `bun install` from repo root to use.
 */
import { useEffect } from 'react'
import React from 'react'
import ReactDOM from 'react-dom'

export function DevA11y() {
  useEffect(() => {
    if (!import.meta.env.DEV) return
    import('@axe-core/react')
      .then((mod) => {
        const init = (mod as { default?: (r: unknown, d: unknown, delay: number) => void }).default
        if (typeof init === 'function') init(React, ReactDOM, 1000)
      })
      .catch(() => {})
  }, [])
  return null
}
