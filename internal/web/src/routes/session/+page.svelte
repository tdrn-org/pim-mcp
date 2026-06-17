<script lang="ts">
	import { getSession, deleteSession, loginWithAPIKey, loginOAuth2 } from '$lib/api';
	import type { SessionInfo } from '$lib/types';
	import ApiKeyModal from '$lib/ApiKeyModal.svelte';

	let session = $state<SessionInfo | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(true);
	let showApiKeyModal = $state(false);
	let apiKeyInput = $state('');
	let loginError = $state<string | null>(null);
	let loginLoading = $state(false);

	async function loadSession() {
		loading = true;
		error = null;
		try {
			session = await getSession();
			// Show API key modal if key is present (first-time session)
			if (session.api_key) {
				showApiKeyModal = true;
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}

	async function handleLogout() {
		try {
			await deleteSession();
			await loadSession();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Logout failed';
		}
	}

	function handleOAuth2Login() {
		loginOAuth2();
	}

	async function handleAPIKeyLogin() {
		if (!apiKeyInput.trim()) return;
		loginLoading = true;
		loginError = null;
		try {
			session = await loginWithAPIKey(apiKeyInput.trim());
			apiKeyInput = '';
		} catch (e) {
			loginError = e instanceof Error ? e.message : 'Login failed';
		} finally {
			loginLoading = false;
		}
	}

	function handleApiKeyDismiss() {
		showApiKeyModal = false;
		// Reload session to get updated state (api_key will be empty now)
		loadSession();
	}

	function formatExpiry(iso: string): string {
		if (!iso || iso === '0001-01-01T00:00:00Z') return 'N/A';
		const d = new Date(iso);
		return d.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	$effect(() => {
		loadSession();
	});
</script>

{#if showApiKeyModal && session?.api_key}
	<ApiKeyModal apiKey={session.api_key} onDismiss={handleApiKeyDismiss} />
{/if}

<div class="flex min-h-screen items-center justify-center p-6">
	{#if loading}
		<div class="flex flex-col items-center gap-4">
			<div class="h-10 w-10 animate-spin rounded-full border-4 border-slate-600 border-t-brand-400"></div>
			<p class="text-slate-400 text-sm">Connecting to PIM server...</p>
		</div>
	{:else if error}
		<div class="flex flex-col items-center gap-4 text-center max-w-md">
			<div class="rounded-full bg-red-500/10 p-4">
				<svg class="h-10 w-10 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
				</svg>
			</div>
			<h2 class="text-xl font-semibold text-slate-200">Connection Error</h2>
			<p class="text-slate-400 text-sm">{error}</p>
			<button
				onclick={() => loadSession()}
				class="mt-2 rounded-lg bg-slate-800 px-4 py-2 text-sm text-slate-300 hover:bg-slate-700 transition-colors"
			>
				Retry
			</button>
		</div>
	{:else if session?.credentials?.valid}
		<div class="flex flex-col items-center gap-6 text-center max-w-md">
			<div class="rounded-full bg-emerald-500/10 p-4">
				<svg class="h-12 w-12 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<h2 class="text-2xl font-semibold text-slate-100">Connected</h2>
			<p class="text-slate-400">
				Authenticated with <span class="text-brand-400 font-medium">{session.provider_name}</span>
			</p>

			<!-- Credential info -->
			<div class="w-full rounded-lg border border-slate-700 bg-slate-900/50 p-4">
				<div class="flex items-center justify-between text-sm">
					<span class="text-slate-400">Credentials</span>
					<span class="rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-medium text-emerald-400">Valid</span>
				</div>
				{#if session.credentials.expiry && session.credentials.expiry !== '0001-01-01T00:00:00Z'}
					<div class="mt-2 flex items-center justify-between text-sm">
						<span class="text-slate-500">Expires</span>
						<span class="text-slate-300">{formatExpiry(session.credentials.expiry)}</span>
					</div>
				{/if}
			</div>

			<div class="flex gap-3 mt-2">
				<button
					onclick={handleLogout}
					class="rounded-lg bg-slate-800 px-5 py-2.5 text-sm text-slate-300 hover:bg-slate-700 transition-colors"
				>
					Disconnect
				</button>
			</div>
		</div>
	{:else}
		<div class="flex flex-col items-center gap-6 text-center max-w-md w-full">
			<div class="rounded-full bg-slate-800 p-4">
				<svg class="h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
			</div>
			<h2 class="text-2xl font-semibold text-slate-100">PIM MCP Server</h2>
			<p class="text-slate-400">
				Connect to <span class="text-brand-400 font-medium">{session?.provider_name ?? 'your provider'}</span> to enable agent access to your personal information.
			</p>

			<!-- OAuth2 Login -->
			<button
				onclick={handleOAuth2Login}
				class="w-full rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white hover:bg-brand-600 transition-colors shadow-lg shadow-brand-500/20"
			>
				Connect to Provider
			</button>

			<!-- Divider -->
			<div class="flex w-full items-center gap-3">
				<div class="h-px flex-1 bg-slate-700"></div>
				<span class="text-xs text-slate-500">or use API key</span>
				<div class="h-px flex-1 bg-slate-700"></div>
			</div>

			<!-- API Key Login -->
			<div class="w-full space-y-3">
				<div class="flex gap-2">
					<input
						type="text"
						bind:value={apiKeyInput}
						placeholder="pim_mcp_..."
						class="flex-1 rounded-lg border border-slate-700 bg-slate-900 px-4 py-2.5 text-sm text-slate-200 placeholder-slate-500 outline-none transition-colors focus:border-brand-500"
						onkeydown={(e) => { if (e.key === 'Enter') handleAPIKeyLogin(); }}
					/>
					<button
						onclick={handleAPIKeyLogin}
						disabled={loginLoading || !apiKeyInput.trim()}
						class="rounded-lg bg-slate-700 px-4 py-2.5 text-sm text-slate-200 transition-colors hover:bg-slate-600 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{#if loginLoading}
							<div class="h-4 w-4 animate-spin rounded-full border-2 border-slate-400 border-t-white"></div>
						{:else}
							Login
						{/if}
					</button>
				</div>
				{#if loginError}
					<p class="text-sm text-red-400">{loginError}</p>
				{/if}
			</div>
		</div>
	{/if}
</div>
