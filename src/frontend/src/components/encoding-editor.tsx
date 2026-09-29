import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Apple, Gauge, Minimize2, Scale, SlidersHorizontal, Sparkles } from "lucide-react";
import type { ReactNode } from "react";
import { recommendEncoding, useGetEncodingCapabilities } from "@/api/gen/encoding/encoding";
import type {
	EncodingCapabilities,
	EncodingGoal,
	EncodingLevel,
	EncodingProfile,
	EncodingRecommendation,
	EncodingRequest,
	EncodingSource,
} from "@/api/gen/model";
import { SwitchRow } from "@/components/switch-row";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectLabel,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

const goals: { value: EncodingGoal; label: string; hint: string; icon: typeof Apple }[] = [
	{ value: "original", label: "Original", hint: "Keep as downloaded", icon: Sparkles },
	{ value: "compatible", label: "iPhone", hint: "Plays everywhere", icon: Apple },
	{ value: "balanced", label: "Balanced", hint: "Smaller, same look", icon: Scale },
	{ value: "smallest", label: "Smallest", hint: "Maximum savings", icon: Minimize2 },
	{ value: "fastest", label: "Fastest", hint: "Hardware first", icon: Gauge },
	{ value: "custom", label: "Custom", hint: "Pick everything", icon: SlidersHorizontal },
];

const levels: { value: EncodingLevel; label: string }[] = [
	{ value: "visually_lossless", label: "Lossless look" },
	{ value: "high", label: "High" },
	{ value: "balanced", label: "Balanced" },
	{ value: "compact", label: "Compact" },
	{ value: "tiny", label: "Tiny" },
];

const heights = [2160, 1440, 1080, 720, 480, 360];
const fpsCaps = [60, 30, 24];
const auto = "auto";

type Props = {
	value: EncodingRequest;
	onChange: (value: EncodingRequest) => void;
	sourceFileId?: string;
	source?: EncodingSource;
	hideOriginal?: boolean;
};

export function EncodingEditor({ value, onChange, sourceFileId, source, hideOriginal }: Props) {
	const caps = useGetEncodingCapabilities({ query: { staleTime: Number.POSITIVE_INFINITY } });
	const canRecommend = Boolean(sourceFileId || source);
	const rec = useQuery({
		queryKey: ["recommend", value.goal, value.goal === "custom" ? value.profile : null, sourceFileId, source],
		queryFn: ({ signal }) => recommendEncoding({ encoding: value, sourceFileId, source }, { signal }),
		enabled: canRecommend && value.goal !== "original",
		placeholderData: keepPreviousData,
		staleTime: Number.POSITIVE_INFINITY,
	});

	const setGoal = (goal: EncodingGoal) => {
		if (goal === "custom") {
			onChange({ goal, profile: value.profile ?? rec.data?.profile ?? defaultProfile(caps.data) });
		} else {
			onChange({ goal, profile: value.profile });
		}
	};

	return (
		<div className="space-y-5">
			<div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
				{goals
					.filter((g) => !(hideOriginal && g.value === "original"))
					.map((g) => (
						<button
							key={g.value}
							type="button"
							onClick={() => setGoal(g.value)}
							className={cn(
								"flex items-center gap-3 rounded-2xl border-2 p-3 text-left transition-colors active:scale-[0.98]",
								value.goal === g.value
									? "border-primary bg-primary/5"
									: "border-transparent bg-muted/60 hover:bg-muted",
							)}
						>
							<g.icon className="size-5 shrink-0" />
							<span className="min-w-0">
								<span className="block font-bold text-sm">{g.label}</span>
								<span className="block truncate text-muted-foreground text-xs">{g.hint}</span>
							</span>
						</button>
					))}
			</div>

			{canRecommend && value.goal !== "original" && <RecommendationCard rec={rec.data} loading={!rec.data} />}

			{value.goal === "custom" &&
				(caps.data ? (
					<CustomProfile
						caps={caps.data}
						source={source}
						profile={value.profile ?? defaultProfile(caps.data)}
						onChange={(profile) => onChange({ goal: "custom", profile })}
					/>
				) : (
					<div className="space-y-3">
						<Skeleton className="h-11 w-full rounded-xl" />
						<Skeleton className="h-11 w-full rounded-xl" />
						<Skeleton className="h-11 w-full rounded-xl" />
					</div>
				))}
		</div>
	);
}

function RecommendationCard({ rec, loading }: { rec?: EncodingRecommendation; loading: boolean }) {
	if (loading || !rec) {
		return (
			<div className="space-y-3 rounded-3xl bg-muted/50 p-5">
				<Skeleton className="h-5 w-1/2 rounded-md" />
				<Skeleton className="h-3.5 w-full rounded-md" />
				<Skeleton className="h-3.5 w-4/5 rounded-md" />
			</div>
		);
	}
	return (
		<div className="rounded-3xl bg-muted/50 p-5">
			<div className="flex flex-wrap items-center gap-2">
				<span className="font-bold">{rec.skip ? "No re-encode needed" : rec.label}</span>
				{rec.iosCompatible && (
					<Badge variant="secondary" className="rounded-full">
						<Apple /> iPhone
					</Badge>
				)}
			</div>
			<ul className="mt-2 space-y-1 text-muted-foreground text-sm">
				{rec.reasons.map((r) => (
					<li key={r}>{r}</li>
				))}
			</ul>
		</div>
	);
}

function defaultProfile(caps?: EncodingCapabilities): EncodingProfile {
	const h264 = caps?.video.find((v) => v.name === "libx264" && v.available);
	return {
		mode: "encode",
		videoEncoder: h264?.name ?? caps?.video.find((v) => v.available)?.name,
		level: "balanced",
		container: "mp4",
		audioEncoder: "aac",
	};
}

function Field({ label, children }: { label: string; children: ReactNode }) {
	return (
		<div className="space-y-2">
			<Label className="font-semibold">{label}</Label>
			{children}
		</div>
	);
}

function CustomProfile({
	caps,
	source,
	profile,
	onChange,
}: {
	caps: EncodingCapabilities;
	source?: EncodingSource;
	profile: EncodingProfile;
	onChange: (p: EncodingProfile) => void;
}) {
	const set = (patch: Partial<EncodingProfile>) => onChange({ ...profile, ...patch });
	const encoder = caps.video.find((v) => v.name === profile.videoEncoder);
	const remux = profile.mode === "remux";
	const available = caps.video.filter((v) => v.available && v.family !== "copy");
	const groups = [
		{ label: "Recommended", items: available.filter((v) => v.curated && !v.hardware) },
		{ label: "Hardware (detected)", items: available.filter((v) => v.hardware) },
		{ label: "All available", items: available.filter((v) => !v.curated) },
	].filter((g) => g.items.length);
	const containers = caps.containers.filter(
		(c) => remux || !encoder?.curated || encoder.containers.includes(c),
	);
	const audio = caps.audio.filter((a) => !a.curated || a.containers.includes(profile.container));
	const maxHeights = heights.filter((h) => !source?.height || h <= source.height);
	const maxFps = fpsCaps.filter((f) => !source?.fps || f < source.fps);

	const pickEncoder = (name: string) => {
		const next = caps.video.find((v) => v.name === name);
		const container =
			next?.curated && !next.containers.includes(profile.container) ? next.containers[0] : profile.container;
		set({ videoEncoder: name, preset: next?.defaultPreset, quality: undefined, container });
	};

	return (
		<div className="space-y-5 rounded-3xl border p-5">
			<SwitchRow
				label="Remux only"
				description="Change the container without re-encoding"
				checked={remux}
				onCheckedChange={(on) => set({ mode: on ? "remux" : "encode" })}
			/>

			{!remux && (
				<>
					<Field label="Video encoder">
						<Select value={profile.videoEncoder} onValueChange={pickEncoder}>
							<SelectTrigger className="h-11! w-full rounded-xl">
								<SelectValue placeholder="Choose an encoder" />
							</SelectTrigger>
							<SelectContent>
								{groups.map((g) => (
									<SelectGroup key={g.label}>
										<SelectLabel>{g.label}</SelectLabel>
										{g.items.map((v) => (
											<SelectItem key={v.name} value={v.name}>
												{v.label}
												{v.ios && <Apple className="text-muted-foreground" />}
											</SelectItem>
										))}
									</SelectGroup>
								))}
							</SelectContent>
						</Select>
					</Field>

					{encoder?.quality && (
						<Field label="Quality">
							<div className="flex flex-wrap gap-2">
								{levels.map((l) => (
									<button
										key={l.value}
										type="button"
										onClick={() => set({ level: l.value, quality: undefined })}
										className={cn(
											"rounded-full px-3.5 py-1.5 font-semibold text-sm transition-colors",
											profile.level === l.value && profile.quality === undefined
												? "bg-primary text-primary-foreground"
												: "bg-muted hover:bg-muted/70",
										)}
									>
										{l.label}
									</button>
								))}
							</div>
							<Input
								type="number"
								inputMode="numeric"
								min={encoder.quality.min}
								max={encoder.quality.max}
								placeholder={`${profile.level ? encoder.quality.levels[profile.level] : ""} (${encoder.quality.min}–${encoder.quality.max}, ${encoder.quality.lowerIsBetter ? "lower is better" : "higher is better"})`}
								value={profile.quality ?? ""}
								onChange={(e) => set({ quality: e.target.value === "" ? undefined : Number(e.target.value) })}
							/>
						</Field>
					)}

					{!encoder?.quality && (
						<Field label="Video bitrate (kbps)">
							<Input
								type="number"
								inputMode="numeric"
								min={100}
								value={profile.videoBitrate ?? ""}
								onChange={(e) =>
									set({ videoBitrate: e.target.value === "" ? undefined : Number(e.target.value) })
								}
							/>
						</Field>
					)}

					<div className="grid grid-cols-2 gap-4">
						{encoder?.presets?.length ? (
							<Field label="Speed">
								<Select
									value={profile.preset ?? auto}
									onValueChange={(v) => set({ preset: v === auto ? undefined : v })}
								>
									<SelectTrigger className="h-11! w-full rounded-xl">
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value={auto}>Default</SelectItem>
										{encoder.presets.map((p) => (
											<SelectItem key={p} value={p}>
												{p}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</Field>
						) : null}
						<Field label="Max resolution">
							<Select
								value={profile.maxHeight ? String(profile.maxHeight) : auto}
								onValueChange={(v) => set({ maxHeight: v === auto ? undefined : Number(v) })}
							>
								<SelectTrigger className="h-11! w-full rounded-xl">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value={auto}>Source</SelectItem>
									{maxHeights.map((h) => (
										<SelectItem key={h} value={String(h)}>
											{h}p
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</Field>
						<Field label="Max frame rate">
							<Select
								value={profile.maxFps ? String(profile.maxFps) : auto}
								onValueChange={(v) => set({ maxFps: v === auto ? undefined : Number(v) })}
							>
								<SelectTrigger className="h-11! w-full rounded-xl">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value={auto}>Source</SelectItem>
									{maxFps.map((f) => (
										<SelectItem key={f} value={String(f)}>
											{f} fps
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</Field>
					</div>
				</>
			)}

			<div className="grid grid-cols-2 gap-4">
				<Field label="Container">
					<Select value={profile.container} onValueChange={(container) => set({ container })}>
						<SelectTrigger className="h-11! w-full rounded-xl">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{containers.map((c) => (
								<SelectItem key={c} value={c}>
									{c.toUpperCase()}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</Field>
				{!remux && (
					<Field label="Audio">
						<Select
							value={profile.audioEncoder ?? "copy"}
							onValueChange={(v) => {
								const a = caps.audio.find((x) => x.name === v);
								set({ audioEncoder: v, audioBitrate: a?.defaultBitrate });
							}}
						>
							<SelectTrigger className="h-11! w-full rounded-xl">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{audio.map((a) => (
									<SelectItem key={a.name} value={a.name}>
										{a.label}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
				)}
			</div>

			{!remux && caps.audio.find((a) => a.name === profile.audioEncoder)?.defaultBitrate && (
				<Field label="Audio bitrate (kbps)">
					<Input
						type="number"
						inputMode="numeric"
						min={32}
						max={512}
						value={profile.audioBitrate ?? ""}
						onChange={(e) =>
							set({ audioBitrate: e.target.value === "" ? undefined : Number(e.target.value) })
						}
					/>
				</Field>
			)}
		</div>
	);
}
