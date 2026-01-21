import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');
	const target = env.VITE_BACKEND_URL || 'http://localhost:8080';
	
	return {
		plugins: [tailwindcss(), sveltekit()],
		server: {
			proxy: {
				'/api': {
					target,
					changeOrigin: true,
					ws: true
				},
				'/swagger': {
					target,
					changeOrigin: true
				},
				'/downloads': {
					target,
					changeOrigin: true
				}
			}
		}
	};
});
