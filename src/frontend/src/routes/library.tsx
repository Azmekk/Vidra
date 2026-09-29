import { useInfiniteQuery } from "@tanstack/react-query";
import { getRouteApi, Link } from "@tanstack/react-router";
import { Plus, Search, SearchX, Video } from "lucide-react";
import { useEffect, useEffectEvent, useRef, useState } from "react";
import { LibrarySkeleton, VideoCardSkeleton } from "@/components/skeletons";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { VideoCard } from "@/components/video-card";
import { libraryQuery } from "@/lib/videos";

const route = getRouteApi("/app/");

const orders = [
	{ value: "created_at_desc", label: "Newest" },
	{ value: "created_at_asc", label: "Oldest" },
	{ value: "name_asc", label: "Name A–Z" },
	{ value: "name_desc", label: "Name Z–A" },
] as const;

export function LibraryPage() {
	const { q = "", order = "created_at_desc" } = route.useSearch();
	const navigate = route.useNavigate();
	const query = useInfiniteQuery(libraryQuery({ search: q || undefined, order }));
	const sentinel = useRef<HTMLDivElement>(null);
	const { hasNextPage, isFetchingNextPage, fetchNextPage } = query;

	useEffect(() => {
		const el = sentinel.current;
		if (!el || !hasNextPage) return;
		const io = new IntersectionObserver(
			([entry]) => entry.isIntersecting && !isFetchingNextPage && fetchNextPage(),
			{ rootMargin: "800px" },
		);
		io.observe(el);
		return () => io.disconnect();
	}, [hasNextPage, isFetchingNextPage, fetchNextPage]);

	if (!query.data) return <LibrarySkeleton />;

	const videos = query.data.pages.flatMap((p) => p.videos);
	const total = query.data.pages[0]?.totalCount ?? 0;

	return (
		<div className="space-y-8">
			<div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
				<div>
					<h1 className="font-extrabold text-4xl tracking-tight">Library</h1>
					<p className="mt-1 font-medium text-lg text-muted-foreground">
						{total === 1 ? "1 video" : `${total} videos`}
					</p>
				</div>
				<Button
					asChild
					size="lg"
					className="h-12 rounded-2xl px-6 font-bold text-base shadow-lg shadow-primary/20 transition-transform hover:scale-105 active:scale-95"
				>
					<Link to="/download">
						<Plus className="size-5 stroke-3" /> Download new
					</Link>
				</Button>
			</div>

			<div className="flex gap-2">
				<SearchInput
					value={q}
					onChange={(value) => navigate({ search: (s) => ({ ...s, q: value || undefined }), replace: true })}
				/>
				<Select
					value={order}
					onValueChange={(v) =>
						navigate({ search: (s) => ({ ...s, order: v as typeof order }), replace: true })
					}
				>
					<SelectTrigger className="h-12! w-32 shrink-0 rounded-2xl font-semibold">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{orders.map((o) => (
							<SelectItem key={o.value} value={o.value}>
								{o.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>

			{videos.length === 0 ? (
				<Empty searching={Boolean(q)} />
			) : (
				<div className="space-y-6">
					{videos.map((v) => (
						<VideoCard key={v.id} video={v} />
					))}
					{hasNextPage && (
						<div ref={sentinel}>
							<VideoCardSkeleton />
						</div>
					)}
				</div>
			)}
		</div>
	);
}

function SearchInput({ value, onChange }: { value: string; onChange: (v: string) => void }) {
	const [text, setText] = useState(value);
	const commit = useEffectEvent((v: string) => onChange(v));

	useEffect(() => {
		if (text === value) return;
		const t = setTimeout(() => commit(text.trim()), 250);
		return () => clearTimeout(t);
	}, [text, value]);

	return (
		<div className="relative flex-1">
			<Search className="pointer-events-none absolute top-1/2 left-4 size-5 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search videos"
				value={text}
				onChange={(e) => setText(e.target.value)}
				className="h-12 rounded-2xl pl-11"
			/>
		</div>
	);
}

function Empty({ searching }: { searching: boolean }) {
	const Icon = searching ? SearchX : Video;
	return (
		<div className="flex flex-col items-center rounded-[2.5rem] border border-dashed px-6 py-16 text-center">
			<div className="mb-5 rounded-3xl bg-muted p-5">
				<Icon className="size-8 text-muted-foreground" />
			</div>
			<h2 className="font-bold text-xl">{searching ? "No matches" : "Nothing here yet"}</h2>
			<p className="mt-1 max-w-xs text-muted-foreground">
				{searching
					? "Try a different search."
					: "Paste a link on the Download page to grab your first video."}
			</p>
		</div>
	);
}
