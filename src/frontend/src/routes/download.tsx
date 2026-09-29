import { useQueryClient } from "@tanstack/react-query";
import { getRouteApi, useNavigate } from "@tanstack/react-router";
import { Apple, ClipboardPaste, Download, Info, Loader2, Sparkles, Zap } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";
import type { EncodingRequest, ServicesVideoMetadata, ServicesVideoOption } from "@/api/gen/model";
import { useGetSettings } from "@/api/gen/settings/settings";
import { useCreateVideo, useGetMetadata, useQuickDownload } from "@/api/gen/videos/videos";
import { EncodingEditor } from "@/components/encoding-editor";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { codecName, formatBytes, formatDuration } from "@/lib/format";
import { cleanName, nameError } from "@/lib/names";
import { extractUrl } from "@/lib/url";
import { cn } from "@/lib/utils";
import { upsertVideo } from "@/lib/videos";

const route = getRouteApi("/app/download");
const urlSchema = z.url({ protocol: /^https?$/, error: "Enter a valid http(s) link" });

export function DownloadPage() {
	const search = route.useSearch();
	const navigate = useNavigate();
	const qc = useQueryClient();
	const shared = search.url ?? search.text ?? search.title;
	const [url, setUrl] = useState(shared ? extractUrl(shared) : "");
	const [error, setError] = useState<string>();
	const metadata = useGetMetadata();
	const quick = useQuickDownload({
		mutation: {
			onSuccess: (video) => {
				upsertVideo(qc, video);
				toast.success("Download started");
				navigate({ to: "/" });
			},
		},
	});

	const autoStarted = useRef(false);
	useEffect(() => {
		if (autoStarted.current || !search.quick || !shared) return;
		autoStarted.current = true;
		quick.mutate({ data: { url: shared } });
	}, [search.quick, shared, quick]);

	const validate = () => {
		const clean = extractUrl(url);
		setUrl(clean);
		const parsed = urlSchema.safeParse(clean);
		setError(parsed.success ? undefined : parsed.error.issues[0]?.message);
		return parsed.success ? parsed.data : undefined;
	};

	const paste = async () => {
		try {
			const text = extractUrl(await navigator.clipboard.readText());
			setUrl(text);
			setError(undefined);
			metadata.reset();
		} catch {
			toast.error("Clipboard access was blocked");
		}
	};

	return (
		<div className="space-y-8">
			<div>
				<h1 className="font-extrabold text-4xl tracking-tight">Download</h1>
				<p className="mt-1 font-medium text-lg text-muted-foreground">
					Paste a link from YouTube, TikTok, Instagram and more.
				</p>
			</div>

			<form
				className="space-y-4 rounded-[2.5rem] border bg-card p-6"
				onSubmit={(e) => {
					e.preventDefault();
					const valid = validate();
					if (valid) quick.mutate({ data: { url: valid } });
				}}
			>
				<div className="space-y-2">
					<div className="flex gap-2">
						<Input
							type="url"
							inputMode="url"
							autoComplete="off"
							placeholder="https://…"
							value={url}
							aria-invalid={Boolean(error)}
							onPaste={(e) => {
								e.preventDefault();
								setUrl(extractUrl(e.clipboardData.getData("text")));
								setError(undefined);
								metadata.reset();
							}}
							onChange={(e) => {
								setUrl(e.target.value);
								setError(undefined);
								metadata.reset();
							}}
							className="h-12 rounded-2xl"
						/>
						<Button
							type="button"
							variant="secondary"
							size="icon"
							className="size-12 shrink-0 rounded-2xl"
							aria-label="Paste"
							onClick={paste}
						>
							<ClipboardPaste className="size-5" />
						</Button>
					</div>
					{error && <p className="px-1 font-medium text-destructive text-sm">{error}</p>}
				</div>
				<div className="grid gap-2 sm:grid-cols-2">
					<Button
						type="submit"
						size="lg"
						className="h-12 rounded-2xl font-bold text-base"
						disabled={quick.isPending}
					>
						{quick.isPending ? <Loader2 className="animate-spin" /> : <Zap />}
						Quick download
					</Button>
					<Button
						type="button"
						variant="secondary"
						size="lg"
						className="h-12 rounded-2xl font-bold text-base"
						disabled={metadata.isPending}
						onClick={() => {
							const valid = validate();
							if (valid) metadata.mutate({ data: { url: valid } });
						}}
					>
						{metadata.isPending ? <Loader2 className="animate-spin" /> : <Info />}
						Fetch info
					</Button>
				</div>
			</form>

			{metadata.isPending && <MetadataSkeleton />}
			{metadata.data && <FormatPicker key={url} url={extractUrl(url)} meta={metadata.data} />}
		</div>
	);
}

function MetadataSkeleton() {
	return (
		<div className="space-y-6 rounded-[2.5rem] border bg-card p-6">
			<div className="flex gap-4">
				<Skeleton className="aspect-video w-36 shrink-0 rounded-2xl" />
				<div className="flex-1 space-y-2 pt-1">
					<Skeleton className="h-5 w-full rounded-md" />
					<Skeleton className="h-5 w-2/3 rounded-md" />
					<Skeleton className="h-4 w-1/3 rounded-md" />
				</div>
			</div>
			<Skeleton className="h-11 w-full rounded-xl" />
			<div className="space-y-2">
				<Skeleton className="h-16 w-full rounded-2xl" />
				<Skeleton className="h-16 w-full rounded-2xl" />
				<Skeleton className="h-16 w-full rounded-2xl" />
			</div>
		</div>
	);
}

function FormatPicker({ url, meta }: { url: string; meta: ServicesVideoMetadata }) {
	const navigate = useNavigate();
	const qc = useQueryClient();
	const settings = useGetSettings();
	const [name, setName] = useState(() => cleanName(meta.title));
	const nameInvalid = nameError(name);
	const [formatId, setFormatId] = useState<string>();
	const [encoding, setEncoding] = useState<EncodingRequest>();
	const option = meta.options.find((o) => o.formatId === formatId);
	const value = encoding ?? settings.data?.defaultEncoding ?? { goal: "original" };

	const create = useCreateVideo({
		mutation: {
			onSuccess: (video) => {
				upsertVideo(qc, video);
				toast.success("Download started");
				navigate({ to: "/" });
			},
		},
	});

	return (
		<div className="space-y-6 rounded-[2.5rem] border bg-card p-6">
			<div className="flex gap-4">
				{meta.thumbnail ? (
					<img
						src={meta.thumbnail}
						alt=""
						referrerPolicy="no-referrer"
						className="aspect-video w-36 shrink-0 rounded-2xl bg-muted object-cover"
					/>
				) : (
					<Skeleton className="aspect-video w-36 shrink-0 animate-none rounded-2xl" />
				)}
				<div className="min-w-0">
					<p className="line-clamp-2 font-bold leading-snug">{meta.title}</p>
					<p className="mt-1 text-muted-foreground text-sm">
						{[meta.uploader, formatDuration(meta.duration), meta.extractor].filter(Boolean).join(" · ")}
					</p>
				</div>
			</div>

			<div className="space-y-2">
				<Label htmlFor="name" className="font-semibold">
					Name
				</Label>
				<Input
					id="name"
					value={name}
					onChange={(e) => setName(e.target.value)}
					maxLength={255}
					aria-invalid={Boolean(nameInvalid)}
				/>
				{nameInvalid && <p className="font-medium text-destructive text-sm">{nameInvalid}</p>}
			</div>

			<div className="space-y-2">
				<Label className="font-semibold">Format</Label>
				<div className="max-h-80 space-y-2 overflow-y-auto overscroll-contain">
					<FormatOption selected={!formatId} onSelect={() => setFormatId(undefined)}>
						<span className="flex items-center gap-2 font-bold">
							<Sparkles className="size-4" /> Best available
						</span>
						<span className="text-muted-foreground text-sm">Highest quality video and audio</span>
					</FormatOption>
					{meta.options.map((o) => (
						<FormatOption
							key={o.formatId}
							selected={formatId === o.formatId}
							onSelect={() => setFormatId(o.formatId)}
						>
							<OptionLabel option={o} />
						</FormatOption>
					))}
				</div>
			</div>

			<div className="space-y-2">
				<Label className="font-semibold">Encoding</Label>
				{settings.data ? (
					<EncodingEditor
						value={value}
						onChange={setEncoding}
						source={option ? toSource(option, meta) : undefined}
					/>
				) : (
					<Skeleton className="h-40 w-full rounded-2xl" />
				)}
			</div>

			<Button
				size="lg"
				className="h-12 w-full rounded-2xl font-bold text-base"
				disabled={create.isPending || Boolean(nameInvalid)}
				onClick={() =>
					create.mutate({
						data: { url, name: name.trim(), sourceTitle: meta.title, formatId, encoding: value },
					})
				}
			>
				{create.isPending ? <Loader2 className="animate-spin" /> : <Download />}
				Download
			</Button>
		</div>
	);
}

function FormatOption({
	selected,
	onSelect,
	children,
}: {
	selected: boolean;
	onSelect: () => void;
	children: React.ReactNode;
}) {
	return (
		<button
			type="button"
			onClick={onSelect}
			className={cn(
				"flex w-full flex-col items-start gap-0.5 rounded-2xl border-2 px-4 py-3 text-left transition-colors",
				selected ? "border-primary bg-primary/5" : "border-transparent bg-muted/60 hover:bg-muted",
			)}
		>
			{children}
		</button>
	);
}

function OptionLabel({ option: o }: { option: ServicesVideoOption }) {
	return (
		<>
			<span className="flex flex-wrap items-center gap-2 font-bold">
				{o.resolution}
				{o.fps ? <span className="font-medium text-muted-foreground">{Math.round(o.fps)}fps</span> : null}
				{o.iosCompatible && (
					<Badge variant="secondary" className="rounded-full">
						<Apple /> iPhone
					</Badge>
				)}
				{o.dynamicRange && o.dynamicRange !== "SDR" && (
					<Badge variant="secondary" className="rounded-full">
						{o.dynamicRange}
					</Badge>
				)}
			</span>
			<span className="text-muted-foreground text-sm">
				{[
					codecName(o.videoCodec),
					o.hasAudio ? codecName(o.audioCodec) : "no audio",
					o.extension.toUpperCase(),
					formatBytes(o.fileSize),
					o.note,
				]
					.filter(Boolean)
					.join(" · ")}
			</span>
		</>
	);
}

function toSource(o: ServicesVideoOption, meta: ServicesVideoMetadata) {
	return {
		videoCodec: o.videoCodec,
		audioCodec: o.audioCodec,
		container: o.extension,
		width: o.width ?? 0,
		height: o.height ?? 0,
		fps: o.fps ?? 0,
		bitrate: Math.round((o.bitrate ?? 0) * 1000),
		duration: meta.duration,
	};
}
