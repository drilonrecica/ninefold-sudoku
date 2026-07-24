import english from './en.json';

export type MessageKey = keyof typeof english;
export type Locale = 'en' | 'pseudo';

const PLACEHOLDER = /\{([a-zA-Z][a-zA-Z0-9]*)\}/g;

export function pseudoLocalize(message: string): string {
  const expanded = message.replace(/[A-Za-z]/g, (character) => {
    const accents: Record<string, string> = {
      a: 'á',
      e: 'ë',
      i: 'ï',
      o: 'ö',
      u: 'ü',
      A: 'Á',
      E: 'Ë',
      I: 'Ï',
      O: 'Ö',
      U: 'Ü',
    };
    return accents[character] ?? character;
  });
  return `［${expanded} ${'·'.repeat(Math.max(2, Math.ceil(message.length * 0.35)))}］`;
}

export function translate(
  key: MessageKey,
  values: Record<string, string | number> = {},
  locale: Locale = 'en',
): string {
  const source = english[key];
  const rendered = source.replace(PLACEHOLDER, (_, name: string) =>
    String(values[name] ?? `{${name}}`),
  );
  return locale === 'pseudo' ? pseudoLocalize(rendered) : rendered;
}

export const t = translate;
