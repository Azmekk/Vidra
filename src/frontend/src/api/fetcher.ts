export class ApiError extends Error {
	constructor(
		readonly status: number,
		message: string,
	) {
		super(message);
	}
}

let onUnauthorized: (() => void) | undefined;

export function setUnauthorizedHandler(handler: () => void) {
	onUnauthorized = handler;
}

export async function apiFetch<T>(url: string, init?: RequestInit): Promise<T> {
	const res = await fetch(url, { ...init, credentials: "same-origin" });
	if (!res.ok) {
		const body = await res.json().catch(() => null);
		if (res.status === 401 && !/^\/api\/auth\/(login|setup|status|password)/.test(url)) onUnauthorized?.();
		const message: string = body?.error ?? res.statusText;
		throw new ApiError(res.status, message.charAt(0).toUpperCase() + message.slice(1));
	}
	if (res.status === 204) return undefined as T;
	const type = res.headers.get("content-type") ?? "";
	return (type.includes("application/json") ? res.json() : res.text()) as Promise<T>;
}

export type ErrorType<_> = ApiError;
