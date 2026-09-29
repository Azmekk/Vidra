import { keepPreviousData, useInfiniteQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ChevronDown, CircleCheck, Search, Video } from "lucide-react";
import { useDeferredValue, useEffect, useRef, useState } from "react";
import { listRecentErrors } from "@/api/gen/errors/errors";
import type { HandlersErrorResponse } from "@/api/gen/model";
import { HeadingSkeleton } from "@/components/skeletons";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { formatRelative } from "@/lib/format";

export function ErrorsPage() {
	const [text, setText] = useState("");
	const search = useDeferredValue(text.trim());
	const query = useInfiniteQuery({
		queryKey: ["/api/errors", "infinite", search],
		queryFn: ({ pageParam, signal }) =>
			listRecentErrors({ search: search || undefined, page: pageParam, limit: 20 }, { signal }),
		initialPageParam: 1,
		getNextPageParam: (last) => (last.currentPage < last.totalPages ? last.currentPage + 1 : undefined),
		placeholderData: keepPreviousData,
	});
	const sentinel = useRef<HTMLDivElement>(null);
	const { hasNextPage, isFetchingNextPage, fetchNextPage } = query;

	useEffect(() => {
		const el = sentinel.current;
		if (!el || !hasNextPage) return;
		const io = new IntersectionObserver(
			([entry]) => entry.isIntersecting && !isFetchingNextPage && fetchNextPage(),
			{ rootMargin: "600px" },
		);
		io.observe(el);
		return () => io.disconnect();
	}, [hasNextPage, isFetchingNextPage, fetchNextPage]);

	if (!query.data) {
		return (
			<div className="space-y-8">
				<HeadingSkeleton />
				<Skeleton className="h-12 w-full rounded-2xl" />
				<ErrorSkeleton />
				<ErrorSkeleton />
				<ErrorSkeleton />
			</div>
		);
	}

	const errors = query.data.pages.flatMap((p) => p.errors);
	const total = query.data.pages[0]?.totalCount ?? 0;

	return (
		<div className="space-y-8">
			<div>
				<h1 className="font-extrabold text-4xl tracking-tight">Errors</h1>
				<p className="mt-1 font-medium text-lg text-muted-foreground">
					{total === 1 ? "1 logged error" : `${total} logged errors`}
				</p>
			</div>

			<div className="relative">
				<Search className="pointer-events-none absolute top-1/2 left-4 size-5 -translate-y-1/2 text-muted-foreground" />
				<Input
					type="search"
					placeholder="Search message, command or video ID"
					value={text}
					onChange={(e) => setText(e.target.value)}
					className="h-12 rounded-2xl pl-11"
				/>
			</div>

			{errors.length === 0 ? (
				<div className="flex flex-col items-center rounded-[2.5rem] border border-dashed px-6 py-16 text-center">
					<div className="mb-5 rounded-3xl bg-muted p-5">
						<CircleCheck className="size-8 text-muted-foreground" />
					</div>
					<h2 className="font-bold text-xl">{search ? "No matches" : "All clear"}</h2>
					<p className="mt-1 text-muted-foreground">
						{search ? "Try a different search." : "Nothing has gone wrong."}
					</p>
				</div>
			) : (
				<div className="space-y-4">
					{errors.map((e) => (
						<ErrorCard key={e.id} error={e} />
					))}
					{hasNextPage && (
						<div ref={sentinel}>
							<ErrorSkeleton />
						</div>
					)}
				</div>
			)}
		</div>
	);
}

function ErrorCard({ error: e }: { error: HandlersErrorResponse }) {
	return (
		<article className="space-y-3 rounded-[2rem] border bg-card p-5">
			<div className="flex items-start justify-between gap-3">
				<p className="font-bold text-destructive leading-snug">{e.errorMessage}</p>
				<time className="shrink-0 text-muted-foreground text-xs" dateTime={e.createdAt}>
					{formatRelative(e.createdAt)}
				</time>
			</div>
			<p className="font-mono text-muted-foreground text-xs">{e.command}</p>
			{e.videoId && (
				<Link
					to="/"
					search={{ q: e.videoId }}
					className="inline-flex items-center gap-1.5 rounded-full bg-muted px-3 py-1.5 font-semibold text-xs transition-colors hover:bg-muted/70"
				>
					<Video className="size-3.5" /> View video
				</Link>
			)}
			{e.output && (
				<details className="group">
					<summary className="flex cursor-pointer list-none items-center gap-1 font-semibold text-sm">
						<ChevronDown className="size-4 transition-transform group-open:rotate-180" /> Output
					</summary>
					<pre className="mt-2 max-h-72 overflow-auto whitespace-pre-wrap break-all rounded-2xl bg-muted p-4 font-mono text-xs">
						{e.output}
					</pre>
				</details>
			)}
		</article>
	);
}

function ErrorSkeleton() {
	return (
		<div className="space-y-3 rounded-[2rem] border bg-card p-5">
			<div className="flex justify-between gap-3">
				<Skeleton className="h-5 w-2/3 rounded-md" />
				<Skeleton className="h-4 w-16 rounded-md" />
			</div>
			<Skeleton className="h-3.5 w-1/2 rounded-md" />
			<Skeleton className="h-4 w-20 rounded-md" />
		</div>
	);
}
