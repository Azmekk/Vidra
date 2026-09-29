import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { Check, Copy, KeyRound, Loader2, Monitor, Moon, RefreshCw, Sun, Trash2 } from "lucide-react";
import { type ReactNode, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import {
	getListApiTokensQueryKey,
	useChangePassword,
	useCreateApiToken,
	useDeleteApiToken,
	useListApiTokens,
} from "@/api/gen/auth/auth";
import { useGetEncodingCapabilities } from "@/api/gen/encoding/encoding";
import type { HandlersCreateAPITokenResponse, ServicesSettings } from "@/api/gen/model";
import { getGetSettingsQueryKey, useGetSettings, useUpdateSettings } from "@/api/gen/settings/settings";
import { useGetSystemInfo } from "@/api/gen/system/system";
import { useUpdateYtdlp } from "@/api/gen/ytdlp/ytdlp";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { EncodingEditor } from "@/components/encoding-editor";
import { Section } from "@/components/section";
import { SectionSkeleton } from "@/components/skeletons";
import { SwitchRow } from "@/components/switch-row";
import { Button } from "@/components/ui/button";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { formatBytes, formatRelative } from "@/lib/format";
import { type Theme, useTheme } from "@/lib/theme";
import { cn } from "@/lib/utils";

export function SettingsPage() {
	const settings = useGetSettings();

	return (
		<div className="space-y-8">
			<div>
				<h1 className="font-extrabold text-4xl tracking-tight">Settings</h1>
				<p className="mt-1 font-medium text-lg text-muted-foreground">Account, downloads and encoding.</p>
			</div>
			<AccountSection />
			<TokensSection />
			<AppearanceSection />
			{settings.data ? (
				<ServerSettings settings={settings.data} />
			) : (
				<>
					<SectionSkeleton rows={4} />
					<SectionSkeleton rows={1} />
					<SectionSkeleton rows={3} />
				</>
			)}
			<SystemSection />
		</div>
	);
}

function useSaveSettings() {
	const qc = useQueryClient();
	const key = getGetSettingsQueryKey();
	const update = useUpdateSettings({
		mutation: {
			onSuccess: (data) => qc.setQueryData(key, data),
			onError: () => qc.invalidateQueries({ queryKey: key }),
		},
	});
	return (patch: Partial<ServicesSettings>) => {
		const current = qc.getQueryData<ServicesSettings>(key);
		if (!current) return;
		const next = { ...current, ...patch };
		qc.setQueryData(key, next);
		update.mutate({ data: next });
	};
}

function ServerSettings({ settings }: { settings: ServicesSettings }) {
	const save = useSaveSettings();

	return (
		<>
			<Section title="Downloads" description="How new downloads are fetched and stored.">
				<SwitchRow
					label="Prefer iPhone-compatible formats"
					description="Pick H.264/AAC when available so most downloads need no re-encode."
					checked={settings.preferCompatibleFormats}
					onCheckedChange={(v) => save({ preferCompatibleFormats: v })}
				/>
				<SwitchRow
					label="Keep original"
					description="Keep the downloaded file next to encoded versions so you can re-encode later."
					checked={settings.keepOriginal}
					onCheckedChange={(v) => save({ keepOriginal: v })}
				/>
				<div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
					<NumberField
						label="Parallel downloads"
						value={settings.maxConcurrentDownloads}
						min={1}
						max={10}
						onCommit={(v) => save({ maxConcurrentDownloads: v })}
					/>
					<NumberField
						label="Parallel encodes"
						value={settings.maxConcurrentEncodes}
						min={1}
						max={4}
						onCommit={(v) => save({ maxConcurrentEncodes: v })}
					/>
					<NumberField
						label="Cached videos"
						value={settings.cacheSize}
						min={0}
						max={5000}
						onCommit={(v) => save({ cacheSize: v })}
					/>
				</div>
			</Section>

			<Section title="Network" description="Route yt-dlp through a proxy, e.g. socks5://host:1080.">
				<TextField
					label="Proxy URL"
					value={settings.proxyUrl}
					placeholder="None"
					onCommit={(v) => save({ proxyUrl: v })}
				/>
			</Section>

			<Section
				title="Default encoding"
				description="Applied to quick downloads and preselected on the download page."
			>
				<EncodingEditor value={settings.defaultEncoding} onChange={(v) => save({ defaultEncoding: v })} />
			</Section>
		</>
	);
}

function NumberField({
	label,
	value,
	min,
	max,
	onCommit,
}: {
	label: string;
	value: number;
	min: number;
	max: number;
	onCommit: (v: number) => void;
}) {
	const [text, setText] = useState(String(value));
	const commit = () => {
		const n = Math.min(max, Math.max(min, Math.round(Number(text))));
		if (Number.isNaN(n)) return setText(String(value));
		setText(String(n));
		if (n !== value) onCommit(n);
	};
	return (
		<div className="space-y-2">
			<Label className="font-semibold">{label}</Label>
			<Input
				type="number"
				inputMode="numeric"
				min={min}
				max={max}
				value={text}
				onChange={(e) => setText(e.target.value)}
				onBlur={commit}
				onKeyDown={(e) => e.key === "Enter" && e.currentTarget.blur()}
			/>
		</div>
	);
}

function TextField({
	label,
	value,
	placeholder,
	onCommit,
}: {
	label: string;
	value: string;
	placeholder?: string;
	onCommit: (v: string) => void;
}) {
	const [text, setText] = useState(value);
	return (
		<div className="space-y-2">
			<Label className="font-semibold">{label}</Label>
			<Input
				value={text}
				placeholder={placeholder}
				autoCapitalize="none"
				autoCorrect="off"
				onChange={(e) => setText(e.target.value)}
				onBlur={() => text.trim() !== value && onCommit(text.trim())}
				onKeyDown={(e) => e.key === "Enter" && e.currentTarget.blur()}
			/>
		</div>
	);
}

const passwordSchema = z
	.object({
		currentPassword: z.string().min(1, "Enter your current password"),
		newPassword: z.string().min(8, "At least 8 characters").max(256),
		confirm: z.string(),
	})
	.refine((v) => v.newPassword === v.confirm, { path: ["confirm"], message: "Passwords don't match" });

function AccountSection() {
	const form = useForm({
		resolver: zodResolver(passwordSchema),
		defaultValues: { currentPassword: "", newPassword: "", confirm: "" },
	});
	const change = useChangePassword({
		mutation: {
			onSuccess: () => {
				form.reset();
				toast.success("Password changed. Other devices were signed out.");
			},
			onError: (e) => form.setError("currentPassword", { message: e.message }),
		},
	});

	const fields = [
		{ name: "currentPassword", label: "Current password", autoComplete: "current-password" },
		{ name: "newPassword", label: "New password", autoComplete: "new-password" },
		{ name: "confirm", label: "Confirm new password", autoComplete: "new-password" },
	] as const;

	return (
		<Section title="Password" description="Changing it signs out every other device.">
			<Form {...form}>
				<form
					onSubmit={form.handleSubmit(({ confirm: _, ...data }) => change.mutate({ data }))}
					className="space-y-4"
				>
					{fields.map((f) => (
						<FormField
							key={f.name}
							control={form.control}
							name={f.name}
							render={({ field }) => (
								<FormItem>
									<FormLabel>{f.label}</FormLabel>
									<FormControl>
										<Input type="password" autoComplete={f.autoComplete} {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
					))}
					<Button type="submit" className="h-11 rounded-xl font-bold" disabled={change.isPending}>
						{change.isPending && <Loader2 className="animate-spin" />}
						Change password
					</Button>
				</form>
			</Form>
		</Section>
	);
}

function TokensSection() {
	const qc = useQueryClient();
	const tokens = useListApiTokens();
	const [name, setName] = useState("");
	const [created, setCreated] = useState<HandlersCreateAPITokenResponse>();
	const [revoke, setRevoke] = useState<string>();
	const [copied, setCopied] = useState(false);
	const refresh = () => qc.invalidateQueries({ queryKey: getListApiTokensQueryKey() });
	const create = useCreateApiToken({
		mutation: {
			onSuccess: (t) => {
				setCreated(t);
				setName("");
				setCopied(false);
				refresh();
			},
		},
	});
	const remove = useDeleteApiToken({ mutation: { onSuccess: refresh } });

	return (
		<Section
			title="API tokens"
			description="For the iOS Shortcut and scripts. Send as “Authorization: Bearer <token>”."
		>
			<form
				className="flex gap-2"
				onSubmit={(e) => {
					e.preventDefault();
					if (name.trim()) create.mutate({ data: { name: name.trim() } });
				}}
			>
				<Input
					placeholder="Token name, e.g. iPhone Shortcut"
					value={name}
					onChange={(e) => setName(e.target.value)}
					maxLength={64}
				/>
				<Button
					type="submit"
					className="h-11 shrink-0 rounded-xl font-bold"
					disabled={!name.trim() || create.isPending}
				>
					{create.isPending ? <Loader2 className="animate-spin" /> : <KeyRound />}
					Create
				</Button>
			</form>

			{created && (
				<div className="space-y-3 rounded-3xl bg-primary/5 p-4 ring-1 ring-primary/20">
					<p className="font-semibold text-sm">Copy “{created.name}” now. It won't be shown again.</p>
					<div className="flex gap-2">
						<code className="min-w-0 flex-1 truncate rounded-xl bg-muted px-3 py-2.5 font-mono text-sm">
							{created.token}
						</code>
						<Button
							variant="secondary"
							size="icon"
							className="size-11 shrink-0 rounded-xl"
							aria-label="Copy token"
							onClick={async () => {
								await navigator.clipboard.writeText(created.token);
								setCopied(true);
							}}
						>
							{copied ? <Check /> : <Copy />}
						</Button>
					</div>
				</div>
			)}

			{tokens.data ? (
				tokens.data.length > 0 && (
					<ul className="divide-y rounded-3xl border">
						{tokens.data.map((t) => (
							<li key={t.id} className="flex items-center gap-3 px-4 py-3">
								<div className="min-w-0 flex-1">
									<p className="truncate font-semibold">{t.name}</p>
									<p className="text-muted-foreground text-xs">
										Created {formatRelative(t.createdAt)} ·{" "}
										{t.lastUsedAt ? `used ${formatRelative(t.lastUsedAt)}` : "never used"}
									</p>
								</div>
								<Button
									variant="ghost"
									size="icon"
									className="size-9 shrink-0 rounded-xl text-destructive"
									aria-label="Revoke"
									onClick={() => setRevoke(t.id)}
								>
									<Trash2 />
								</Button>
							</li>
						))}
					</ul>
				)
			) : (
				<div className="space-y-2">
					<Skeleton className="h-14 w-full rounded-2xl" />
					<Skeleton className="h-14 w-full rounded-2xl" />
				</div>
			)}

			<ConfirmDialog
				open={Boolean(revoke)}
				onOpenChange={(open) => !open && setRevoke(undefined)}
				title="Revoke this token?"
				description="Anything using it will stop working immediately."
				action="Revoke"
				onConfirm={() => revoke && remove.mutate({ id: revoke })}
			/>
		</Section>
	);
}

const themes: { value: Theme; label: string; icon: typeof Sun }[] = [
	{ value: "light", label: "Light", icon: Sun },
	{ value: "dark", label: "Dark", icon: Moon },
	{ value: "system", label: "System", icon: Monitor },
];

function AppearanceSection() {
	const { theme, setTheme } = useTheme();
	return (
		<Section title="Appearance" description="Saved on this device.">
			<div className="grid grid-cols-3 gap-2">
				{themes.map((t) => (
					<button
						key={t.value}
						type="button"
						onClick={() => setTheme(t.value)}
						className={cn(
							"flex flex-col items-center gap-2 rounded-2xl border-2 py-4 font-semibold text-sm transition-colors",
							theme === t.value
								? "border-primary bg-primary/5"
								: "border-transparent bg-muted/60 hover:bg-muted",
						)}
					>
						<t.icon className="size-5" />
						{t.label}
					</button>
				))}
			</div>
		</Section>
	);
}

function SystemSection() {
	const info = useGetSystemInfo();
	const caps = useGetEncodingCapabilities({ query: { staleTime: Number.POSITIVE_INFINITY } });
	const [output, setOutput] = useState<string>();
	const update = useUpdateYtdlp({
		mutation: {
			onSuccess: (r) => {
				setOutput(r.output);
				toast.success("yt-dlp updated");
			},
		},
	});
	const hw = caps.data?.video.filter((v) => v.hardware && v.available).map((v) => v.label) ?? [];

	return (
		<Section title="System">
			<dl className="grid grid-cols-2 gap-4 text-sm">
				<Stat
					label="Storage used"
					value={info.data ? formatBytes(info.data.downloadsSize) || "0 B" : undefined}
				/>
				<Stat label="Vidra" value={__APP_VERSION__} />
				<Stat label="ffmpeg" value={caps.data?.ffmpegVersion.split("-")[0]} />
				<Stat label="Hardware encoders" value={caps.data ? hw.join(", ") || "None" : undefined} />
			</dl>
			<Button
				variant="secondary"
				className="h-11 w-full rounded-xl font-bold"
				disabled={update.isPending}
				onClick={() => update.mutate()}
			>
				<RefreshCw className={cn(update.isPending && "animate-spin")} />
				{update.isPending ? "Updating yt-dlp…" : "Update yt-dlp"}
			</Button>
			{output && (
				<pre className="max-h-60 overflow-auto whitespace-pre-wrap rounded-2xl bg-muted p-4 font-mono text-xs">
					{output}
				</pre>
			)}
		</Section>
	);
}

function Stat({ label, value }: { label: string; value?: ReactNode }) {
	return (
		<div className="min-w-0">
			<dt className="text-muted-foreground">{label}</dt>
			<dd className="mt-0.5 truncate font-semibold">
				{value ?? <Skeleton className="mt-1 h-4 w-20 rounded-md" />}
			</dd>
		</div>
	);
}
