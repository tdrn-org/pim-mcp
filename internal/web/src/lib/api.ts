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

/** Login with API key for session recovery. Returns SessionInfo on success. */
export async function loginWithAPIKey(apiKey: string): Promise<SessionInfo> {
	const rsp = await fetch(`${BASE}/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ api_key: apiKey })
	});
	if (rsp.status === 401) throw new Error('Invalid API key');
	if (!rsp.ok) throw new Error(`Login failed: ${rsp.status}`);
	return rsp.json();
}

/** Initiate OAuth2 provider login. Submits a form so the browser follows the 302 redirect. */
export function loginOAuth2(): void {
	const form = document.createElement('form');
	form.method = 'POST';
	form.action = `${BASE}/login`;
	form.style.display = 'none';
	document.body.appendChild(form);
	form.submit();
	document.body.removeChild(form);
}
