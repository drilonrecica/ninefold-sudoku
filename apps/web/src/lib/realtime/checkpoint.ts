export interface ConnectionCheckpoint {
  roomCode: string;
  roomVersion?: number;
  matchId?: string;
  matchVersion?: number;
  matchEventNumber?: number;
  pendingRequestIds: string[];
  updatedAt: number;
}

const databaseName = 'ninefold';
const storeName = 'connection_checkpoints';
const databaseVersion = 1;

export function sanitizeCheckpoint(value: ConnectionCheckpoint): ConnectionCheckpoint {
  return {
    roomCode: value.roomCode.trim().toUpperCase().slice(0, 8),
    roomVersion: safeInteger(value.roomVersion),
    matchId: safeUuid(value.matchId),
    matchVersion: safeInteger(value.matchVersion),
    matchEventNumber: safeInteger(value.matchEventNumber),
    pendingRequestIds: [...new Set(value.pendingRequestIds.filter((id) => safeUuid(id)))].slice(
      0,
      32,
    ),
    updatedAt: Number.isSafeInteger(value.updatedAt) ? value.updatedAt : Date.now(),
  };
}

export async function loadCheckpoint(roomCode: string): Promise<ConnectionCheckpoint | null> {
  if (!globalThis.indexedDB) return null;
  const db = await openDatabase();
  return new Promise<ConnectionCheckpoint | null>((resolve, reject) => {
    const request = db.transaction(storeName, 'readonly').objectStore(storeName).get(roomCode);
    request.onsuccess = () =>
      resolve(request.result ? sanitizeCheckpoint(request.result as ConnectionCheckpoint) : null);
    request.onerror = () => reject(request.error);
  }).finally(() => db.close());
}

export async function loadCheckpointByMatchId(
  matchId: string,
): Promise<ConnectionCheckpoint | null> {
  if (!globalThis.indexedDB || !safeUuid(matchId)) return null;
  const db = await openDatabase();
  return new Promise<ConnectionCheckpoint | null>((resolve, reject) => {
    const request = db.transaction(storeName, 'readonly').objectStore(storeName).openCursor();
    request.onsuccess = () => {
      const cursor = request.result;
      if (!cursor) {
        resolve(null);
        return;
      }
      const value = sanitizeCheckpoint(cursor.value as ConnectionCheckpoint);
      if (value.matchId === matchId) resolve(value);
      else cursor.continue();
    };
    request.onerror = () => reject(request.error);
  }).finally(() => db.close());
}

export async function saveCheckpoint(value: ConnectionCheckpoint): Promise<void> {
  if (!globalThis.indexedDB) return;
  const checkpoint = sanitizeCheckpoint(value);
  const db = await openDatabase();
  await new Promise<void>((resolve, reject) => {
    const transaction = db.transaction(storeName, 'readwrite');
    transaction.objectStore(storeName).put(checkpoint, checkpoint.roomCode);
    transaction.oncomplete = () => resolve();
    transaction.onerror = () => reject(transaction.error);
    transaction.onabort = () => reject(transaction.error);
  }).finally(() => db.close());
}

export async function deleteCheckpoint(roomCode: string): Promise<void> {
  if (!globalThis.indexedDB) return;
  const db = await openDatabase();
  await new Promise<void>((resolve, reject) => {
    const transaction = db.transaction(storeName, 'readwrite');
    transaction.objectStore(storeName).delete(roomCode);
    transaction.oncomplete = () => resolve();
    transaction.onerror = () => reject(transaction.error);
  }).finally(() => db.close());
}

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(databaseName, databaseVersion);
    request.onupgradeneeded = () => {
      if (!request.result.objectStoreNames.contains(storeName)) {
        request.result.createObjectStore(storeName);
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

function safeInteger(value: number | undefined): number | undefined {
  return Number.isSafeInteger(value) && (value ?? -1) >= 0 ? value : undefined;
}

function safeUuid(value: string | undefined): string | undefined {
  return value &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(value)
    ? value
    : undefined;
}
