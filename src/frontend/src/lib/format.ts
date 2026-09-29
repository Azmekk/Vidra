import type { VideoFile } from "@/lib/videos";

export function formatBytes(bytes?: number) {
	if (!bytes) return "";
	const units = ["B", "KB", "MB", "GB", "TB"];
	const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
	return `${(bytes / 1024 ** i).toFixed(i < 2 ? 0 : 1)} ${units[i]}`;
}

export function formatDuration(seconds?: number) {
	if (!seconds) return "";
	const s = Math.round(seconds);
	const h = Math.floor(s / 3600);
	const m = Math.floor((s % 3600) / 60);
	const pad = (n: number) => String(n).padStart(2, "0");
	return h ? `${h}:${pad(m)}:${pad(s % 60)}` : `${m}:${pad(s % 60)}`;
}

const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
const steps: [Intl.RelativeTimeFormatUnit, number][] = [
	["year", 31536000],
	["month", 2592000],
	["week", 604800],
	["day", 86400],
	["hour", 3600],
	["minute", 60],
];

export function formatRelative(iso: string) {
	const diff = (new Date(iso).getTime() - Date.now()) / 1000;
	for (const [unit, secs] of steps) {
		if (Math.abs(diff) >= secs) return rtf.format(Math.round(diff / secs), unit);
	}
	return "just now";
}

const codecNames: Record<string, string> = {
	h264: "H.264",
	hevc: "HEVC",
	av1: "AV1",
	vp9: "VP9",
	vp8: "VP8",
	aac: "AAC",
	opus: "Opus",
	mp3: "MP3",
	flac: "FLAC",
};

export const codecName = (c?: string) => (c ? (codecNames[c] ?? c.toUpperCase()) : "");

export function fileSummary(f: VideoFile) {
	return [
		codecName(f.videoCodec),
		f.height ? `${f.height}p` : "",
		f.fps ? `${Math.round(f.fps)}fps` : "",
		f.container?.toUpperCase(),
		formatBytes(f.fileSize),
	]
		.filter(Boolean)
		.join(" · ");
}

export const randomNamePattern = /^[a-z]+-[a-z]+-\d{4}$/;
