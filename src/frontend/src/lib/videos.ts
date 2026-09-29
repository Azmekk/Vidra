import {
	type InfiniteData,
	infiniteQueryOptions,
	keepPreviousData,
	type QueryClient,
} from "@tanstack/react-query";
import type {
	HandlersPaginatedVideoResponse,
	ListVideosParams,
	ServicesProgress,
	ServicesVideoDTO,
	ServicesVideoFileDTO,
} from "@/api/gen/model";
import {
	getGetVideoQueryKey,
	getListProgressQueryKey,
	listVideos,
	useListProgress,
} from "@/api/gen/videos/videos";

export type Video = ServicesVideoDTO;
export type VideoFile = ServicesVideoFileDTO;
export type Progress = ServicesProgress;
type Pages = InfiniteData<HandlersPaginatedVideoResponse, number>;

export const pageSize = 24;
export const activeStatuses = new Set(["queued", "downloading", "encoding"]);

export const libraryQuery = (params: Omit<ListVideosParams, "page" | "limit">) =>
	infiniteQueryOptions({
		queryKey: ["/api/videos", "infinite", params] as const,
		queryFn: ({ pageParam, signal }) =>
			listVideos({ ...params, page: pageParam, limit: pageSize }, { signal }),
		initialPageParam: 1,
		placeholderData: keepPreviousData,
		getNextPageParam: (last) => (last.currentPage < last.totalPages ? last.currentPage + 1 : undefined),
	});

const isLibraryKey = (key: readonly unknown[]) => key[0] === "/api/videos" && key[1] === "infinite";

export function primaryFile(video: Video): VideoFile | undefined {
	return video.files.find((f) => f.id === video.primaryFileId) ?? video.files.at(-1);
}

export function upsertVideo(qc: QueryClient, video: Video) {
	qc.setQueryData(getGetVideoQueryKey(video.id), video);
	let found = false;
	qc.setQueriesData<Pages>({ predicate: (q) => isLibraryKey(q.queryKey) }, (data) => {
		if (!data) return data;
		const pages = data.pages.map((p) => {
			if (!p.videos.some((v) => v.id === video.id)) return p;
			found = true;
			return { ...p, videos: p.videos.map((v) => (v.id === video.id ? video : v)) };
		});
		return { ...data, pages };
	});
	if (!found) qc.invalidateQueries({ predicate: (q) => isLibraryKey(q.queryKey) });
	const done = new Set(video.files.filter((f) => !activeStatuses.has(f.status)).map((f) => f.id));
	qc.setQueryData<Progress[]>(getListProgressQueryKey(), (list) => list?.filter((p) => !done.has(p.fileId)));
}

export function removeVideo(qc: QueryClient, id: string) {
	qc.removeQueries({ queryKey: getGetVideoQueryKey(id), exact: true });
	qc.setQueriesData<Pages>({ predicate: (q) => isLibraryKey(q.queryKey) }, (data) =>
		data
			? {
					...data,
					pages: data.pages.map((p) => ({
						...p,
						totalCount: p.totalCount - 1,
						videos: p.videos.filter((v) => v.id !== id),
					})),
				}
			: data,
	);
}

export function setProgress(qc: QueryClient, progress: Progress) {
	qc.setQueryData<Progress[]>(getListProgressQueryKey(), (list = []) => [
		...list.filter((p) => p.fileId !== progress.fileId),
		progress,
	]);
}

export function useVideoProgress(videoId: string) {
	return useListProgress({
		query: {
			select: (list) => list.filter((p) => p.videoId === videoId),
			staleTime: Number.POSITIVE_INFINITY,
		},
	}).data;
}
