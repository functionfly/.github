import { mergeCorsOrigins } from "../cors-origins";

export default () => {
  // SECURITY: Require JWT_SECRET in production
  const jwtSecret = process.env.JWT_SECRET;
  if (!jwtSecret && process.env.NODE_ENV === 'production') {
    throw new Error('JWT_SECRET environment variable is required in production');
  }

  return {
    port: parseInt(process.env.PORT || "3000", 10),
    database: {
      url:
        process.env.DATABASE_URL ||
        "postgresql://postgres:postgres@localhost:5432/functionfly_blog",
    },
    jwt: {
      secret: jwtSecret || "dev-only-secret-do-not-use-in-production",
      expiresIn: process.env.JWT_EXPIRES_IN || "7d",
    },
    cors: {
      origin: mergeCorsOrigins(process.env.CORS_ORIGIN),
    },
    rateLimit: {
      ttl: parseInt(process.env.RATE_LIMIT_TTL || "60000", 10),
      limit: parseInt(process.env.RATE_LIMIT_LIMIT || "100", 10),
    },
  };
};
