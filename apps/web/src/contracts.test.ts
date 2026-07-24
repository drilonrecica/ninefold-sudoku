import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import Ajv2020 from 'ajv/dist/2020.js';
import { parse } from 'yaml';
import { describe, expect, it } from 'vitest';
import type { components } from '../../../contracts/generated/typescript/http/index';
import type { ClientMessage } from '../../../contracts/generated/typescript/realtime/client';
import type { ReplayDocument } from '../../../contracts/generated/typescript/replay/replay';

const repositoryRoot = resolve(process.cwd(), '../..');

async function json(relativePath: string): Promise<Record<string, unknown>> {
  return JSON.parse(await readFile(resolve(repositoryRoot, relativePath), 'utf8')) as Record<
    string,
    unknown
  >;
}

describe('generated contracts and shared fixtures', () => {
  it('decodes HTTP fixtures without losing Unicode or safe integers', async () => {
    const success = (await json(
      'contracts/fixtures/http-success.json',
    )) as components['schemas']['SuccessEnvelope'];
    expect(success.version).toBe(Number.MAX_SAFE_INTEGER);
    expect(success.data.displayName).toBe('Éva 🧩');

    for (const name of [
      'error-room.json',
      'error-lifecycle.json',
      'error-gameplay.json',
      'error-concurrency.json',
      'error-replay.json',
    ]) {
      const envelope = (await json(
        `contracts/fixtures/${name}`,
      )) as components['schemas']['ErrorEnvelope'];
      expect(envelope.error.code).toMatch(/^[A-Z][A-Z0-9_]+$/);
    }
  });

  it('validates OpenAPI envelope constraints', async () => {
    const source = parse(
      await readFile(resolve(repositoryRoot, 'contracts/openapi/ninefold.openapi.yaml'), 'utf8'),
    ) as { components: { schemas: Record<string, unknown> } };
    const ajv = new Ajv2020({ allErrors: true, strict: false, formats: { int64: true } });
    const validateSuccess = ajv.compile({
      $ref: '#/components/schemas/SuccessEnvelope',
      components: source.components,
    });
    const validateError = ajv.compile({
      $ref: '#/components/schemas/ErrorEnvelope',
      components: source.components,
    });
    const success = await json('contracts/fixtures/http-success.json');
    const error = await json('contracts/fixtures/error-concurrency.json');
    expect(validateSuccess(success)).toBe(true);
    expect(validateError(error)).toBe(true);

    const missingId = structuredClone(success);
    delete missingId.requestId;
    expect(validateSuccess(missingId)).toBe(false);

    expect(validateSuccess({ ...success, version: Number.MAX_SAFE_INTEGER + 1 })).toBe(false);
    expect(validateSuccess({ ...success, unexpected: true })).toBe(false);

    const credentialDetail = structuredClone(error) as {
      error: { details: Record<string, unknown> };
    };
    credentialDetail.error.details.sessionToken = 'not-allowed';
    expect(validateError(credentialDetail)).toBe(false);
  });

  it('rejects malformed realtime values and unsupported schema versions', async () => {
    const schema = await json('contracts/websocket/client-message.schema.json');
    const validate = new Ajv2020({ allErrors: true }).compile(schema);
    const fixture = (await json(
      'contracts/fixtures/realtime-client.json',
    )) as unknown as ClientMessage;
    expect(validate(fixture)).toBe(true);

    const malformed = structuredClone(fixture);
    malformed.requestId = 'not-a-uuid';
    expect(validate(malformed)).toBe(false);
    expect(validate({ ...fixture, clientSequence: Number.MAX_SAFE_INTEGER + 1 })).toBe(false);

    const unsupported = await json('contracts/fixtures/realtime-unsupported-schema.json');
    expect(validate(unsupported)).toBe(false);
  });

  it('validates replay fixtures', async () => {
    const schema = await json('contracts/replay/replay.schema.json');
    const validate = new Ajv2020({ allErrors: true }).compile(schema);
    const fixture = (await json('contracts/fixtures/replay.json')) as unknown as ReplayDocument;
    expect(validate(fixture)).toBe(true);
    expect(validate({ ...fixture, replayId: '0190a7d3-8d9a-4f31-8d2a-1242f6f0d10b' })).toBe(false);
  });
});
