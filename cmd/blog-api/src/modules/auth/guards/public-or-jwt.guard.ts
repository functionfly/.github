import { Injectable, ExecutionContext } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { AuthGuard } from '@nestjs/passport';

/**
 * When BLOG_PUBLIC_API=true (e.g. local dev), allows requests without a valid JWT
 * and injects a dev user so dashboard can call blog-api with the main app's token.
 * Otherwise delegates to JWT auth.
 */
@Injectable()
export class PublicOrJwtAuthGuard extends AuthGuard('jwt') {
  constructor(private configService: ConfigService) {
    super();
  }

  async canActivate(context: ExecutionContext): Promise<boolean> {
    if (this.configService.get<boolean>('BLOG_PUBLIC_API')) {
      const request = context.switchToHttp().getRequest();
      request.user = {
        userId: 'local-dev',
        email: 'dev@local',
        role: 'admin',
      };
      return true;
    }
    return (await super.canActivate(context)) as boolean;
  }
}
