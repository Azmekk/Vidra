import { keepPreviousData, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Music, Wand2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { recommendEncoding } from "@/api/gen/encoding/encoding";
import type { EncodingRequest } from "@/api/gen/model";
import { useCreateVersion } from "@/api/gen/versions/versions";
import { getGetVideoQueryKey } from "@/api/gen/videos/videos";
import { AudioFormatPicker } from "@/components/audio-format-picker";
import { EncodingEditor, RecommendationCard } from "@/components/encoding-editor";
import { Sheet } from "@/components/sheet";
import { SwitchRow } from "@/components/switch-row";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { fileSummary } from "@/lib/format";
import type { Video } from "@/lib/videos";

type Mode = "video" | "audio";

export function EncodeSheet({
	video,
	open,
	onOpenChange,
	sourceFileId,
	audio,
}: {
	video: Video;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	sourceFileId?: string;
	audio?: boolean;
}) {
	const qc = useQueryClient();
	const [mode, setMode] = useState<Mode>(audio ? "audio" : "video");
	const sources = video.files.filter((f) => f.status === "completed" && (mode === "video" || f.audioCodec));
	const fallback =
		sources.find((f) => f.kind === "original") ??
		sources.find((f) => f.id === video.primaryFileId) ??
		sources[0];
	const [source, setSource] = useState(sourceFileId);
	const [encoding, setEncoding] = useState<EncodingRequest>({ goal: "balanced" });
	const [audioContainer, setAudioContainer] = useState("mp3");
	const [makePrimary, setMakePrimary] = useState(!audio);
	const activeSource = sources.find((f) => f.id === source)?.id ?? fallback?.id;
	const request: EncodingRequest =
		mode === "audio" ? { goal: "audio", profile: { mode: "audio", container: audioContainer } } : encoding;

	const rec = useQuery({
		queryKey: ["recommend", "audio", audioContainer, activeSource],
		queryFn: ({ signal }) => recommendEncoding({ encoding: request, sourceFileId: activeSource }, { signal }),
		enabled: mode === "audio" && Boolean(activeSource),
		placeholderData: keepPreviousData,
		staleTime: Number.POSITIVE_INFINITY,
	});

	const create = useCreateVersion({
		mutation: {
			onSuccess: () => {
				qc.invalidateQueries({ queryKey: getGetVideoQueryKey(video.id) });
				toast.success(mode === "audio" ? "Audio conversion queued" : "Re-encode queued");
				onOpenChange(false);
			},
		},
	});

	return (
		<Sheet
			open={open}
			onOpenChange={onOpenChange}
			title="New version"
			description={video.name}
			footer={
				<Button
					size="lg"
					className="h-12 w-full rounded-2xl font-bold text-base"
					disabled={!activeSource || create.isPending || (mode === "audio" && rec.data?.skip)}
					onClick={() =>
						create.mutate({
							id: video.id,
							data: { sourceFileId: activeSource, encoding: request, makePrimary },
						})
					}
				>
					{create.isPending ? <Loader2 className="animate-spin" /> : mode === "audio" ? <Music /> : <Wand2 />}
					{mode === "audio" ? "Convert to audio" : "Start re-encode"}
				</Button>
			}
		>
			<div className="space-y-6">
				<Tabs
					value={mode}
					onValueChange={(v) => {
						setMode(v as Mode);
						setMakePrimary(v === "video");
					}}
				>
					<TabsList className="h-11 w-full rounded-2xl">
						<TabsTrigger value="video" className="rounded-xl font-semibold">
							<Wand2 /> Video
						</TabsTrigger>
						<TabsTrigger value="audio" className="rounded-xl font-semibold">
							<Music /> Audio
						</TabsTrigger>
					</TabsList>
				</Tabs>
				{sources.length > 1 && (
					<div className="space-y-2">
						<Label className="font-semibold">Source version</Label>
						<Select value={activeSource} onValueChange={setSource}>
							<SelectTrigger className="h-auto! min-h-11 w-full rounded-xl py-2 text-left">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{sources.map((f) => (
									<SelectItem key={f.id} value={f.id}>
										<span className="flex flex-col">
											<span className="font-semibold">{f.label}</span>
											<span className="text-muted-foreground text-xs">{fileSummary(f)}</span>
										</span>
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				)}
				{mode === "audio" ? (
					activeSource ? (
						<div className="space-y-5">
							<AudioFormatPicker value={audioContainer} onChange={setAudioContainer} />
							<RecommendationCard rec={rec.data} loading={!rec.data} />
						</div>
					) : (
						<p className="rounded-2xl bg-muted/50 p-4 text-muted-foreground text-sm">
							No finished version has an audio track.
						</p>
					)
				) : (
					<EncodingEditor value={encoding} onChange={setEncoding} sourceFileId={activeSource} hideOriginal />
				)}
				<SwitchRow
					label="Make default when done"
					checked={makePrimary}
					onCheckedChange={setMakePrimary}
					className="rounded-2xl bg-muted/50 p-4"
				/>
			</div>
		</Sheet>
	);
}
