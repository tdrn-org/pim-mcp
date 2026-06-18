import type { SessionInfo } from './types';

const BASE = '/api/v1';

export async function getSession(): Promise<SessionInfo> {
	const rsp = await fetch(`${BASE}/session`);
	if (rsp.status === 401) throw new Error('No session (401)');
	if (!rsp.ok) throw new Error(`Session fetch failed: ${rsp.status}`);
	return rsp.json();
}

export async function deleteSession(): Promise<void> {
	const rsp = await fetch(`${BASE}/session`, { method: 'DELETE' });
	if (!rsp.ok) throw new Error(`Session delete failed: ${rsp.status}`);
}

/** Login with API key for session recovery. Submits a form so the browser follows the 302 redirect to /session. */
export function loginWithAPIKey(apiKey: string): void {
	const form = document.createElement('form');
	form.method = 'POST';
	form.action = `${BASE}/login`;
	form.style.display = 'none';

	const input = document.createElement('input');
	input.type = 'hidden';
	input.name = 'api_key';
	input.value = apiKey;
	form.appendChild(input);

	document.body.appendChild(form);
	form.submit();
	document.body.removeChild(form);
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
