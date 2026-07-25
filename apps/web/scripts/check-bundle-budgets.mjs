import { readFile } from 'node:fs/promises';
import { gzipSync } from 'node:zlib';

const outputRoot = new URL('../.svelte-kit/output/client/', import.meta.url);
const manifest = JSON.parse(await readFile(new URL('.vite/manifest.json', outputRoot), 'utf8'));

const entries = {
  start: Object.keys(manifest).find((key) => manifest[key].name === 'entry/start'),
  app: Object.keys(manifest).find((key) => manifest[key].name === 'entry/app'),
  layout: '.svelte-kit/generated/client-optimized/nodes/0.js',
  home: '.svelte-kit/generated/client-optimized/nodes/2.js',
  multiplayer: '.svelte-kit/generated/client-optimized/nodes/8.js',
  solo: '.svelte-kit/generated/client-optimized/nodes/13.js',
};

for (const [name, key] of Object.entries(entries)) {
  if (!key || !manifest[key]) throw new Error(`Missing ${name} entry in the Vite build manifest`);
}

async function compressedJavaScript(entryKeys) {
  const visited = new Set();
  let bytes = 0;

  async function include(key) {
    if (visited.has(key)) return;
    visited.add(key);
    const entry = manifest[key];
    if (!entry) throw new Error(`Missing imported build entry ${key}`);
    if (entry.file.endsWith('.js')) {
      const contents = await readFile(new URL(entry.file, outputRoot));
      bytes += gzipSync(contents).byteLength;
    }
    await Promise.all((entry.imports ?? []).map(include));
  }

  await Promise.all(entryKeys.map(include));
  return bytes;
}

const shared = [entries.start, entries.app, entries.layout];
const results = {
  homepage: await compressedJavaScript([...shared, entries.home]),
  multiplayer: await compressedJavaScript([...shared, entries.multiplayer]),
  solo: await compressedJavaScript([...shared, entries.solo]),
};
const limits = { homepage: 100 * 1024, multiplayer: 200 * 1024, solo: 200 * 1024 };

let failed = false;
for (const [name, bytes] of Object.entries(results)) {
  const kib = (bytes / 1024).toFixed(1);
  const limitKib = limits[name] / 1024;
  console.log(`${name}: ${kib} KiB gzip (limit ${limitKib} KiB)`);
  if (bytes > limits[name]) failed = true;
}

if (failed) process.exitCode = 1;
