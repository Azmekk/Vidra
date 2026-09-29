import { useQueryClient } from "@tanstack/react-query";
import { Loader2, Wand2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import type { EncodingRequest } from "@/api/gen/model";
import { useCreateVersion } from "@/api/gen/versions/versions";
import { getGetVideoQueryKey } from "@/api/gen/videos/videos";
import { EncodingEditor } from "@/components/encoding-editor";
import { Sheet } from "@/components/sheet";
import { SwitchRow } from "@/components/switch-row";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { fileSummary } from "@/lib/format";
import type { Video } from "@/lib/videos";

export function EncodeSheet({
	video,
	open,
	onOpenChange,
	sourceFileId,
}: {
	video: Video;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	sourceFileId?: string;
}) {
	const qc = useQueryClient();
	const sources = video.files.filter((f) => f.status === "completed");
	const fallback =
		sources.find((f) => f.kind === "original") ?? sources.find((f) => f.id === video.primaryFileId);
	const [source, setSource] = useState(sourceFileId ?? fallback?.id);
	const [encoding, setEncoding] = useState<EncodingRequest>({ goal: "balanced" });
	const [makePrimary, setMakePrimary] = useState(true);
	const activeSource = sourceFileId ?? source;

	const create = useCreateVersion({
		mutation: {
			onSuccess: () => {
				qc.invalidateQueries({ queryKey: getGetVideoQueryKey(video.id) });
				toast.success("Re-encode queued");
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
					disabled={!activeSource || create.isPending}
					onClick={() =>
						create.mutate({
							id: video.id,
							data: { sourceFileId: activeSource, encoding, makePrimary },
						})
					}
				>
					{create.isPending ? <Loader2 className="animate-spin" /> : <Wand2 />}
					Start re-encode
				</Button>
			}
		>
			<div className="space-y-6">
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
				<EncodingEditor value={encoding} onChange={setEncoding} sourceFileId={activeSource} hideOriginal />
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
