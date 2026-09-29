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
		if (res.status === 401 && !url.startsWith("/api/auth/")) onUnauthorized?.();
		throw new ApiError(res.status, body?.error ?? res.statusText);
	}
	if (res.status === 204) return undefined as T;
	const type = res.headers.get("content-type") ?? "";
	return (type.includes("application/json") ? res.json() : res.text()) as Promise<T>;
}
