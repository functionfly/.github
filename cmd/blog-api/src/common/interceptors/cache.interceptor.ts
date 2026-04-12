import { Injectable, NestInterceptor, ExecutionContext, CallHandler } from '@nestjs/common';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';

/**
 * Cache headers interceptor for public blog endpoints.
 * Adds appropriate Cache-Control headers for CDN optimization.
 */
@Injectable()
export class PublicCacheInterceptor implements NestInterceptor {
  private readonly defaultMaxAge = 300; // 5 minutes for dynamic content
  private readonly staleWhileRevalidate = 600; // 10 minutes stale-while-revalidate

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const response = context.switchToHttp().getResponse();

    return next.handle().pipe(
      tap(() => {
        // Set Cache-Control for public blog endpoints
        response.setHeader(
          'Cache-Control',
          `public, max-age=${this.defaultMaxAge}, stale-while-revalidate=${this.staleWhileRevalidate}`
        );
        // Additional headers for CDN optimization
        response.setHeader('Vary', 'Accept-Encoding');
      }),
    );
  }
}

/**
 * Long-term cache interceptor for static/frequently accessed content.
 * Use for category lists, author lists, and single post pages.
 */
@Injectable()
export class LongTermCacheInterceptor implements NestInterceptor {
  private readonly maxAge = 3600; // 1 hour
  private readonly staleWhileRevalidate = 86400; // 24 hours

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const response = context.switchToHttp().getResponse();

    return next.handle().pipe(
      tap(() => {
        response.setHeader(
          'Cache-Control',
          `public, max-age=${this.maxAge}, stale-while-revalidate=${this.staleWhileRevalidate}`
        );
        response.setHeader('Vary', 'Accept-Encoding');
      }),
    );
  }
}
