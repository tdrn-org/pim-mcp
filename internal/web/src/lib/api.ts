import type { SessionInfo } from './types';

const BASE = '/api/v1';

export async function getSession(): Promise<SessionInfo> {
	const rsp = await fetch(`${BASE}/session`);
	if (!rsp.ok) throw new Error(`Session fetch failed: ${rsp.status}`);
	return rsp.json();
}

export async function deleteSession(): Promise<void> {
	const rsp = await fetch(`${BASE}/session`, { method: 'DELETE' });
	if (!rsp.ok) throw new Error(`Session delete failed: ${rsp.status}`);
}

export function loginURL(): string {
	return `${BASE}/login`;
}
