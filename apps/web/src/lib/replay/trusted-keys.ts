const developmentKeys = {
  'development-1': 'O2onvM62pC1io6jQKm8Nc2UyFXcd4kOmOsBIoYtZ2ik=',
} as const;

// Public replay keys are build-time trust anchors. Production images must
// provide both values; ordinary development and test builds retain only the
// fixed development fixture key.
export function buildTrustedReplayKeys(
  keyId: string | undefined,
  publicKey: string | undefined,
  productionBuild: boolean,
): Readonly<Record<string, string>> {
  const hasKeyId = Boolean(keyId);
  const hasPublicKey = Boolean(publicKey);
  if (!hasKeyId && !hasPublicKey && !productionBuild) return developmentKeys;
  if (!hasKeyId || !hasPublicKey) {
    throw new Error('replay key ID and public key must be provided together');
  }
  if (!/^[A-Za-z][A-Za-z0-9_-]{0,63}$/.test(keyId!)) {
    throw new Error('replay key ID must be a safe token');
  }
  let decoded: Uint8Array;
  try {
    decoded = Uint8Array.from(atob(publicKey!), (value) => value.charCodeAt(0));
  } catch {
    throw new Error('replay public key must be valid base64');
  }
  if (decoded.length !== 32 || btoa(String.fromCharCode(...decoded)) !== publicKey) {
    throw new Error('replay public key must be a canonical base64-encoded raw Ed25519 key');
  }
  return Object.freeze({ [keyId!]: publicKey! });
}

export const trustedReplayKeys = buildTrustedReplayKeys(
  import.meta.env.VITE_NINEFOLD_REPLAY_SIGNING_KEY_ID,
  import.meta.env.VITE_NINEFOLD_REPLAY_PUBLIC_KEY,
  import.meta.env.VITE_NINEFOLD_PRODUCTION_BUILD === 'true',
);
