import { cn } from "@/lib/utils";

export function Logo({ className }: { className?: string }) {
	return (
		<svg viewBox="0 0 64 64" aria-hidden="true" className={cn("size-9", className)}>
			<rect width="64" height="64" rx="18" className="fill-primary" />
			<path
				d="M25 11.5v24a2 2 0 0 0 3 1.7l19.2-12a2 2 0 0 0 0-3.4L28 9.8a2 2 0 0 0-3 1.7Z"
				className="fill-primary-foreground"
			/>
			<path
				d="M16 42v4a6 6 0 0 0 6 6h20a6 6 0 0 0 6-6v-4"
				fill="none"
				strokeWidth="5"
				strokeLinecap="round"
				className="stroke-primary-foreground"
			/>
		</svg>
	);
}
