import { describe, it, expect } from 'vitest'
import { getCanonicalNavPath, ROUTES, ROUTE_BUILDERS } from './constants'

describe('getCanonicalNavPath', () => {
  it('returns exact path when pathname is a main nav path', () => {
    expect(getCanonicalNavPath('/dashboard')).toBe('/dashboard')
    expect(getCanonicalNavPath(ROUTES.FUNCTIONS)).toBe(ROUTES.FUNCTIONS)
    expect(getCanonicalNavPath(ROUTES.SETTINGS)).toBe(ROUTES.SETTINGS)
  })

  it('returns parent path when pathname is under a main nav path (not dashboard)', () => {
    expect(getCanonicalNavPath('/functions/123')).toBe('/functions')
    expect(getCanonicalNavPath('/admin/tenants/abc')).toBe('/admin/tenants')
  })

  it('returns null for unknown or non-main paths', () => {
    expect(getCanonicalNavPath('/')).toBeNull()
    expect(getCanonicalNavPath('/login')).toBeNull()
    expect(getCanonicalNavPath('/some/random/path')).toBeNull()
  })
})

describe('ROUTE_BUILDERS', () => {
  it('function builds /fx/:author/:name', () => {
    expect(ROUTE_BUILDERS.function('acme', 'my-fn')).toBe('/fx/acme/my-fn')
  })

  it('functionWithVersion builds with version', () => {
    expect(ROUTE_BUILDERS.functionWithVersion('acme', 'my-fn', '1.0.0')).toBe(
      '/fx/acme/my-fn@1.0.0'
    )
  })
})
