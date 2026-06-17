<script lang="ts">
	// No client-side JS needed — csr = false strips this block at build time.
</script>

<div class="flex min-h-screen items-center justify-center bg-slate-950 p-6">
	<div class="flex flex-col items-center gap-8 text-center max-w-lg">
		<!-- Icon -->
		<div class="rounded-full bg-brand-500/10 p-5">
			<svg class="h-14 w-14 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
			</svg>
		</div>

		<!-- Title -->
		<h1 class="text-4xl font-bold text-slate-100 tracking-tight">PIM MCP</h1>
		<p class="text-lg text-slate-400 leading-relaxed">
			Personal Information Management via
			<span class="text-brand-400 font-medium">Model Context Protocol</span>.
			Connect your provider to give AI agents secure access to your emails,
			calendar, tasks, and contacts.
		</p>

		<!-- Human CTA -->
		<a
			href="/session"
			class="rounded-lg bg-brand-500 px-8 py-3 text-base font-medium text-white hover:bg-brand-600 transition-colors shadow-lg shadow-brand-500/20"
		>
			Open Session
		</a>

		<!-- Divider -->
		<div class="flex w-full items-center gap-3">
			<div class="h-px flex-1 bg-slate-800"></div>
			<span class="text-xs font-mono text-slate-600">for agents</span>
			<div class="h-px flex-1 bg-slate-800"></div>
		</div>

		<!-- Agent Instructions -->
		<div class="w-full space-y-3 rounded-lg border border-slate-800 bg-slate-900/60 p-5 text-left">
			<p class="text-sm text-slate-300">
				<span class="text-slate-500">MCP endpoint:</span>
				<code class="ml-1 rounded bg-slate-800 px-2 py-0.5 text-xs text-brand-300 font-mono">POST /mcp</code>
			</p>
			<p class="text-sm text-slate-300">
				<span class="text-slate-500">Authentication:</span>
				<code class="ml-1 rounded bg-slate-800 px-2 py-0.5 text-xs text-brand-300 font-mono">X-API-Key: pim_mcp_...</code>
			</p>
			<p class="text-sm text-slate-300">
				<span class="text-slate-500">API key available at:</span>
				<code class="ml-1 rounded bg-slate-800 px-2 py-0.5 text-xs text-brand-300 font-mono">GET /api/v1/session</code>
			</p>
			<p class="text-sm text-slate-500 leading-relaxed mt-2">
				Use <code class="text-xs text-slate-400">tools/list</code> to discover email, calendar,
				task, and contact tools. StreamableHTTP transport, JSON-RPC 2.0.
			</p>
		</div>

		<!-- Footer -->
		<p class="text-xs text-slate-600">
			<a href="https://github.com/tdrn-org/pim-mcp" class="hover:text-slate-400 transition-colors">tdrn-org/pim-mcp</a>
			<span class="mx-2">·</span>
			Apache 2.0
		</p>
	</div>
</div>
