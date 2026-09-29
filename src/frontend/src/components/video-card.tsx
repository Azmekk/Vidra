import { useQueryClient } from "@tanstack/react-query";
import {
	Apple,
	Check,
	Copy,
	Download,
	Ellipsis,
	ExternalLink,
	Film,
	Layers,
	Link2,
	Loader2,
	Pencil,
	Play,
	Share,
	Trash2,
	Type,
	Wand2,
	X,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { useCreateVersion } from "@/api/gen/versions/versions";
import { useDeleteVideo, useUpdateVideo } from "@/api/gen/videos/videos";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { EncodeSheet } from "@/components/encode-sheet";
import { JobProgress } from "@/components/job-progress";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { VersionsSheet } from "@/components/versions-sheet";
import { formatDuration, randomNamePattern } from "@/lib/format";
import { canShareFiles, downloadUrl, isIOS, useSaveToPhotos } from "@/lib/save";
import { cn } from "@/lib/utils";
import {
	activeStatuses,
	primaryFile,
	removeVideo,
	upsertVideo,
	useVideoProgress,
	type Video,
	type VideoFile,
} from "@/lib/videos";

const statusStyles: Record<string, string> = {
	queued: "bg-muted/90 text-foreground",
	downloading: "bg-blue-500/90 text-white",
	encoding: "bg-violet-500/90 text-white",
	error: "bg-destructive/90 text-white",
	canceled: "bg-muted/90 text-foreground",
};

export function VideoCard({ video }: { video: Video }) {
	const [playing, setPlaying] = useState(false);
	const [versionsOpen, setVersionsOpen] = useState(false);
	const [encode, setEncode] = useState<{ open: boolean; source?: string; key: number }>({
		open: false,
		key: 0,
	});
	const [confirmDelete, setConfirmDelete] = useState(false);
	const [renaming, setRenaming] = useState(false);
	const progress = useVideoProgress(video.id);
	const qc = useQueryClient();

	const primary = primaryFile(video);
	const ready = primary?.status === "completed";
	const active = video.files.filter((f) => activeStatuses.has(f.status));
	const completedCount = video.files.filter((f) => f.status === "completed").length;

	const remove = useDeleteVideo({
		mutation: {
			onMutate: () => removeVideo(qc, video.id),
			onSuccess: () => toast.success("Video deleted"),
			onError: () => qc.invalidateQueries({ queryKey: ["/api/videos"] }),
		},
	});

	const openEncode = (source?: string) => {
		setVersionsOpen(false);
		setEncode((e) => ({ open: true, source, key: e.key + 1 }));
	};

	const copy = async (text: string, what: string) => {
		await navigator.clipboard.writeText(text);
		toast.success(`${what} copied`);
	};

	return (
		<article className="group overflow-hidden rounded-[2.5rem] border bg-card transition-shadow hover:shadow-2xl hover:shadow-primary/5">
			<div className="relative aspect-video w-full overflow-hidden bg-muted">
				{playing && primary?.url ? (
					// biome-ignore lint/a11y/useMediaCaption: user videos have no captions
					<video
						src={primary.url}
						controls
						autoPlay
						playsInline
						className="size-full bg-black object-contain"
					/>
				) : (
					<button
						type="button"
						className="relative block size-full disabled:cursor-default"
						onClick={() => setPlaying(true)}
						disabled={!ready}
						aria-label="Play"
					>
						<Thumbnail video={video} />
						<div className="absolute top-4 right-4 left-4 flex items-start justify-between gap-2">
							<span />
							{video.status !== "completed" && (
								<Badge
									className={cn(
										"rounded-full border-none px-3 py-1 font-bold uppercase backdrop-blur-md",
										statusStyles[video.status],
									)}
								>
									{video.status}
								</Badge>
							)}
						</div>
						{video.duration ? (
							<span className="absolute right-4 bottom-4 rounded-lg bg-black/70 px-2 py-0.5 font-bold text-white text-xs tabular-nums">
								{formatDuration(video.duration)}
							</span>
						) : null}
						{ready && (
							<span className="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover:bg-black/30 [@media(hover:none)]:bg-black/20">
								<span className="rounded-full bg-white p-4 text-black opacity-0 shadow-2xl transition-all group-hover:opacity-100 [@media(hover:none)]:opacity-100">
									<Play className="size-7 fill-current" />
								</span>
							</span>
						)}
					</button>
				)}
				{completedCount > 1 && !playing && (
					<button
						type="button"
						onClick={() => setVersionsOpen(true)}
						className="absolute top-4 left-4 inline-flex items-center gap-1.5 rounded-full bg-black/60 px-3 py-1 font-bold text-white text-xs backdrop-blur-md transition-colors hover:bg-black/75"
					>
						<Layers className="size-3.5" /> {completedCount} versions
					</button>
				)}
			</div>

			<div className="space-y-5 p-6">
				<div className="min-w-0 space-y-1.5">
					{renaming ? (
						<RenameField video={video} onDone={() => setRenaming(false)} />
					) : (
						<div className="flex items-start gap-2">
							<h3
								className="line-clamp-2 min-w-0 flex-1 font-bold text-xl leading-tight tracking-tight"
								title={video.name}
							>
								{video.name}
							</h3>
							<Button
								variant="ghost"
								size="icon"
								className="-mt-1 size-8 shrink-0 rounded-lg text-muted-foreground"
								aria-label="Rename"
								onClick={() => setRenaming(true)}
							>
								<Pencil />
							</Button>
						</div>
					)}
					<div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-muted-foreground text-xs">
						{video.uploader && <span className="font-medium">{video.uploader}</span>}
						<button
							type="button"
							onClick={() => copy(video.id, "ID")}
							className="max-w-40 truncate font-mono transition-colors hover:text-foreground"
						>
							{video.id}
						</button>
					</div>
					{!renaming &&
						video.sourceTitle &&
						video.sourceTitle !== video.name &&
						randomNamePattern.test(video.name) && <UseSourceTitle video={video} />}
				</div>

				{active.length > 0 && (
					<div className="space-y-3">
						{active.map((f) => (
							<JobProgress key={f.id} file={f} progress={progress?.find((p) => p.fileId === f.id)} />
						))}
					</div>
				)}

				<div className="flex gap-2">
					<PrimaryAction video={video} file={ready ? primary : undefined} status={video.status} />
					<Button
						variant="secondary"
						size="icon"
						className="size-12 shrink-0 rounded-2xl"
						aria-label="Re-encode"
						disabled={completedCount === 0}
						onClick={() => openEncode()}
					>
						<Wand2 className="size-5" />
					</Button>
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button
								variant="secondary"
								size="icon"
								className="size-12 shrink-0 rounded-2xl"
								aria-label="More"
							>
								<Ellipsis className="size-5" />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end" className="w-56 rounded-2xl p-1.5">
							<DropdownMenuItem onSelect={() => setRenaming(true)}>
								<Pencil /> Rename
							</DropdownMenuItem>
							<DropdownMenuItem onSelect={() => setVersionsOpen(true)}>
								<Layers /> Versions…
							</DropdownMenuItem>
							{ready && (
								<DropdownMenuItem asChild>
									<a href={primary.url} target="_blank" rel="noreferrer">
										<ExternalLink /> Open in new tab
									</a>
								</DropdownMenuItem>
							)}
							{ready && canShareFiles && (
								<DropdownMenuItem asChild>
									<a href={downloadUrl(primary)} download>
										<Download /> Download file
									</a>
								</DropdownMenuItem>
							)}
							<DropdownMenuItem onSelect={() => copy(video.originalUrl, "Source URL")}>
								<Link2 /> Copy source URL
							</DropdownMenuItem>
							<DropdownMenuItem onSelect={() => copy(video.id, "ID")}>
								<Copy /> Copy ID
							</DropdownMenuItem>
							<DropdownMenuSeparator />
							<DropdownMenuItem variant="destructive" onSelect={() => setConfirmDelete(true)}>
								<Trash2 /> Delete video
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
			</div>

			<VersionsSheet
				video={video}
				open={versionsOpen}
				onOpenChange={setVersionsOpen}
				onReencode={openEncode}
			/>
			<EncodeSheet
				key={encode.key}
				video={video}
				open={encode.open}
				sourceFileId={encode.source}
				onOpenChange={(open) => setEncode((e) => ({ ...e, open }))}
			/>
			<ConfirmDialog
				open={confirmDelete}
				onOpenChange={setConfirmDelete}
				title="Delete this video?"
				description={`"${video.name}" and all of its versions will be removed from disk.`}
				action="Delete"
				onConfirm={() => remove.mutate({ id: video.id })}
			/>
		</article>
	);
}

function Thumbnail({ video }: { video: Video }) {
	const [loaded, setLoaded] = useState(false);
	if (!video.thumbnailUrl) {
		return (
			<span className="flex size-full items-center justify-center text-muted-foreground">
				<Film className="size-12 opacity-30" />
			</span>
		);
	}
	return (
		<img
			src={video.thumbnailUrl}
			alt=""
			loading="lazy"
			decoding="async"
			onLoad={() => setLoaded(true)}
			className={cn(
				"size-full object-cover transition-[opacity,scale] duration-500 group-hover:scale-105",
				loaded ? "opacity-100" : "opacity-0",
			)}
		/>
	);
}

function RenameField({ video, onDone }: { video: Video; onDone: () => void }) {
	const qc = useQueryClient();
	const [name, setName] = useState(video.name);
	const update = useUpdateVideo({ mutation: { onSuccess: (v) => upsertVideo(qc, v) } });

	const save = () => {
		const trimmed = name.trim();
		if (trimmed && trimmed !== video.name) {
			upsertVideo(qc, { ...video, name: trimmed });
			update.mutate({ id: video.id, data: { name: trimmed } });
		}
		onDone();
	};

	return (
		<form
			className="flex items-center gap-2"
			onSubmit={(e) => {
				e.preventDefault();
				save();
			}}
		>
			<Input
				value={name}
				onChange={(e) => setName(e.target.value)}
				onKeyDown={(e) => e.key === "Escape" && onDone()}
				autoFocus
				maxLength={255}
				className="font-semibold"
			/>
			<Button
				type="submit"
				variant="secondary"
				size="icon"
				className="size-11 shrink-0 rounded-xl"
				aria-label="Save"
			>
				<Check />
			</Button>
			<Button
				type="button"
				variant="ghost"
				size="icon"
				className="size-11 shrink-0 rounded-xl"
				aria-label="Cancel"
				onClick={onDone}
			>
				<X />
			</Button>
		</form>
	);
}

function UseSourceTitle({ video }: { video: Video }) {
	const qc = useQueryClient();
	const update = useUpdateVideo({ mutation: { onSuccess: (v) => upsertVideo(qc, v) } });
	const title = video.sourceTitle ?? "";
	return (
		<button
			type="button"
			onClick={() => {
				upsertVideo(qc, { ...video, name: title });
				update.mutate({ id: video.id, data: { name: title } });
			}}
			className="mt-1 inline-flex max-w-full items-center gap-1.5 rounded-full bg-muted px-3 py-1.5 font-semibold text-xs transition-colors hover:bg-muted/70"
		>
			<Type className="size-3.5 shrink-0" />
			<span className="truncate">Use “{title}”</span>
		</button>
	);
}

function PrimaryAction({ video, file, status }: { video: Video; file?: VideoFile; status: string }) {
	const qc = useQueryClient();
	const save = useSaveToPhotos(video.name, file);
	const makeCompatible = useCreateVersion({
		mutation: {
			onSuccess: () => {
				qc.invalidateQueries({ queryKey: ["/api/videos"] });
				toast.success("Making an iPhone version");
			},
		},
	});
	const base = "h-12 flex-1 rounded-2xl font-bold text-base";

	if (!file) {
		return (
			<Button className={base} disabled>
				{activeStatuses.has(status) ? <Loader2 className="animate-spin" /> : <Download />}
				{activeStatuses.has(status) ? "Preparing…" : "Unavailable"}
			</Button>
		);
	}

	if (!canShareFiles) {
		return (
			<Button asChild className={base}>
				<a href={downloadUrl(file)} download>
					<Download /> Download
				</a>
			</Button>
		);
	}

	if (isIOS && !file.iosCompatible) {
		return (
			<Button
				className={base}
				disabled={makeCompatible.isPending}
				onClick={() =>
					makeCompatible.mutate({
						id: video.id,
						data: { sourceFileId: file.id, encoding: { goal: "compatible" }, makePrimary: true },
					})
				}
			>
				<Apple /> Make iPhone version
			</Button>
		);
	}

	switch (save.state.kind) {
		case "fetching":
			return (
				<Button className={cn(base, "relative overflow-hidden")} variant="secondary" onClick={save.cancel}>
					<span
						className="absolute inset-y-0 left-0 bg-primary/15 transition-[width]"
						style={{ width: `${Math.round(save.state.progress * 100)}%` }}
					/>
					<Loader2 className="animate-spin" />
					<span className="tabular-nums">Preparing {Math.round(save.state.progress * 100)}%</span>
				</Button>
			);
		case "ready":
			return (
				<Button className={base} onClick={save.share}>
					<Share /> Tap to save
				</Button>
			);
		default:
			return (
				<Button className={base} onClick={save.prepare}>
					<Download /> Save to Photos
				</Button>
			);
	}
}
