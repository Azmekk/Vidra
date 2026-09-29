const urlInText = /\bhttps?:\/\/[^\s<>"'`]+/i;
const bareUrlInText = /\b(?:www\.)?[a-z0-9-]+(?:\.[a-z0-9-]+)*\.[a-z]{2,}\/[^\s<>"'`]*/i;
const trailing = /[.,;:!?…»”’]+$/;

function trimTrailing(url: string) {
	let s = url.replace(trailing, "");
	for (const [open, close] of [
		["(", ")"],
		["[", "]"],
		["{", "}"],
	] as const) {
		while (s.endsWith(close) && s.split(open).length < s.split(close).length) s = s.slice(0, -1);
	}
	return s.replace(trailing, "");
}

export function extractUrl(text: string) {
	const trimmed = text.trim();
	const full = trimmed.match(urlInText)?.[0];
	if (full) return trimTrailing(full);
	const bare = trimmed.match(bareUrlInText)?.[0];
	return bare ? `https://${trimTrailing(bare)}` : trimmed;
}
