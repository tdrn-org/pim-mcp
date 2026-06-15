<script lang="ts">
	import { getSession, deleteSession, loginURL } from '$lib/api';
	import type { SessionInfo } from '$lib/types';

	let session = $state<SessionInfo | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(true);

	async function loadSession() {
		loading = true;
		error = null;
		try {
			session = await getSession();
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

	function handleLogin() {
		window.location.href = loginURL();
	}

	$effect(() => {
		loadSession();
	});
</script>

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
	{:else if session?.logged_in}
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
		<div class="flex flex-col items-center gap-6 text-center max-w-md">
			<div class="rounded-full bg-slate-800 p-4">
				<svg class="h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
			</div>
			<h2 class="text-2xl font-semibold text-slate-100">PIM MCP Server</h2>
			<p class="text-slate-400">
				Connect to <span class="text-brand-400 font-medium">{session?.provider_name ?? 'your provider'}</span> to enable agent access to your personal information.
			</p>
			<button
				onclick={handleLogin}
				class="mt-2 rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white hover:bg-brand-600 transition-colors shadow-lg shadow-brand-500/20"
			>
				Connect to Provider
			</button>
		</div>
	{/if}
</div>
