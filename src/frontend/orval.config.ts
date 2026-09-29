import { defineConfig } from "orval";

const input = "../backend/gen/docs/swagger/swagger.json";

export default defineConfig({
	vidra: {
		input,
		output: {
			mode: "tags-split",
			target: "src/api/gen",
			schemas: "src/api/gen/model",
			client: "react-query",
			httpClient: "fetch",
			clean: true,
			override: {
				mutator: { path: "src/api/fetcher.ts", name: "apiFetch" },
				fetch: { includeHttpResponseReturnType: false },
				query: { useSuspenseQuery: false, signal: true },
			},
		},
	},
	vidraZod: {
		input,
		output: {
			mode: "single",
			target: "src/api/gen/zod.ts",
			client: "zod",
			fileExtension: ".ts",
		},
	},
});
