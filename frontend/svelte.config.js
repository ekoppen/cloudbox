import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Use adapter-node for production Docker deployment
		adapter: adapter({
			// Configure for production deployment
			out: 'build',
			precompress: false,
			envPrefix: '',
			polyfill: false
		})
	}
};

export default config;
