import { useQueryClient } from "@tanstack/react-query";
import {
	CircleAlert,
	CircleCheck,
	Cloud,
	HardDrive,
	Loader2,
	Pencil,
	Play,
	Plug,
	Plus,
	Trash2,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { z } from "zod";
import {
	getListBackupTargetsQueryKey,
	useCreateBackupTarget,
	useDeleteBackupTarget,
	useListBackupTargets,
	useRunBackupTarget,
	useTestBackupTarget,
	useUpdateBackupTarget,
} from "@/api/gen/backups/backups";
import type { ServicesBackupTargetDTO, ServicesBackupTargetInput } from "@/api/gen/model";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Section } from "@/components/section";
import { Sheet } from "@/components/sheet";
import { SwitchRow } from "@/components/switch-row";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { formatRelative } from "@/lib/format";
import { cn } from "@/lib/utils";

type Provider = "s3" | "drive" | "mega" | "local";

type Field = {
	key: string;
	label: string;
	required?: boolean;
	secret?: boolean;
	multiline?: boolean;
	placeholder?: string;
	help?: string;
	options?: string[];
};

const providers: Record<
	Provider,
	{ label: string; pathLabel: string; pathPlaceholder: string; fields: Field[] }
> = {
	s3: {
		label: "S3 / R2",
		pathLabel: "Bucket and folder",
		pathPlaceholder: "my-bucket/vidra",
		fields: [
			{
				key: "provider",
				label: "Service",
				required: true,
				options: ["AWS", "Cloudflare", "Wasabi", "Minio", "Other"],
			},
			{
				key: "endpoint",
				label: "Endpoint",
				placeholder: "https://<account>.r2.cloudflarestorage.com",
				help: "Leave empty for AWS.",
			},
			{ key: "region", label: "Region", placeholder: "auto" },
			{ key: "access_key_id", label: "Access key ID", required: true },
			{ key: "secret_access_key", label: "Secret access key", required: true, secret: true },
		],
	},
	drive: {
		label: "Google Drive",
		pathLabel: "Folder",
		pathPlaceholder: "Vidra",
		fields: [
			{
				key: "client_id",
				label: "Client ID",
				help: "Optional, but Google's shared ID is heavily rate limited.",
			},
			{ key: "client_secret", label: "Client secret", secret: true },
			{
				key: "token",
				label: "Token",
				required: true,
				secret: true,
				multiline: true,
				placeholder: '{"access_token":"…","refresh_token":"…"}',
				help: 'On any computer with rclone, run: rclone authorize "drive" (add your client ID and secret if you use them), sign in, then paste the JSON it prints.',
			},
			{ key: "root_folder_id", label: "Root folder ID", help: "Optional. Limits access to one folder." },
		],
	},
	mega: {
		label: "MEGA",
		pathLabel: "Folder",
		pathPlaceholder: "Vidra",
		fields: [
			{ key: "user", label: "Email", required: true },
			{ key: "pass", label: "Password", required: true, secret: true },
		],
	},
	local: {
		label: "Local folder",
		pathLabel: "Absolute path",
		pathPlaceholder: "/backups/vidra",
		fields: [],
	},
};

const intervals = [
	{ hours: 0, label: "Only when run manually" },
	{ hours: 6, label: "Every 6 hours" },
	{ hours: 12, label: "Every 12 hours" },
	{ hours: 24, label: "Daily" },
	{ hours: 168, label: "Weekly" },
];

const intervalLabel = (hours: number) =>
	intervals.find((i) => i.hours === hours)?.label ?? `Every ${hours} hours`;

const statusStyles: Record<string, { label: string; className: string; icon: typeof CircleCheck }> = {
	ok: {
		label: "Up to date",
		className: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400",
		icon: CircleCheck,
	},
	error: { label: "Failed", className: "bg-destructive/15 text-destructive", icon: CircleAlert },
	running: { label: "Running", className: "bg-blue-500/15 text-blue-600 dark:text-blue-400", icon: Loader2 },
	never: { label: "Not run yet", className: "bg-muted text-muted-foreground", icon: Cloud },
};

export function BackupsSection() {
	const overview = useListBackupTargets();
	const [editing, setEditing] = useState<{ open: boolean; target?: ServicesBackupTargetDTO; key: number }>({
		open: false,
		key: 0,
	});
	const open = (target?: ServicesBackupTargetDTO) =>
		setEditing((e) => ({ open: true, target, key: e.key + 1 }));

	return (
		<Section title="Backups" description="Copy videos and database snapshots to cloud storage with rclone.">
			{!overview.data ? (
				<div className="space-y-2">
					<Skeleton className="h-24 w-full rounded-3xl" />
					<Skeleton className="h-11 w-full rounded-xl" />
				</div>
			) : !overview.data.available ? (
				<p className="rounded-2xl bg-muted/60 p-4 text-muted-foreground text-sm">
					rclone isn't installed on the server. The Docker image includes it.
				</p>
			) : (
				<>
					{overview.data.targets.length > 0 && (
						<ul className="space-y-3">
							{overview.data.targets.map((t) => (
								<TargetRow key={t.id} target={t} onEdit={() => open(t)} />
							))}
						</ul>
					)}
					<Button variant="secondary" className="h-11 w-full rounded-xl font-bold" onClick={() => open()}>
						<Plus /> Add backup target
					</Button>
					<p className="text-muted-foreground text-xs">{overview.data.version}</p>
				</>
			)}
			<TargetSheet
				key={editing.key}
				open={editing.open}
				target={editing.target}
				onOpenChange={(o) => setEditing((e) => ({ ...e, open: o }))}
			/>
		</Section>
	);
}

function TargetRow({ target: t, onEdit }: { target: ServicesBackupTargetDTO; onEdit: () => void }) {
	const qc = useQueryClient();
	const [confirm, setConfirm] = useState(false);
	const refresh = () => qc.invalidateQueries({ queryKey: getListBackupTargetsQueryKey() });
	const test = useTestBackupTarget({
		mutation: { onSuccess: () => toast.success(`${t.name} is reachable`) },
	});
	const run = useRunBackupTarget({ mutation: { onSuccess: () => toast.success("Backup started") } });
	const remove = useDeleteBackupTarget({ mutation: { onSuccess: refresh } });
	const status = statusStyles[t.lastStatus] ?? statusStyles.never;
	const Icon = t.provider === "local" ? HardDrive : Cloud;

	return (
		<li className="space-y-3 rounded-3xl border p-4">
			<div className="flex items-start gap-3">
				<div className="rounded-2xl bg-muted p-2.5">
					<Icon className="size-5" />
				</div>
				<div className="min-w-0 flex-1">
					<div className="flex flex-wrap items-center gap-2">
						<span className={cn("font-bold", !t.enabled && "text-muted-foreground")}>{t.name}</span>
						<Badge className={cn("rounded-full border-none", status.className)}>
							<status.icon className={cn(t.lastStatus === "running" && "animate-spin")} />
							{status.label}
						</Badge>
						{!t.enabled && (
							<Badge variant="secondary" className="rounded-full">
								Paused
							</Badge>
						)}
					</div>
					<p className="truncate text-muted-foreground text-sm">
						{providers[t.provider as Provider]?.label ?? t.provider}
						{t.path && ` · ${t.path}`}
						{t.lastRunAt && ` · ${formatRelative(t.lastRunAt)}`}
					</p>
					<p className="text-muted-foreground text-sm">
						Full backup: {intervalLabel(t.intervalHours).toLowerCase()}
						{t.lastFullAt && `, last ${formatRelative(t.lastFullAt)}`}
					</p>
				</div>
				<Button
					variant="ghost"
					size="icon"
					className="size-9 shrink-0 rounded-xl"
					aria-label="Edit"
					onClick={onEdit}
				>
					<Pencil />
				</Button>
			</div>
			{t.lastStatus === "error" && t.lastError && (
				<p className="break-words rounded-2xl bg-destructive/10 px-3 py-2 text-destructive text-sm">
					{t.lastError}
				</p>
			)}
			<div className="grid grid-cols-3 gap-2">
				<Button
					variant="secondary"
					className="h-10 rounded-xl font-semibold"
					disabled={test.isPending}
					onClick={() => test.mutate({ id: t.id })}
				>
					{test.isPending ? <Loader2 className="animate-spin" /> : <Plug />} Test
				</Button>
				<Button
					variant="secondary"
					className="h-10 rounded-xl font-semibold"
					disabled={run.isPending || t.lastStatus === "running"}
					onClick={() => run.mutate({ id: t.id })}
				>
					<Play /> Run
				</Button>
				<Button
					variant="secondary"
					className="h-10 rounded-xl font-semibold text-destructive"
					onClick={() => setConfirm(true)}
				>
					<Trash2 /> Remove
				</Button>
			</div>
			<ConfirmDialog
				open={confirm}
				onOpenChange={setConfirm}
				title="Remove this backup target?"
				description="Files already uploaded stay where they are."
				action="Remove"
				onConfirm={() => remove.mutate({ id: t.id })}
			/>
		</li>
	);
}

function validate(provider: Provider, values: ServicesBackupTargetInput, editing: boolean) {
	const spec = providers[provider];
	const config = z.object(
		Object.fromEntries(
			spec.fields.map((f) => [
				f.key,
				f.required && !(f.secret && editing)
					? z.string().trim().min(1, `${f.label} is required`)
					: z.string().optional(),
			]),
		),
	);
	return z
		.object({
			name: z.string().trim().min(1, "Name is required").max(64),
			path:
				provider === "local"
					? z
							.string()
							.trim()
							.regex(/^(\/|[A-Za-z]:[\\/])/, "Use an absolute path")
					: z.string().trim(),
			config,
		})
		.safeParse(values);
}

function TargetSheet({
	open,
	onOpenChange,
	target,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	target?: ServicesBackupTargetDTO;
}) {
	const qc = useQueryClient();
	const [provider, setProvider] = useState<Provider>((target?.provider as Provider) ?? "s3");
	const [values, setValues] = useState<ServicesBackupTargetInput>(
		target
			? {
					name: target.name,
					provider: target.provider,
					path: target.path,
					config: target.config,
					includeVideos: target.includeVideos,
					includeDatabase: target.includeDatabase,
					enabled: target.enabled,
					intervalHours: target.intervalHours,
				}
			: {
					name: "",
					provider: "s3",
					path: "",
					config: { provider: "Cloudflare" },
					includeVideos: true,
					includeDatabase: true,
					enabled: true,
					intervalHours: 24,
				},
	);
	const [errors, setErrors] = useState<Record<string, string>>({});
	const spec = providers[provider];
	const done = (msg: string) => {
		qc.invalidateQueries({ queryKey: getListBackupTargetsQueryKey() });
		toast.success(msg);
		onOpenChange(false);
	};
	const create = useCreateBackupTarget({ mutation: { onSuccess: () => done("Backup target added") } });
	const update = useUpdateBackupTarget({ mutation: { onSuccess: () => done("Backup target saved") } });
	const pending = create.isPending || update.isPending;

	const setConfig = (key: string, value: string) =>
		setValues((v) => ({ ...v, config: { ...v.config, [key]: value } }));

	const submit = () => {
		const result = validate(provider, values, Boolean(target));
		if (!result.success) {
			setErrors(Object.fromEntries(result.error.issues.map((i) => [i.path.join("."), i.message])));
			return;
		}
		setErrors({});
		const data = { ...values, provider };
		if (target) update.mutate({ id: target.id, data });
		else create.mutate({ data });
	};

	return (
		<Sheet
			open={open}
			onOpenChange={onOpenChange}
			title={target ? "Edit backup target" : "New backup target"}
			footer={
				<Button
					size="lg"
					className="h-12 w-full rounded-2xl font-bold text-base"
					disabled={pending}
					onClick={submit}
				>
					{pending && <Loader2 className="animate-spin" />}
					{target ? "Save" : "Add target"}
				</Button>
			}
		>
			<div className="space-y-5">
				{!target && (
					<div className="grid grid-cols-2 gap-2">
						{(Object.keys(providers) as Provider[]).map((p) => (
							<button
								key={p}
								type="button"
								onClick={() => {
									setProvider(p);
									setValues((v) => ({
										...v,
										config: p === "s3" ? { provider: "Cloudflare" } : ({} as Record<string, string>),
									}));
									setErrors({});
								}}
								className={cn(
									"rounded-2xl border-2 p-3 text-left font-bold text-sm transition-colors",
									provider === p
										? "border-primary bg-primary/5"
										: "border-transparent bg-muted/60 hover:bg-muted",
								)}
							>
								{providers[p].label}
							</button>
						))}
					</div>
				)}

				<TextInput
					label="Name"
					value={values.name}
					placeholder="e.g. R2 bucket"
					error={errors.name}
					onChange={(name) => setValues((v) => ({ ...v, name }))}
				/>

				{spec.fields.map((f) =>
					f.options ? (
						<div key={f.key} className="space-y-2">
							<Label className="font-semibold">{f.label}</Label>
							<Select value={values.config[f.key] ?? ""} onValueChange={(v) => setConfig(f.key, v)}>
								<SelectTrigger className="h-11! w-full rounded-xl">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{f.options.map((o) => (
										<SelectItem key={o} value={o}>
											{o === "Cloudflare" ? "Cloudflare R2" : o}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>
					) : (
						<TextInput
							key={f.key}
							label={f.label}
							value={values.config[f.key] ?? ""}
							placeholder={f.secret && target ? "Unchanged" : f.placeholder}
							help={f.help}
							secret={f.secret}
							multiline={f.multiline}
							error={errors[`config.${f.key}`]}
							onChange={(v) => setConfig(f.key, v)}
						/>
					),
				)}

				<TextInput
					label={spec.pathLabel}
					value={values.path}
					placeholder={spec.pathPlaceholder}
					error={errors.path}
					onChange={(path) => setValues((v) => ({ ...v, path }))}
				/>

				<div className="space-y-4 rounded-2xl bg-muted/50 p-4">
					<SwitchRow
						label="Videos"
						description="Upload each version when it finishes and remove deleted ones."
						checked={values.includeVideos}
						onCheckedChange={(includeVideos) => setValues((v) => ({ ...v, includeVideos }))}
					/>
					<SwitchRow
						label="Database"
						description="Snapshot shortly after changes and on every full backup, keeping the last 7."
						checked={values.includeDatabase}
						onCheckedChange={(includeDatabase) => setValues((v) => ({ ...v, includeDatabase }))}
					/>
					<div className="space-y-2">
						<Label className="font-semibold">Full backup</Label>
						<Select
							value={String(values.intervalHours)}
							onValueChange={(v) => setValues((s) => ({ ...s, intervalHours: Number(v) }))}
						>
							<SelectTrigger className="h-11! w-full rounded-xl bg-background">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{!intervals.some((i) => i.hours === values.intervalHours) && (
									<SelectItem value={String(values.intervalHours)}>
										{intervalLabel(values.intervalHours)}
									</SelectItem>
								)}
								{intervals.map((i) => (
									<SelectItem key={i.hours} value={String(i.hours)}>
										{i.label}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
						<p className="text-muted-foreground text-sm">
							Copies anything missing and takes a database snapshot. Run does the same on demand.
						</p>
					</div>
					<SwitchRow
						label="Enabled"
						checked={values.enabled}
						onCheckedChange={(enabled) => setValues((v) => ({ ...v, enabled }))}
					/>
				</div>
			</div>
		</Sheet>
	);
}

function TextInput({
	label,
	value,
	onChange,
	placeholder,
	help,
	error,
	secret,
	multiline,
}: {
	label: string;
	value: string;
	onChange: (v: string) => void;
	placeholder?: string;
	help?: string;
	error?: string;
	secret?: boolean;
	multiline?: boolean;
}) {
	const shared = {
		value,
		placeholder,
		"aria-invalid": Boolean(error),
		autoCapitalize: "none",
		autoCorrect: "off",
		spellCheck: false,
	} as const;
	return (
		<div className="space-y-2">
			<Label className="font-semibold">{label}</Label>
			{multiline ? (
				<textarea
					{...shared}
					rows={4}
					onChange={(e) => onChange(e.target.value)}
					className="w-full rounded-xl border border-input bg-transparent px-3 py-2 font-mono shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 aria-invalid:border-destructive dark:bg-input/30"
				/>
			) : (
				<Input
					{...shared}
					type={secret ? "password" : "text"}
					autoComplete={secret ? "new-password" : "off"}
					onChange={(e) => onChange(e.target.value)}
				/>
			)}
			{error ? (
				<p className="font-medium text-destructive text-sm">{error}</p>
			) : (
				help && <p className="text-muted-foreground text-sm">{help}</p>
			)}
		</div>
	);
}
