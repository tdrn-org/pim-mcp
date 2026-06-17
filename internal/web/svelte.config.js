import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'spa.html',
			precompress: false,
			strict: true
		}),
		prerender: {
			handleHttpError: ({ status, path, referrer }) => {
				// Ignore missing favicon and other non-critical 404s during prerender
				if (status === 404) {
					console.warn(`[prerender] 404 ${path} (linked from ${referrer}) — ignored`);
					return;
				}
				throw new Error(`${status} ${path} (linked from ${referrer})`);
			}
		},
		appDir: '_app'
	}
};

export default config;
