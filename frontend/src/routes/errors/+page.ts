import { errorsApi } from '$lib/api-client';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	try {
		const response = await errorsApi.listRecentErrors("", 50); // Fetch last 50
		return {
			errors: response.data
		};
	} catch (e) {
		console.error('Failed to load errors', e);
		return {
			errors: []
		};
	}
};
