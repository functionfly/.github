/**
 * @fileoverview Smoke tests for the SDK type definitions. The
 * TypeScript compiler is the source of truth for type correctness;
 * this file only exercises runtime behaviour that doesn't require
 * the build step.
 */
import test from "node:test";
import assert from "node:assert/strict";
import { VaultAPIError } from "../errors";
import { SecretType } from "../types/secrets";
import { DynamicDBType } from "../types/dynamic";

test("VaultAPIError surfaces status and code", () => {
  const err = new VaultAPIError(403, "FORBIDDEN", "forbidden", "nope", { x: 1 });
  assert.equal(err.status, 403);
  assert.equal(err.code, "FORBIDDEN");
  assert.equal(err.message, "FORBIDDEN: nope");
  assert.deepEqual(err.details, { x: 1 });
});

test("SecretType enum is exhaustive", () => {
  const values = new Set<string>(Object.values(SecretType) as string[]);
  assert.equal(values.size, 4);
  assert.ok(values.has(SecretType.APIKey));
  assert.ok(values.has(SecretType.OAuthToken));
  assert.ok(values.has(SecretType.Password));
  assert.ok(values.has(SecretType.Certificate));
});

test("DynamicDBType enum has Postgres and MySQL only", () => {
  const values = new Set<string>(Object.values(DynamicDBType) as string[]);
  assert.equal(values.size, 2);
  assert.ok(values.has(DynamicDBType.Postgres));
  assert.ok(values.has(DynamicDBType.MySQL));
});
