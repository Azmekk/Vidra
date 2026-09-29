import { cn } from "@/lib/utils";

export function Logo({ className }: { className?: string }) {
	return (
		<svg viewBox="0 0 64 64" aria-hidden="true" className={cn("size-9", className)}>
			<rect width="64" height="64" rx="16" className="fill-primary" />
			<path
				d="M26 15v22l17-11Z"
				strokeWidth="5"
				strokeLinejoin="round"
				className="fill-primary-foreground stroke-primary-foreground"
			/>
			<path d="M22 48.5h20" strokeWidth="5" strokeLinecap="round" className="stroke-primary-foreground" />
		</svg>
	);
}
