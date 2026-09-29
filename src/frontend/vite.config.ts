import { fileURLToPath, URL } from "node:url";
import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import pkg from "./package.json" with { type: "json" };

const backend = process.env.VITE_BACKEND_URL ?? "http://localhost:8080";

export default defineConfig({
	define: { __APP_VERSION__: JSON.stringify(pkg.version) },
	plugins: [react(), babel({ presets: [reactCompilerPreset()] }), tailwindcss()],
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
