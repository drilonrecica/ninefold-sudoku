import { describe, expect, it } from 'vitest';

import english from './en.json';
import { pseudoLocalize, translate } from './index';

describe('localization', () => {
  it('replaces named placeholders', () => {
    expect(translate('countdown.starts', { count: 3 })).toBe('Match starts in 3');
  });

  it('expands the pseudo locale without changing machine values', () => {
    const result = pseudoLocalize('Join room 7KMP4R');
    expect(result).toContain('7KMP4R');
    expect(result.length).toBeGreaterThan('Join room 7KMP4R'.length * 1.35);
  });

  it('pseudo-localizes every English catalog entry', () => {
    for (const key of Object.keys(english) as (keyof typeof english)[]) {
      const source = translate(key, { count: 3 });
      const result = translate(key, { count: 3 }, 'pseudo');
      expect(result).toMatch(/^［.*］$/);
      expect(result.length).toBeGreaterThan(source.length * 1.3);
    }
  });
});
