import { drizzle } from 'drizzle-orm/node-postgres';
import pg from 'pg';
import * as schema from './schema/index';
import { ConfigService } from '@nestjs/config';

const { Pool } = pg;

export const getDatabase = (configService: ConfigService) => {
  const databaseUrl = configService.get<string>('DATABASE_URL');

  const pool = new Pool({
    connectionString: databaseUrl,
    max: 20,
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 5000,
  });

  return drizzle(pool, { schema });
};

export type Database = ReturnType<typeof getDatabase>;
export { schema };
