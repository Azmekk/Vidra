import { defineConfig } from "@vite-pwa/assets-generator/config";

const tile = { background: "#0f172b", fit: "contain" } as const;

export default defineConfig({
	headLinkOptions: { preset: "2023" },
	preset: {
		transparent: { sizes: [64, 192, 512], favicons: [[48, "favicon.ico"]] },
		maskable: { sizes: [512], padding: 0.1, resizeOptions: tile },
		apple: { sizes: [180], padding: 0, resizeOptions: tile },
	},
	images: ["public/favicon.svg"],
});
