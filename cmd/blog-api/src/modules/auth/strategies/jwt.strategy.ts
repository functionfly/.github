import { Injectable, UnauthorizedException } from '@nestjs/common';
import { PassportStrategy } from '@nestjs/passport';
import { ExtractJwt, Strategy } from 'passport-jwt';
import { ConfigService } from '@nestjs/config';

export interface JwtPayload {
  /** JWT subject — the user's UUID */
  sub: string;
  /** User's email address */
  email: string;
  /** Platform role (e.g. 'super_admin', 'admin', or '' for regular users) */
  role: string;
  /** Username (optional — only set if the user has chosen one) */
  username?: string;
  /** Tenant UUID the user belongs to */
  tenant_id?: string;
  /** Explicit permission strings granted to this user */
  permissions?: string[];
}

@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
  constructor(private readonly configService: ConfigService) {
    const secret = configService.get<string>('jwt.secret');
    if (!secret) {
      throw new Error(
        'JWT secret is not configured. Set the JWT_SECRET environment variable.',
      );
    }
    super({
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      ignoreExpiration: false,
      secretOrKey: secret,
    });
  }

  async validate(payload: JwtPayload) {
    if (!payload.sub) {
      throw new UnauthorizedException('Invalid token: missing subject');
    }

    return {
      userId: payload.sub,
      email: payload.email,
      role: payload.role,
      username: payload.username,
      tenantId: payload.tenant_id,
      permissions: payload.permissions ?? [],
    };
  }
}
