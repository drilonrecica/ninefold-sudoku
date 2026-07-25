import { describe, expect, it } from 'vitest';

import { buildTrustedReplayKeys } from './trusted-keys';

const publicKey = 'O2onvM62pC1io6jQKm8Nc2UyFXcd4kOmOsBIoYtZ2ik=';

describe('replay trust build configuration', () => {
  it('uses the development key only outside production images', () => {
    expect(buildTrustedReplayKeys(undefined, undefined, false)).toHaveProperty('development-1');
    expect(() => buildTrustedReplayKeys(undefined, undefined, true)).toThrow(/provided together/);
  });

  it('accepts a valid injected raw Ed25519 public key', () => {
    expect(buildTrustedReplayKeys('replay-1234abcd', publicKey, true)).toEqual({
      'replay-1234abcd': publicKey,
    });
  });

  it('rejects missing, mismatched, and malformed inputs', () => {
    expect(() => buildTrustedReplayKeys('replay-1', undefined, true)).toThrow(/provided together/);
    expect(() => buildTrustedReplayKeys(undefined, publicKey, true)).toThrow(/provided together/);
    expect(() => buildTrustedReplayKeys('../key', publicKey, true)).toThrow(/safe token/);
    expect(() => buildTrustedReplayKeys('replay-1', 'not-base64', true)).toThrow(/base64/);
    expect(() => buildTrustedReplayKeys('replay-1', 'YQ==', true)).toThrow(/Ed25519/);
  });
});
