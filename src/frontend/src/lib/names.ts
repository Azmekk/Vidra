const forbidden = new Set(["<", ">", ":", '"', "/", "\\", "|", "?", "*"]);
const reserved = /^(con|prn|aux|nul|com[1-9]|lpt[1-9])(\..*)?$/i;

const isForbidden = (c: string) => {
	const code = c.charCodeAt(0);
	return forbidden.has(c) || code < 32 || code === 127;
};

/** Mirrors the server's check, since a video's name is also the name of its files. */
export function nameError(name: string) {
	const n = name.trim();
	if (!n) return "Name is required";
	if ([...n].length > 255) return "Name must be at most 255 characters";
	if ([...n].some(isForbidden)) return "Name can't contain < > : \" / \\ | ? *";
	if (/^\.+$/.test(n) || reserved.test(n)) return "Name is reserved by the file system";
}

/** Turns a source title into a valid name, e.g. "Show: Part 1" into "Show - Part 1". */
export function cleanName(title: string) {
	const name = [...title.replace(/\s*[:|]\s*/g, " - ").replace(/[/\\]/g, "-")]
		.filter((c) => !isForbidden(c))
		.join("")
		.replace(/\s+/g, " ")
		.trim();
	return [...name].slice(0, 255).join("");
}
