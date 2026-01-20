import { videosApi } from '$lib/api-client';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	try {
		const response = await videosApi.listVideos("", "", 1, 10);
		return {
			paginatedVideos: response.data
		};
	} catch (e) {
		console.error('Failed to load videos', e);
		return {
			paginatedVideos: {
				videos: [],
				totalCount: 0,
				totalPages: 0,
				currentPage: 1,
				limit: 10
			}
		};
	}
};
