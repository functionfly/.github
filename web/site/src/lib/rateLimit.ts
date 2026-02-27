// Simple in-memory rate limiter
// In production, consider using Redis or a dedicated rate limiting service

interface RateLimitEntry {
  count: number
  resetTime: number
}

class RateLimiter {
  private store = new Map<string, RateLimitEntry>()
  private cleanupInterval: NodeJS.Timeout

  constructor(
    private windowMs: number = 15 * 60 * 1000, // 15 minutes
    private maxRequests: number = 100,
    private cleanupMs: number = 60 * 1000 // Clean up every minute
  ) {
    // Periodic cleanup of expired entries
    this.cleanupInterval = setInterval(() => {
      this.cleanup()
    }, cleanupMs)
  }

  check(key: string): { allowed: boolean; remaining: number; resetTime: number } {
    const now = Date.now()
    const entry = this.store.get(key)

    if (!entry || now > entry.resetTime) {
      // First request or window expired
      this.store.set(key, {
        count: 1,
        resetTime: now + this.windowMs,
      })
      return {
        allowed: true,
        remaining: this.maxRequests - 1,
        resetTime: now + this.windowMs,
      }
    }

    if (entry.count >= this.maxRequests) {
      return {
        allowed: false,
        remaining: 0,
        resetTime: entry.resetTime,
      }
    }

    entry.count++
    return {
      allowed: true,
      remaining: this.maxRequests - entry.count,
      resetTime: entry.resetTime,
    }
  }

  private cleanup() {
    const now = Date.now()
    for (const [key, entry] of this.store.entries()) {
      if (now > entry.resetTime) {
        this.store.delete(key)
      }
    }
  }

  destroy() {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval)
    }
  }
}

// Global rate limiter instance
const globalRateLimiter = new RateLimiter()

export function checkRateLimit(
  request: Request,
  options: { windowMs?: number; maxRequests?: number } = {}
): { allowed: boolean; remaining: number; resetTime: number; headers: Record<string, string> } {
  const clientIP = getClientIP(request)
  const key = `ratelimit:${clientIP}`

  const result = globalRateLimiter.check(key)

  const headers = {
    'X-RateLimit-Limit': options.maxRequests?.toString() || '100',
    'X-RateLimit-Remaining': result.remaining.toString(),
    'X-RateLimit-Reset': result.resetTime.toString(),
    'X-RateLimit-Window': (options.windowMs || 15 * 60 * 1000).toString(),
  }

  return {
    ...result,
    headers,
  }
}

function getClientIP(request: Request): string {
  // Try to get IP from various headers
  const forwardedFor = request.headers.get('x-forwarded-for')
  const realIP = request.headers.get('x-real-ip')
  const cfConnectingIP = request.headers.get('cf-connecting-ip')

  if (cfConnectingIP) return cfConnectingIP
  if (realIP) return realIP
  if (forwardedFor) return forwardedFor.split(',')[0].trim()

  // Fallback to a default key for server-side requests
  return 'unknown'
}

// Cleanup on process exit
process.on('SIGINT', () => globalRateLimiter.destroy())
process.on('SIGTERM', () => globalRateLimiter.destroy())