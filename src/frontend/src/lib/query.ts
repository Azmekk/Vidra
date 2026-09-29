import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
import { QueryClient } from "@tanstack/react-query";
import { del, get, set } from "idb-keyval";
import { toast } from "sonner";
import { ApiError } from "@/api/fetcher";

export const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 30_000,
			gcTime: 24 * 60 * 60 * 1000,
			retry: (count, error) => !(error instanceof ApiError && error.status < 500) && count < 2,
		},
		mutations: {
			onError: (error) => {
				if (!(error instanceof ApiError && error.status === 401)) toast.error(error.message);
			},
		},
	},
});

const storageKey = "vidra-query-cache";

export const persister = createAsyncStoragePersister({
	key: storageKey,
	storage: { getItem: get, setItem: set, removeItem: del },
	serialize: (data) => data as never,
	deserialize: (data) => data as never,
});

export async function clearCache() {
	queryClient.clear();
	await del(storageKey);
}
