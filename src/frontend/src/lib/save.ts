import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import type { VideoFile } from "@/lib/videos";

const mimeTypes: Record<string, string> = {
	mp4: "video/mp4",
	m4v: "video/mp4",
	mov: "video/quicktime",
	webm: "video/webm",
	mkv: "video/x-matroska",
};

const isIOS =
	/iPhone|iPad|iPod/.test(navigator.userAgent) ||
	(navigator.userAgent.includes("Macintosh") && navigator.maxTouchPoints > 1);

export const canSaveToPhotos = (() => {
	try {
		return (
			isIOS &&
			window.isSecureContext &&
			typeof navigator.canShare === "function" &&
			navigator.canShare({ files: [new File([], "v.mp4", { type: "video/mp4" })] })
		);
	} catch {
		return false;
	}
})();

export function downloadUrl(file: VideoFile) {
	return `${file.url}?download=1`;
}

export function fileName(name: string, file: VideoFile) {
	const safe = name.replace(/[\\/:*?"<>|]+/g, " ").trim() || "video";
	return `${safe}.${file.container ?? "mp4"}`;
}

export type SaveState =
	| { kind: "idle" }
	| { kind: "fetching"; progress: number }
	| { kind: "ready"; file: File };

export function useSaveToPhotos(name: string, videoFile?: VideoFile) {
	const [state, setState] = useState<SaveState>({ kind: "idle" });
	const abort = useRef<AbortController | null>(null);

	useEffect(() => () => abort.current?.abort(), []);

	useEffect(() => {
		if (state.kind !== "ready") return;
		const t = setTimeout(() => setState({ kind: "idle" }), 120_000);
		return () => clearTimeout(t);
	}, [state]);

	async function prepare() {
		if (!videoFile?.url) return;
		abort.current = new AbortController();
		setState({ kind: "fetching", progress: 0 });
		try {
			const res = await fetch(videoFile.url, { signal: abort.current.signal });
			if (!res.ok || !res.body) throw new Error(`Download failed (${res.status})`);
			const total = Number(res.headers.get("content-length")) || videoFile.fileSize || 0;
			const reader = res.body.getReader();
			const chunks: Uint8Array<ArrayBuffer>[] = [];
			let received = 0;
			for (;;) {
				const { done, value } = await reader.read();
				if (done) break;
				chunks.push(value);
				received += value.length;
				if (total) setState({ kind: "fetching", progress: received / total });
			}
			const type = mimeTypes[videoFile.container ?? ""] ?? "video/mp4";
			setState({ kind: "ready", file: new File(chunks, fileName(name, videoFile), { type }) });
		} catch (e) {
			if ((e as Error).name !== "AbortError") toast.error((e as Error).message);
			setState({ kind: "idle" });
		}
	}

	async function share() {
		if (state.kind !== "ready") return;
		try {
			await navigator.share({ files: [state.file] });
			setState({ kind: "idle" });
		} catch (e) {
			if ((e as Error).name !== "AbortError") toast.error((e as Error).message);
		}
	}

	function cancel() {
		abort.current?.abort();
		setState({ kind: "idle" });
	}

	return { state, prepare, share, cancel };
}
