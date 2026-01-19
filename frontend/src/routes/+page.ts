import { videosApi } from '$lib/api-client';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	try {
		const response = await videosApi.listVideos("", "");
		return {
			videos: response.data
		};
	} catch (e) {
		console.error('Failed to load videos', e);
		return {
			videos: []
		};
	}
};
