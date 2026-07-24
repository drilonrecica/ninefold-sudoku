import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import openapiTS, { astToString } from 'openapi-typescript';
import { compileFromFile } from 'json-schema-to-typescript';

const root = fileURLToPath(new URL('../', import.meta.url));
const header = '// Code generated from contracts. DO NOT EDIT.\n\n';

async function writeGenerated(relativePath, content) {
  const target = new URL(`../${relativePath}`, import.meta.url);
  await mkdir(new URL('.', target), { recursive: true });
  await writeFile(target, header + content.replace(/^\/\/ Code generated.*\n\n/, ''));
}

const openapiPath = new URL('../contracts/openapi/ninefold.openapi.yaml', import.meta.url);
const openapi = await openapiTS(openapiPath);
await writeGenerated('contracts/generated/typescript/http/index.ts', astToString(openapi));

const schemas = [
  [
    'contracts/websocket/client-message.schema.json',
    'contracts/generated/typescript/realtime/client.ts',
  ],
  [
    'contracts/websocket/server-message.schema.json',
    'contracts/generated/typescript/realtime/server.ts',
  ],
  ['contracts/replay/replay.schema.json', 'contracts/generated/typescript/replay/replay.ts'],
  ['contracts/replay/proof.schema.json', 'contracts/generated/typescript/replay/proof.ts'],
];

for (const [source, target] of schemas) {
  const sourcePath = fileURLToPath(new URL(`../${source}`, import.meta.url));
  const output = await compileFromFile(sourcePath, {
    bannerComment: '',
    additionalProperties: false,
    enableConstEnums: false,
    unreachableDefinitions: true,
  });
  await writeGenerated(target, output);
}

for (const relativePath of [
  'contracts/generated/go/http/types.gen.go',
  'contracts/generated/go/realtime/types.gen.go',
  'contracts/generated/go/replay/types.gen.go',
]) {
  const target = new URL(`../${relativePath}`, import.meta.url);
  const content = await readFile(target, 'utf8');
  await writeGenerated(relativePath, content);
}

if (!root) {
  throw new Error('repository root unavailable');
}
