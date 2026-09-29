import { Logo } from "@/components/logo";
import { Skeleton } from "@/components/ui/skeleton";

export function HeadingSkeleton() {
	return (
		<div className="space-y-3">
			<Skeleton className="h-10 w-48 rounded-xl" />
			<Skeleton className="h-5 w-64 rounded-lg" />
		</div>
	);
}

export function VideoCardSkeleton() {
	return (
		<div className="overflow-hidden rounded-[2.5rem] border bg-card">
			<Skeleton className="aspect-video w-full rounded-none" />
			<div className="space-y-4 p-6">
				<Skeleton className="h-7 w-3/4 rounded-lg" />
				<Skeleton className="h-4 w-1/3 rounded-md" />
				<div className="flex gap-2 pt-2">
					<Skeleton className="h-12 flex-1 rounded-2xl" />
					<Skeleton className="size-12 rounded-2xl" />
					<Skeleton className="size-12 rounded-2xl" />
				</div>
			</div>
		</div>
	);
}

export function SectionSkeleton({ rows = 3 }: { rows?: number }) {
	return (
		<div className="space-y-5 rounded-[2.5rem] border bg-card p-6">
			<Skeleton className="h-6 w-40 rounded-lg" />
			{Array.from({ length: rows }, (_, i) => (
				// biome-ignore lint/suspicious/noArrayIndexKey: static placeholders
				<div key={i} className="flex items-center justify-between gap-4">
					<div className="flex-1 space-y-2">
						<Skeleton className="h-4 w-1/2 rounded-md" />
						<Skeleton className="h-3 w-3/4 rounded-md" />
					</div>
					<Skeleton className="h-8 w-14 rounded-full" />
				</div>
			))}
		</div>
	);
}

export function LibrarySkeleton() {
	return (
		<div className="space-y-8">
			<HeadingSkeleton />
			<Skeleton className="h-12 w-full rounded-2xl" />
			<div className="space-y-6">
				<VideoCardSkeleton />
				<VideoCardSkeleton />
			</div>
		</div>
	);
}

export function PageSkeleton() {
	return (
		<div className="space-y-8">
			<HeadingSkeleton />
			<SectionSkeleton />
			<SectionSkeleton rows={2} />
		</div>
	);
}

export function ShellSkeleton() {
	return (
		<div className="flex min-h-dvh flex-col">
			<header className="sticky top-0 z-40 border-b bg-background/80 pt-[env(safe-area-inset-top)]">
				<div className="mx-auto flex h-16 max-w-2xl items-center gap-2.5 px-4">
					<Logo className="size-8" />
					<span className="font-bold text-xl tracking-tight">Vidra</span>
					<Skeleton className="ml-auto size-10 rounded-xl" />
				</div>
			</header>
			<main className="mx-auto w-full max-w-2xl flex-1 px-4 pt-8">
				<LibrarySkeleton />
			</main>
		</div>
	);
}

export function AuthSkeleton() {
	return (
		<div className="flex min-h-dvh items-center justify-center px-4">
			<div className="w-full max-w-sm space-y-6 rounded-[2.5rem] border bg-card p-8">
				<div className="flex flex-col items-center gap-3">
					<Logo className="mb-2 size-14" />
					<Skeleton className="h-8 w-44 rounded-lg" />
					<Skeleton className="h-4 w-56 rounded-md" />
				</div>
				<Skeleton className="h-11 w-full rounded-xl" />
				<Skeleton className="h-11 w-full rounded-xl" />
				<Skeleton className="h-12 w-full rounded-2xl" />
			</div>
		</div>
	);
}
