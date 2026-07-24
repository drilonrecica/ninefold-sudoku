import { describe, expect, it } from 'vitest';

import { defaultGamePreferences, loadGamePreferences } from './preferences';

describe('game preferences', () => {
  it('keeps sound and haptics off by default', () => {
    const storage = { getItem: () => null };
    expect(loadGamePreferences(storage)).toEqual(defaultGamePreferences);
  });

  it('restores explicit input mode without enabling effects', () => {
    const storage = {
      getItem: (key: string) => (key === 'ninefold.inputMode' ? 'number-first' : null),
    };
    expect(loadGamePreferences(storage)).toEqual({
      inputMode: 'number-first',
      soundEnabled: false,
      hapticsEnabled: false,
    });
  });
});
