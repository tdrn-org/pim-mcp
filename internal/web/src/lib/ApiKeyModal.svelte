<script lang="ts">
	interface Props {
		apiKey: string;
		onDismiss: () => void;
	}

	let { apiKey, onDismiss }: Props = $props();

	let copied = $state(false);

	async function copyToClipboard() {
		try {
			await navigator.clipboard.writeText(apiKey);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			// Fallback for older browsers
			const textarea = document.createElement('textarea');
			textarea.value = apiKey;
			textarea.style.position = 'fixed';
			textarea.style.opacity = '0';
			document.body.appendChild(textarea);
			textarea.select();
			document.execCommand('copy');
			document.body.removeChild(textarea);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		}
	}
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" role="dialog" aria-modal="true">
	<div class="mx-4 w-full max-w-lg rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-2xl shadow-brand-500/10">
		<!-- Header -->
		<div class="mb-5 flex items-center gap-3">
			<div class="rounded-full bg-brand-500/10 p-2">
				<svg class="h-6 w-6 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
				</svg>
			</div>
			<div>
				<h3 class="text-lg font-semibold text-slate-100">Your API Key</h3>
				<p class="text-sm text-slate-400">Save this now — it will never be shown again</p>
			</div>
		</div>

		<!-- Key display -->
		<div class="mb-4 rounded-lg border border-slate-700 bg-slate-950 p-4">
			<code class="break-all text-sm font-mono text-brand-300">{apiKey}</code>
		</div>

		<!-- Warning -->
		<div class="mb-5 flex items-start gap-2 rounded-lg bg-amber-500/10 p-3 text-sm text-amber-300">
			<svg class="mt-0.5 h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
			</svg>
			<span>This key is shown <strong>only once</strong>. Copy it now and give it to your agent. If you lose it, you'll need to regenerate a new one.</span>
		</div>

		<!-- Actions -->
		<div class="flex gap-3">
			<button
				onclick={copyToClipboard}
				class="flex items-center gap-2 rounded-lg bg-brand-500 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-brand-600 shadow-lg shadow-brand-500/20"
			>
				{#if copied}
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					Copied!
				{:else}
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
					</svg>
					Copy to Clipboard
				{/if}
			</button>
			<button
				onclick={onDismiss}
				class="rounded-lg bg-slate-800 px-5 py-2.5 text-sm text-slate-300 transition-colors hover:bg-slate-700"
			>
				I've Saved It
			</button>
		</div>
	</div>
</div>
