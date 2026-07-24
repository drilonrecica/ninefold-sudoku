// Public replay keys are build-time trust anchors. Publish a new web build
// containing a rotated key before the server signs with it.
export const trustedReplayKeys: Readonly<Record<string, string>> = {
  'development-1': 'O2onvM62pC1io6jQKm8Nc2UyFXcd4kOmOsBIoYtZ2ik=',
};
