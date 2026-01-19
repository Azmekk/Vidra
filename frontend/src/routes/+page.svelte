<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { videosApi } from '$lib/api-client';
	import type { HandlersVideoResponse, ServicesDownloadProgressDTO } from '$api/index';
	import * as Card from "$lib/components/ui/card/index.js";
	import * as Button from "$lib/components/ui/button/index.js";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Progress } from "$lib/components/ui/progress/index.js";
	import { Plus, Trash2, Play, ExternalLink, Clock, Download, HardDrive } from "@lucide/svelte";

	let { data } = $props();
	let videos: HandlersVideoResponse[] = $state(data.videos || []);
	let progressMap: Record<string, ServicesDownloadProgressDTO> = $state({});
	let interval: any;
	let listInterval: any;

	async function fetchVideos() {
		try {
			const res = await videosApi.listVideos();
			videos = res.data;
		} catch (e) {
			console.error('Error fetching videos', e);
		}
	}

	async function updateProgress() {
		try {
			const res = await videosApi.listAllProgress();
			progressMap = res.data;
		} catch (e) {
			console.error('Error fetching progress', e);
		}
	}

	async function deleteVideo(id: string) {
		if (!confirm('Are you sure you want to delete this video?')) return;
		try {
			await videosApi.deleteVideo(id);
			await fetchVideos();
		} catch (e) {
			console.error('Error deleting video', e);
		}
	}

	onMount(() => {
		updateProgress();
		interval = setInterval(updateProgress, 1000);
		listInterval = setInterval(fetchVideos, 5000);

		return () => {
			clearInterval(interval);
			clearInterval(listInterval);
		}
	});

	function getProgress(id: string) {
		return progressMap[id];
	}

	let activeVideoId: string | null = $state(null);

	function getStatusColor(status: string) {
		switch (status.toLowerCase()) {
			case 'completed': return 'bg-green-500/10 text-green-500 border-green-500/20';
			case 'downloading': return 'bg-blue-500/10 text-blue-500 border-blue-500/20';
			case 'encoding': return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20';
			case 'error': return 'bg-red-500/10 text-red-500 border-red-500/20';
			default: return 'bg-slate-500/10 text-slate-500 border-slate-500/20';
		}
	}

	function togglePlay(id: string) {
		if (activeVideoId === id) {
			activeVideoId = null;
		} else {
			activeVideoId = id;
		}
	}
</script>

<div class="space-y-10">
	<div class="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-4xl font-extrabold tracking-tight">Library</h1>
			<p class="mt-1 text-lg text-muted-foreground font-medium">Your downloaded collection.</p>
		</div>
		<Button.Root href="/download" size="lg" class="rounded-2xl px-6 py-6 text-base font-bold shadow-lg shadow-primary/20 transition-all hover:scale-105 active:scale-95">
			<Plus class="mr-2 h-5 w-5 stroke-[3px]" />
			Download New
		</Button.Root>
	</div>

	<div class="grid grid-cols-1 gap-8">
		{#each videos as video (video.id)}
			{@const prog = video.id ? getProgress(video.id) : undefined}
			{@const currentStatus = prog?.status || video.downloadStatus || 'unknown'}
			{@const isProcessing = ['downloading', 'encoding', 'pending'].includes(currentStatus.toLowerCase())}
			{@const isCompleted = currentStatus.toLowerCase() === 'completed'}
			{@const isActive = activeVideoId === video.id}
			
			<div class="group relative flex flex-col overflow-hidden rounded-[2rem] border bg-card transition-all hover:shadow-2xl hover:shadow-primary/5">
				{#if isActive && isCompleted}
					<div class="aspect-video w-full bg-black">
						<video 
							src={`/downloads/${video.fileName}`} 
							controls 
							autoplay 
							class="h-full w-full"
						>
							<track kind="captions" />
						</video>
					</div>
				{:else}
					<button 
						class="relative aspect-video w-full overflow-hidden bg-muted transition-all"
						onclick={() => isCompleted && video.id && togglePlay(video.id)}
						disabled={!isCompleted}
					>
						{#if video.thumbnailFileName}
							<img
								src={`/downloads/${video.thumbnailFileName}`}
								alt={video.name}
								class="h-full w-full object-cover transition-transform duration-700 group-hover:scale-110"
							/>
						{:else}
							<div class="flex h-full items-center justify-center text-muted-foreground">
								<HardDrive class="h-12 w-12 opacity-20" />
							</div>
						{/if}
						
						<div class="absolute top-4 right-4">
							<Badge variant="outline" class={`px-3 py-1 rounded-full border-none font-bold backdrop-blur-md ${getStatusColor(currentStatus)}`}>
								{currentStatus.toUpperCase()}
							</Badge>
						</div>

						{#if isCompleted}
							<div class="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 transition-all duration-300 group-hover:bg-black/40 group-hover:opacity-100">
								<div class="rounded-full bg-white p-5 text-black shadow-2xl transition-transform duration-300 hover:scale-110 active:scale-90">
									<Play class="h-8 w-8 fill-current" />
								</div>
							</div>
						{/if}
					</button>
				{/if}
				
				<div class="flex flex-1 flex-col p-6 sm:p-8 min-w-0">
					<div class="flex items-start justify-between gap-4">
						<h3 class="text-xl font-bold leading-tight tracking-tight sm:text-2xl line-clamp-2" title={video.name}>
							{video.name || 'Untitled Video'}
						</h3>
						<div class="flex shrink-0 gap-2">
							{#if isCompleted}
								<Button.Root variant="secondary" size="icon" href={`/downloads/${video.fileName}`} download class="h-11 w-11 rounded-2xl">
									<Download class="h-5 w-5" />
								</Button.Root>
							{/if}
							<Button.Root 
								variant="secondary" 
								size="icon" 
								onclick={() => video.id && deleteVideo(video.id)} 
								class="h-11 w-11 rounded-2xl text-destructive hover:bg-destructive/10 hover:text-destructive"
							>
								<Trash2 class="h-5 w-5" />
							</Button.Root>
						</div>
					</div>
					
					<div class="mt-6">
						{#if isProcessing}
							<div class="space-y-3">
								<div class="flex items-center justify-between text-sm font-bold">
									<span class="flex items-center gap-2">
										{#if currentStatus.toLowerCase() === 'encoding'}
											<div class="relative flex h-3 w-3">
												<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-yellow-400 opacity-75"></span>
												<span class="relative inline-flex h-3 w-3 rounded-full bg-yellow-500"></span>
											</div>
											<span class="text-yellow-600 dark:text-yellow-500 uppercase tracking-wider text-xs">Encoding...</span>
										{:else}
											<div class="relative flex h-3 w-3">
												<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-400 opacity-75"></span>
												<span class="relative inline-flex h-3 w-3 rounded-full bg-blue-500"></span>
											</div>
											<span class="text-blue-600 dark:text-blue-500 uppercase tracking-wider text-xs">{prog?.speed || 'Downloading...'}</span>
										{/if}
									</span>
									<span class="text-lg tabular-nums">{prog?.percent?.toFixed(0) || 0}%</span>
								</div>
								<Progress value={prog?.percent || 0} class="h-3 rounded-full" />
								{#if prog?.eta}
									<p class="text-right text-xs font-medium text-muted-foreground">ETA: {prog.eta}</p>
								{/if}
							</div>
						{:else}
							<div class="flex items-center gap-2 text-sm font-medium text-muted-foreground">
								<Clock class="h-4 w-4" />
								<span>Added {new Date(video.createdAt || '').toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}</span>
							</div>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>

	{#if videos.length === 0}
		<div class="flex min-h-[400px] flex-col items-center justify-center rounded-[3rem] border-2 border-dashed p-12 text-center animate-in fade-in zoom-in duration-500">
			<div class="rounded-3xl bg-muted p-6">
				<Download class="h-12 w-12 text-muted-foreground opacity-50" />
			</div>
			<h2 class="mt-6 text-2xl font-bold tracking-tight">Your library is empty</h2>
			<p class="mt-2 text-muted-foreground max-w-[250px] mx-auto">Start by downloading your first video to see it here.</p>
			<Button.Root href="/download" size="lg" class="mt-8 rounded-2xl">
				<Plus class="mr-2 h-5 w-5" />
				Download New Video
			</Button.Root>
		</div>
	{/if}
</div>