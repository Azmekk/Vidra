import { useGetEncodingCapabilities } from "@/api/gen/encoding/encoding";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export const originalAudio = "original";

export function AudioFormatPicker({
	value,
	onChange,
	includeOriginal,
}: {
	value: string;
	onChange: (container: string) => void;
	includeOriginal?: boolean;
}) {
	const caps = useGetEncodingCapabilities({ query: { staleTime: Number.POSITIVE_INFINITY } });
	const count = (caps.data?.audioFormats.length ?? 4) + (includeOriginal ? 1 : 0);

	if (!caps.data) {
		return (
			<div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
				{Array.from({ length: count }, (_, i) => (
					// biome-ignore lint/suspicious/noArrayIndexKey: static placeholders
					<Skeleton key={i} className="h-[3.75rem] rounded-2xl" />
				))}
			</div>
		);
	}

	const options = [
		...(includeOriginal
			? [{ container: originalAudio, label: "Original", description: "No conversion" }]
			: []),
		...caps.data.audioFormats,
	];

	return (
		<div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
			{options.map((f) => (
				<button
					key={f.container}
					type="button"
					onClick={() => onChange(f.container)}
					className={cn(
						"rounded-2xl border-2 p-3 text-left transition-colors active:scale-[0.98]",
						value === f.container
							? "border-primary bg-primary/5"
							: "border-transparent bg-muted/60 hover:bg-muted",
					)}
				>
					<span className="block font-bold text-sm">{f.label}</span>
					<span className="block truncate text-muted-foreground text-xs">{f.description}</span>
				</button>
			))}
		</div>
	);
}
