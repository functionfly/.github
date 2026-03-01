#!/usr/bin/env node
/**
 * Creates the blog database if it doesn't exist.
 * Uses DATABASE_URL from env or default (localhost:5433/functionfly_blog).
 * Run: node scripts/create-db.mjs   or  bun run db:create
 */
import { Client } from 'pg';

const url = process.env.DATABASE_URL || 'postgresql://postgres:postgres@localhost:5433/functionfly_blog';
const parsed = new URL(url);
const dbName = parsed.pathname.slice(1) || 'functionfly_blog';
const baseUrl = url.replace(parsed.pathname, '/postgres');

const client = new Client({ connectionString: baseUrl });
try {
  await client.connect();
  const res = await client.query(
    'SELECT 1 FROM pg_database WHERE datname = $1',
    [dbName]
  );
  if (res.rows.length > 0) {
    console.log(`Database "${dbName}" already exists.`);
  } else {
    await client.query(`CREATE DATABASE "${dbName}"`);
    console.log(`Created database "${dbName}".`);
  }
} catch (err) {
  console.error(err.message);
  process.exit(1);
} finally {
  await client.end();
}
