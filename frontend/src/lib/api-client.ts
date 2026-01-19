import { VideosApi, ErrorsApi, YtdlpApi, Configuration } from '../api';
import { browser } from '$app/environment';

// For SSR, we need the full URL. For client, we use the proxy (relative path).
let BASE_PATH = '';

if (!browser) {
    // Server-side
    if (typeof process !== 'undefined' && process.env && process.env.VITE_BACKEND_URL) {
        BASE_PATH = process.env.VITE_BACKEND_URL;
    } else {
        BASE_PATH = 'http://localhost:8080';
    }
}

const config = new Configuration({
    basePath: BASE_PATH
});

export const videosApi = new VideosApi(config);
export const errorsApi = new ErrorsApi(config);
export const ytdlpApi = new YtdlpApi(config);
