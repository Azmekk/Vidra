<script lang="ts">
  import type {
    HandlersVideoResponse,
    ServicesDownloadProgressDTO,
  } from "$api/index";
  import * as Button from "$lib/components/ui/button/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Badge } from "$lib/components/ui/badge/index.js";
  import { Progress } from "$lib/components/ui/progress/index.js";
  import {
    Trash2,
    Play,
    ExternalLink,
    Clock,
    Download,
    HardDrive,
    Copy,
    Check,
    Pencil,
    X,
    Save,
  } from "@lucide/svelte";

  let {
    video,
    progress,
    isPlaying,
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
    video: HandlersVideoResponse;
    progress: ServicesDownloadProgressDTO | undefined;
    isPlaying: boolean;
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

  const currentStatus = $derived(
    progress?.status || video.downloadStatus || "unknown",
  );
  const isProcessing = $derived(
    ["downloading", "encoding", "pending"].includes(currentStatus.toLowerCase()),
  );
  const isCompleted = $derived(currentStatus.toLowerCase() === "completed");
  const isEditing = $derived(editingId === video.id);

  function getStatusColor(status: string) {
    switch (status.toLowerCase()) {
      case "completed":
        return "bg-green-500/10 text-green-500 border-green-500/20";
      case "downloading":
        return "bg-blue-500/10 text-blue-500 border-blue-500/20";
      case "encoding":
        return "bg-yellow-500/10 text-yellow-500 border-yellow-500/20";
      case "error":
        return "bg-red-500/10 text-red-500 border-red-500/20";
      default:
        return "bg-slate-500/10 text-slate-500 border-slate-500/20";
    }
  }
</script>

<div
  class="group relative flex flex-col overflow-hidden rounded-[2rem] border bg-card transition-all hover:shadow-2xl hover:shadow-primary/5"
>
  {#if isPlaying && isCompleted}
    <div class="aspect-video w-full overflow-hidden bg-black">
      <!-- svelte-ignore a11y_media_has_caption -->
      <video
        src={`/downloads/${video.fileName}`}
        controls
        autoplay
        playsinline
        preload="metadata"
        class="h-full w-full object-contain"
      >
      </video>
    </div>
  {:else}
    <button
      class="relative aspect-video w-full overflow-hidden bg-black transition-all hover:cursor-pointer"
      onclick={() => isCompleted && video.id && onTogglePlay(video.id)}
      disabled={!isCompleted}
    >
      {#if video.thumbnailFileName}
        <img
          src={`/downloads/${video.thumbnailFileName}`}
          alt={video.name}
          class="h-full w-full object-contain transition-transform duration-700 group-hover:scale-110"
        />
      {:else}
        <div
          class="flex h-full items-center justify-center text-muted-foreground"
        >
          <HardDrive class="h-12 w-12 opacity-20" />
        </div>
      {/if}

      <div class="absolute top-4 right-4">
        <Badge
          variant="outline"
          class={`px-3 py-1 rounded-full border-none font-bold backdrop-blur-md ${getStatusColor(currentStatus)}`}
        >
          {currentStatus.toUpperCase()}
        </Badge>
      </div>

      {#if isCompleted}
        <div
          class="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 [@media(hover:none)]:opacity-100 [@media(hover:none)]:bg-black/40 transition-all duration-300 group-hover:bg-black/40 group-hover:opacity-100"
        >
          <div
            class="rounded-full bg-white p-5 text-black shadow-2xl transition-transform duration-300 hover:scale-110 active:scale-90"
          >
            <Play class="h-8 w-8 fill-current" />
          </div>
        </div>
      {/if}
    </button>
  {/if}

  <div class="flex flex-1 flex-col p-6 sm:p-8 min-w-0">
    <div class="flex items-start justify-between gap-6">
      <div class="space-y-1 min-w-0 flex-1">
        {#if isEditing}
          <div class="flex items-center gap-2">
            <Input
              bind:value={editingName}
              class="h-10 rounded-lg bg-muted/50 border-primary/20 focus-visible:ring-primary/20"
              autofocus
              onkeydown={(e) => {
                if (e.key === "Enter") onRename(video.id!);
                if (e.key === "Escape") onCancelEditing();
              }}
            />
            <Button.Root
              variant="ghost"
              size="icon"
              onclick={() => onRename(video.id!)}
              class="h-10 w-10 shrink-0 text-green-600 hover:bg-green-500/10 hover:cursor-pointer"
            >
              <Save class="h-5 w-5" />
            </Button.Root>
            <Button.Root
              variant="ghost"
              size="icon"
              onclick={onCancelEditing}
              class="h-10 w-10 shrink-0 text-muted-foreground hover:bg-muted hover:cursor-pointer"
            >
              <X class="h-5 w-5" />
            </Button.Root>
          </div>
        {:else}
          <h3
            class="text-xl font-bold leading-tight tracking-tight sm:text-2xl line-clamp-2"
            title={video.name}
          >
            {video.name || "Untitled Video"}
          </h3>
        {/if}

        {#if video.id}
          <button
            onclick={() => onCopyId(video.id!, video.id!)}
            class="flex items-center gap-1.5 text-xs font-mono text-muted-foreground hover:text-primary transition-colors group/guid"
          >
            <span class="truncate max-w-[150px]">{video.id}</span>
            {#if copiedId === video.id}
              <Check class="h-3 w-3 text-green-500" />
            {:else}
              <Copy class="h-3 w-3 opacity-0 group-hover/guid:opacity-100" />
            {/if}
          </button>
        {/if}
      </div>
      <div class="grid grid-cols-2 gap-2 shrink-0">
        {#if !isEditing}
          <Button.Root
            variant="secondary"
            size="icon"
            onclick={() => onStartEditing(video)}
            class="h-11 w-11 rounded-2xl hover:cursor-pointer"
          >
            <Pencil class="h-5 w-5" />
          </Button.Root>
        {/if}
        {#if isCompleted}
          <Button.Root
            variant="secondary"
            size="icon"
            onclick={() => window.open(`/downloads/${video.fileName}`, "_blank")}
            class="h-11 w-11 rounded-2xl hover:cursor-pointer"
          >
            <ExternalLink class="h-5 w-5" />
          </Button.Root>
          <Button.Root
            variant="secondary"
            size="icon"
            href={`/downloads/${video.fileName}`}
            download={video.fileName}
            disabled={video.downloadStatus !== "completed"}
            class="h-11 w-11 rounded-2xl hover:cursor-pointer"
          >
            <Download class="h-5 w-5" />
          </Button.Root>
        {/if}
        <Button.Root
          variant="secondary"
          size="icon"
          onclick={() => video.id && onDelete(video.id)}
          class="h-11 w-11 rounded-2xl text-destructive hover:bg-destructive/10 hover:text-destructive hover:cursor-pointer"
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
              {#if currentStatus.toLowerCase() === "encoding"}
                <div class="relative flex h-3 w-3">
                  <span
                    class="absolute inline-flex h-full w-full animate-ping rounded-full bg-yellow-400 opacity-75"
                  ></span>
                  <span
                    class="relative inline-flex h-3 w-3 rounded-full bg-yellow-500"
                  ></span>
                </div>
                <span
                  class="text-yellow-600 dark:text-yellow-500 uppercase tracking-wider text-xs"
                  >Encoding...</span
                >
              {:else}
                <div class="relative flex h-3 w-3">
                  <span
                    class="absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-400 opacity-75"
                  ></span>
                  <span
                    class="relative inline-flex h-3 w-3 rounded-full bg-blue-500"
                  ></span>
                </div>
                <span
                  class="text-blue-600 dark:text-blue-500 uppercase tracking-wider text-xs"
                  >{progress?.speed || "Downloading..."}</span
                >
              {/if}
            </span>
            <span class="text-lg tabular-nums">
              {(currentStatus.toLowerCase() === "encoding"
                ? progress?.encodingPercent
                : progress?.percent
              )?.toFixed(0) || 0}%
            </span>
          </div>
          <Progress
            value={(currentStatus.toLowerCase() === "encoding"
              ? progress?.encodingPercent
              : progress?.percent) || 0}
            class="h-3 rounded-full"
          />
          {#if progress?.eta}
            <p class="text-right text-xs font-medium text-muted-foreground">
              ETA: {progress.eta}
            </p>
          {/if}
        </div>
      {:else}
        <div
          class="flex items-center gap-2 text-sm font-medium text-muted-foreground"
        >
          <Clock class="h-4 w-4" />
          <span
            >Added {new Date(video.createdAt || "").toLocaleDateString(
              undefined,
              { month: "short", day: "numeric", year: "numeric" },
            )}</span
          >
        </div>
      {/if}
    </div>
  </div>
</div>
