import { env } from '$env/dynamic/private';

export function apiBaseUrl(): string {
  return (env.NINEFOLD_API_BASE_URL ?? 'http://127.0.0.1:8080').replace(/\/$/, '');
}
