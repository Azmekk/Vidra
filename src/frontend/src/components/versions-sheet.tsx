import { useQueryClient } from "@tanstack/react-query";
import { Apple, Download, EllipsisVertical, Play, Plus, Star, Trash2, Wand2 } from "lucide-react";
import { useState } from "react";
import { useDeleteVersion, useUpdateVersion } from "@/api/gen/versions/versions";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { JobProgress } from "@/components/job-progress";
import { Sheet } from "@/components/sheet";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { fileSummary, formatRelative } from "@/lib/format";
import { downloadUrl } from "@/lib/save";
import { cn } from "@/lib/utils";
import {
	activeStatuses,
	type Progress,
	upsertVideo,
	useVideoProgress,
	type Video,
	type VideoFile,
} from "@/lib/videos";

export function VersionsSheet({
	video,
	open,
	onOpenChange,
	onReencode,
}: {
	video: Video;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onReencode: (sourceFileId?: string) => void;
}) {
	const progress = useVideoProgress(video.id);
	const files = [...video.files].reverse();

	return (
		<Sheet
			open={open}
			onOpenChange={onOpenChange}
			title="Versions"
			description={video.name}
			footer={
				<Button
					size="lg"
					variant="secondary"
					className="h-12 w-full rounded-2xl font-bold text-base"
					onClick={() => onReencode()}
				>
					<Plus /> New version
				</Button>
			}
		>
			<ul className="space-y-3">
				{files.map((f) => (
					<VersionRow
						key={f.id}
						video={video}
						file={f}
						progress={progress?.find((p) => p.fileId === f.id)}
						onReencode={() => onReencode(f.id)}
					/>
				))}
			</ul>
		</Sheet>
	);
}

function VersionRow({
	video,
	file,
	progress,
	onReencode,
}: {
	video: Video;
	file: VideoFile;
	progress?: Progress;
	onReencode: () => void;
}) {
	const qc = useQueryClient();
	const [confirm, setConfirm] = useState(false);
	const primary = file.id === video.primaryFileId;
	const active = activeStatuses.has(file.status);
	const completed = file.status === "completed";
	const onVideo = { mutation: { onSuccess: (v: Video) => upsertVideo(qc, v) } };
	const update = useUpdateVersion(onVideo);
	const remove = useDeleteVersion(onVideo);

	return (
		<li
			className={cn("rounded-3xl border p-4 transition-colors", primary && "border-primary/40 bg-primary/5")}
		>
			<div className="flex items-start gap-3">
				<div className="min-w-0 flex-1">
					<div className="flex flex-wrap items-center gap-2">
						<span className="font-bold">{file.label}</span>
						{primary && (
							<Badge className="rounded-full">
								<Star className="fill-current" /> Default
							</Badge>
						)}
						{file.iosCompatible && (
							<Badge variant="secondary" className="rounded-full">
								<Apple /> iPhone
							</Badge>
						)}
						{(file.status === "error" || file.status === "canceled") && (
							<Badge variant="destructive" className="rounded-full capitalize">
								{file.status}
							</Badge>
						)}
					</div>
					<p className="mt-1 text-muted-foreground text-sm">
						{[fileSummary(file), formatRelative(file.createdAt)].filter(Boolean).join(" · ")}
					</p>
				</div>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button
							variant="ghost"
							size="icon"
							className="size-9 shrink-0 rounded-xl"
							aria-label="Version actions"
						>
							<EllipsisVertical />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end" className="w-52 rounded-2xl p-1.5">
						{completed && (
							<>
								<DropdownMenuItem asChild>
									<a href={file.url} target="_blank" rel="noreferrer">
										<Play /> Play
									</a>
								</DropdownMenuItem>
								<DropdownMenuItem asChild>
									<a href={downloadUrl(file)} download>
										<Download /> Download
									</a>
								</DropdownMenuItem>
								{!primary && (
									<DropdownMenuItem
										onSelect={() => update.mutate({ id: video.id, fileId: file.id, data: { primary: true } })}
									>
										<Star /> Set as default
									</DropdownMenuItem>
								)}
								<DropdownMenuItem onSelect={onReencode}>
									<Wand2 /> Re-encode from this
								</DropdownMenuItem>
								<DropdownMenuSeparator />
							</>
						)}
						<DropdownMenuItem
							variant="destructive"
							disabled={video.files.length < 2}
							onSelect={() => setConfirm(true)}
						>
							<Trash2 /> Delete version
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</div>
			{active && (
				<div className="mt-3">
					<JobProgress file={file} progress={progress} />
				</div>
			)}
			<ConfirmDialog
				open={confirm}
				onOpenChange={setConfirm}
				title="Delete this version?"
				description={`"${file.label}" will be removed from disk. Other versions are kept.`}
				action="Delete"
				onConfirm={() => remove.mutate({ id: video.id, fileId: file.id })}
			/>
		</li>
	);
}
