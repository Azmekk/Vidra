import { fileURLToPath, URL } from "node:url";
import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";
import pkg from "./package.json" with { type: "json" };

const backend = process.env.VITE_BACKEND_URL ?? "http://localhost:8080";

export default defineConfig({
	define: { __APP_VERSION__: JSON.stringify(pkg.version) },
	plugins: [
		react(),
		babel({ presets: [reactCompilerPreset()] }),
		tailwindcss(),
		VitePWA({
			strategies: "injectManifest",
			srcDir: "src",
			filename: "sw.ts",
			registerType: "autoUpdate",
			includeAssets: ["favicon.ico", "favicon.svg", "apple-touch-icon-180x180.png"],
			manifest: {
				id: "/",
				name: "Vidra",
				short_name: "Vidra",
				description: "Download, re-encode and save videos to your phone.",
				start_url: "/",
				scope: "/",
				display: "standalone",
				orientation: "portrait",
				background_color: "#020618",
				theme_color: "#020618",
				icons: [
					{ src: "pwa-64x64.png", sizes: "64x64", type: "image/png" },
					{ src: "pwa-192x192.png", sizes: "192x192", type: "image/png" },
					{ src: "pwa-512x512.png", sizes: "512x512", type: "image/png" },
					{ src: "maskable-icon-512x512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
				],
				share_target: {
					action: "/download",
					method: "GET",
					params: { title: "title", text: "text", url: "url" },
				},
				shortcuts: [
					{ name: "Download", url: "/download", icons: [{ src: "pwa-192x192.png", sizes: "192x192" }] },
				],
			},
			injectManifest: {
				globPatterns: ["**/*.{js,css,html,svg,png,ico,woff2}"],
			},
		}),
	],
	resolve: {
		alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
	},
	server: {
		proxy: {
			"/api": { target: backend, ws: true, changeOrigin: false },
			"/swagger": backend,
		},
	},
	build: {
		outDir: "../backend/web/build/app",
		emptyOutDir: true,
	},
});
