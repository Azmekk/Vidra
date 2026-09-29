import { VideosApi, ErrorsApi, YtdlpApi, SystemApi, SettingsApi, Configuration } from '../api';
import { browser } from '$app/environment';

// In the browser, we prefer relative paths to use the Vite proxy.
// On the server, we need the full URL.
const BASE_PATH = browser ? '' : (import.meta.env.VITE_BACKEND_URL || 'http://localhost:8080');

const config = new Configuration({
    basePath: BASE_PATH
});

export const videosApi = new VideosApi(config);
export const errorsApi = new ErrorsApi(config);
export const ytdlpApi = new YtdlpApi(config);
export const systemApi = new SystemApi(config);
export const settingsApi = new SettingsApi(config);
