import { errorsApi } from '$lib/api-client';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	try {
		const response = await errorsApi.listRecentErrors("", 1, 10);
		return {
			paginatedErrors: response.data
		};
	} catch (e) {
		console.error('Failed to load errors', e);
		return {
			paginatedErrors: {
				errors: [],
				totalCount: 0,
				totalPages: 0,
				currentPage: 1,
				limit: 10
			}
		};
	}
};
