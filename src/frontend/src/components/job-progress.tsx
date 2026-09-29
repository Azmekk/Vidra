import { X } from "lucide-react";
import { useCancelVersion } from "@/api/gen/versions/versions";
import { Button } from "@/components/ui/button";
import { Progress as Bar } from "@/components/ui/progress";
import type { Progress, VideoFile } from "@/lib/videos";

const stageLabels: Record<string, string> = {
	queued: "Queued",
	downloading: "Downloading",
	encoding: "Encoding",
	processing: "Processing",
};

export function JobProgress({ file, progress }: { file: VideoFile; progress?: Progress }) {
	const cancel = useCancelVersion();
	const stage = progress?.stage ?? file.status;
	const percent = Math.max(0, Math.min(100, progress?.percent ?? 0));
	const details = [
		stageLabels[stage] ?? stage,
		file.label,
		progress?.percent ? `${Math.round(percent)}%` : "",
		progress?.speed,
		progress?.eta ? `ETA ${progress.eta}` : "",
	]
		.filter((part) => part && !part.includes("Unknown"))
		.join(" · ");

	return (
		<div className="flex items-center gap-3">
			<div className="min-w-0 flex-1 space-y-2">
				<p className="truncate font-semibold text-muted-foreground text-sm tabular-nums">{details}</p>
				<Bar value={stage === "queued" ? 0 : percent} className="h-2" />
			</div>
			<Button
				variant="ghost"
				size="icon"
				className="size-9 shrink-0 rounded-xl"
				aria-label="Cancel"
				disabled={cancel.isPending}
				onClick={() => cancel.mutate({ id: file.videoId, fileId: file.id })}
			>
				<X />
			</Button>
		</div>
	);
}
