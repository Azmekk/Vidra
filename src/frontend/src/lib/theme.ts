import { useSyncExternalStore } from "react";

export type Theme = "light" | "dark" | "system";

const key = "vidra-theme";
const media = window.matchMedia("(prefers-color-scheme: dark)");
const listeners = new Set<() => void>();

function read(): Theme {
	const value = localStorage.getItem(key);
	return value === "light" || value === "dark" ? value : "system";
}

function apply() {
	const theme = read();
	const dark = theme === "dark" || (theme === "system" && media.matches);
	document.documentElement.classList.toggle("dark", dark);
	document.documentElement.style.colorScheme = dark ? "dark" : "light";
	for (const l of listeners) l();
}

apply();
media.addEventListener("change", apply);

export function setTheme(theme: Theme) {
	if (theme === read()) return;
	localStorage.setItem(key, theme);
	apply();
}

export function useTheme() {
	const theme = useSyncExternalStore((l) => {
		listeners.add(l);
		return () => listeners.delete(l);
	}, read);
	const resolved = theme === "system" ? (media.matches ? "dark" : "light") : theme;
	return { theme, resolved, setTheme } as const;
}
