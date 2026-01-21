<script lang="ts">
  import type {
    HandlersVideoResponse,
    ServicesDownloadProgressDTO,
  } from "$api/index";
  import * as Button from "$lib/components/ui/button/index.js";
  import { Plus, Download } from "@lucide/svelte";
  import VideoCard from "./video-card.svelte";

  let {
    videos,
    progressMap,
    activeVideoId,
    copiedId,
    editingId,
    editingName = $bindable(""),
    onTogglePlay,
    onStartEditing,
    onCancelEditing,
    onRename,
    onDelete,
    onCopyId,
  }: {
    videos: HandlersVideoResponse[];
    progressMap: Record<string, ServicesDownloadProgressDTO>;
    activeVideoId: string | null;
    copiedId: string | null;
    editingId: string | null;
    editingName: string;
    onTogglePlay: (id: string) => void;
    onStartEditing: (video: HandlersVideoResponse) => void;
    onCancelEditing: () => void;
    onRename: (id: string) => void;
    onDelete: (id: string) => void;
    onCopyId: (text: string, id: string) => void;
  } = $props();
</script>

{#if videos.length > 0}
  <div class="grid grid-cols-1 gap-8">
    {#each videos as video (video.id)}
      <VideoCard
        {video}
        progress={video.id ? progressMap[video.id] : undefined}
        isPlaying={activeVideoId === video.id}
        {copiedId}
        {editingId}
        bind:editingName
        {onTogglePlay}
        {onStartEditing}
        {onCancelEditing}
        {onRename}
        {onDelete}
        {onCopyId}
      />
    {/each}
  </div>
{:else}
  <div
    class="flex min-h-[400px] flex-col items-center justify-center rounded-[3rem] border-2 border-dashed p-12 text-center animate-in fade-in zoom-in duration-500"
  >
    <div class="rounded-3xl bg-muted p-6">
      <Download class="h-12 w-12 text-muted-foreground opacity-50" />
    </div>
    <h2 class="mt-6 text-2xl font-bold tracking-tight">
      Your library is empty
    </h2>
    <p class="mt-2 text-muted-foreground max-w-[250px] mx-auto">
      Start by downloading your first video to see it here.
    </p>
    <Button.Root href="/download" size="lg" class="mt-8 rounded-2xl">
      <Plus class="mr-2 h-5 w-5" />
      Download New Video
    </Button.Root>
  </div>
{/if}
