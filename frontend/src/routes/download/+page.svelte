<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { videosApi, settingsApi } from "$lib/api-client";
  import * as Button from "$lib/components/ui/button/index.js";
  import * as Select from "$lib/components/ui/select/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import { Slider } from "$lib/components/ui/slider/index.js";
  import type {
    HandlersMetadataRequest,
    ServicesVideoMetadata,
    ServicesVideoOption,
  } from "$api/index";
  import {
    Search,
    Download,
    ArrowLeft,
    Info,
    Youtube,
    FilePlay,
    Clock,
    Type,
    X,
    Settings2,
  } from "@lucide/svelte";
  import { Switch } from "$lib/components/ui/switch/index.js";

  let url = $state("");
  let customName = $state("");
  let loading = $state(false);
  let metadata: ServicesVideoMetadata | null = $state(null);
  let selectedFormatId = $state<string>("");
  let reEncode = $state(true);
  let error = $state<string | null>(null);
  let searchQuery = $state("");

  // Encoding options
  let videoCodec = $state<string>("libx264");
  let audioCodec = $state<string>("aac");
  let crfValue = $state<number[]>([23]);

  onMount(async () => {
    try {
      const res = await settingsApi.getSettings();
      reEncode = res.data.defaultReEncode ?? true;
      videoCodec = res.data.defaultVideoCodec || "libx264";
      audioCodec = res.data.defaultAudioCodec || "aac";
      crfValue = [res.data.defaultCrf ?? 23];
    } catch (e) {
      console.error("Failed to load default settings", e);
    }
  });

  const videoCodecOptions = [
    {
      value: "libx264",
      label: "H.264",
      description: "Best compatibility",
      output: ".mp4",
    },
    {
      value: "libvpx-vp9",
      label: "VP9 Software",
      description: "Better compression, slower",
      output: ".webm",
    },
    {
      value: "vp9_qsv",
      label: "VP9 Hardware (QSV)",
      description: "Fast, requires Intel GPU",
      output: ".webm",
    },
  ];

  const audioCodecOptions = [
    { value: "aac", label: "AAC", description: "Universal compatibility" },
    {
      value: "libopus",
      label: "Opus",
      description: "Better quality at low bitrate",
    },
  ];

  const crfRanges: Record<
    string,
    { min: number; max: number; default: number }
  > = {
    libx264: { min: 18, max: 28, default: 23 },
    "libvpx-vp9": { min: 24, max: 40, default: 32 },
    vp9_qsv: { min: 18, max: 35, default: 25 },
  };

  const currentCrfRange = $derived(crfRanges[videoCodec] || crfRanges.libx264);
  const selectedVideoCodecInfo = $derived(
    videoCodecOptions.find((o) => o.value === videoCodec) ||
      videoCodecOptions[0],
  );
  const selectedAudioCodecInfo = $derived(
    audioCodecOptions.find((o) => o.value === audioCodec) ||
      audioCodecOptions[0],
  );

  // Reset CRF to default when codec changes
  $effect(() => {
    const range = crfRanges[videoCodec];
    if (range) {
      crfValue = [range.default];
    }
  });

  const isPlaylist = $derived(url.includes("list="));

  function getResolutionScore(res?: string) {
    if (!res) return 0;
    const parts = res.toLowerCase().split("x");
    if (parts.length === 2) {
      const w = parseInt(parts[0]);
      const h = parseInt(parts[1]);
      return Math.min(w, h) || Math.max(w, h) || 0;
    }
    const match = res.match(/(\d+)/);
    return match ? parseInt(match[0]) : 0;
  }

  const filteredOptions = $derived.by(() => {
    if (!metadata?.options) return [];
    return metadata.options
      .filter((option: ServicesVideoOption) => {
        const isAudioOnly =
          option.vcodec === "none" ||
          (!option.resolution && option.acodec !== "none");
        if (isAudioOnly) return false;

        if (!searchQuery) return true;
        const query = searchQuery.toLowerCase();
        return (
          option.resolution?.toLowerCase().includes(query) ||
          option.extension?.toLowerCase().includes(query) ||
          option.note?.toLowerCase().includes(query)
        );
      })
      .sort((a, b) => {
        const scoreA = getResolutionScore(a.resolution);
        const scoreB = getResolutionScore(b.resolution);
        if (scoreA !== scoreB) return scoreB - scoreA;
        return (b.file_size || 0) - (a.file_size || 0);
      });
  });

  const selectedFormatLabel = $derived.by(() => {
    if (!selectedFormatId) return "Best Quality (Auto)";
    if (!metadata?.options) return "Select a format";
    const option = metadata.options.find(
      (o: ServicesVideoOption) => o.format_id === selectedFormatId,
    );
    if (!option) return "Select a format";
    return `${option.resolution || "Unknown"} - ${option.extension} (${formatFileSize(option.file_size)})`;
  });

  async function checkMetadata() {
    if (!url) return;
    loading = true;
    error = null;
    metadata = null;
    searchQuery = "";
    selectedFormatId = ""; // Reset to auto
    try {
      const req: HandlersMetadataRequest = { url };
      const res = await videosApi.getMetadata(req);
      metadata = res.data;
      if (metadata?.title && !customName) {
        customName = metadata.title;
      } else if (metadata?.title) {
        customName = metadata.title;
      }
    } catch (e: any) {
      console.error(e);
      error = "Failed to fetch metadata. Please check the URL and try again.";
    } finally {
      loading = false;
    }
  }

  async function startDownload() {
    if (!url) return;
    loading = true;
    error = null;
    try {
      await videosApi.createVideo({
        downloadUrl: url,
        formatId: selectedFormatId || undefined,
        name: customName || metadata?.title || "New Download",
        reEncode: reEncode,
        encodingOptions: reEncode
          ? {
              videoCodec: videoCodec,
              audioCodec: audioCodec,
              crf: crfValue[0],
            }
          : undefined,
      });
      goto("/");
    } catch (e) {
      console.error(e);
      error = "Failed to start download. Please try again.";
      loading = false;
    }
  }

  function formatFileSize(bytes?: number) {
    if (!bytes) return "Unknown size";
    const units = ["B", "KB", "MB", "GB"];
    let size = bytes;
    let i = 0;
    while (size >= 1024 && i < units.length - 1) {
      size /= 1024;
      i++;
    }
    return `${size.toFixed(1)} ${units[i]}`;
  }
</script>

<div
  class="mx-auto space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700"
>
  <div class="flex flex-col gap-6">
    <div class="flex items-center gap-4">
      <Button.Root
        variant="secondary"
        size="icon"
        href="/"
        class="shrink-0 h-12 w-12 rounded-2xl"
      >
        <ArrowLeft class="h-5 w-5" />
      </Button.Root>
      <h1 class="text-4xl font-extrabold tracking-tight">New Download</h1>
    </div>
    <p class="text-lg text-muted-foreground font-medium leading-relaxed">
      Paste a video link from YouTube, Vimeo, or other supported platforms to
      begin.
    </p>
  </div>

  <div class="space-y-6">
    <div
      class="relative overflow-hidden rounded-[2.5rem] border bg-card p-2 shadow-2xl shadow-primary/5 transition-all focus-within:ring-2 focus-within:ring-primary/20"
    >
      <div class="flex flex-col gap-2 sm:flex-row">
        <div class="relative flex-1">
          <Search
            class="absolute left-6 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground"
          />
          <input
            type="text"
            bind:value={url}
            placeholder="Paste URL here..."
            class="h-16 w-full rounded-[2rem] border-none bg-transparent pl-14 pr-4 text-lg font-medium focus:outline-none focus:ring-0"
            disabled={loading}
            onkeydown={(e) =>
              e.key === "Enter" &&
              (metadata ? startDownload() : checkMetadata())}
          />
        </div>
        <Button.Root
          onclick={checkMetadata}
          disabled={loading || !url}
          variant="secondary"
          class="h-16 rounded-[2rem] px-8 text-lg font-bold transition-all hover:scale-[1.02] active:scale-[0.98]"
        >
          {loading && !metadata ? "Analysing..." : "Fetch Info"}
        </Button.Root>
      </div>
    </div>

    {#if isPlaylist}
      <div
        class="flex items-center gap-3 rounded-2xl bg-amber-500/10 p-4 text-amber-600 border border-amber-500/20 animate-in slide-in-from-top-2 duration-300"
      >
        <Info class="h-5 w-5 shrink-0" />
        <span class="text-sm font-bold"
          >Playlist detected. Only the first video will be downloaded.</span
        >
      </div>
    {/if}
  </div>

  {#if error}
    <div
      class="flex items-center gap-4 rounded-[2rem] bg-destructive/10 p-6 text-destructive border border-destructive/20 animate-in zoom-in-95 duration-300"
    >
      <div class="rounded-full bg-destructive/20 p-2">
        <Info class="h-6 w-6" />
      </div>

      <span class="font-bold text-lg">{error}</span>
    </div>
  {/if}

  {#if loading && !metadata}
    <div
      class="overflow-hidden rounded-[2.5rem] border bg-card p-8 animate-pulse"
    >
      <div class="flex flex-col gap-8 sm:flex-row">
        <div
          class="aspect-video w-full rounded-[1.5rem] bg-muted sm:w-72"
        ></div>

        <div class="flex-1 space-y-4">
          <div class="h-8 w-3/4 rounded-full bg-muted"></div>

          <div class="space-y-3">
            <div class="h-4 w-full rounded-full bg-muted"></div>

            <div class="h-4 w-full rounded-full bg-muted"></div>

            <div class="h-4 w-2/3 rounded-full bg-muted"></div>
          </div>
        </div>
      </div>
    </div>
  {/if}

  {#if metadata}
    <div
      class="group relative overflow-hidden rounded-[2.5rem] border bg-card shadow-2xl shadow-primary/5 transition-all animate-in zoom-in-95 slide-in-from-top-4 duration-500"
    >
      <div class="p-8 space-y-10">
        <div class="flex flex-col gap-8 sm:flex-row">
          {#if metadata.thumbnail}
            <div
              class="relative aspect-video w-full shrink-0 overflow-hidden rounded-[1.5rem] border shadow-2xl sm:w-72"
            >
              <img
                src={metadata.thumbnail}
                alt={metadata.title}
                class="h-full w-full object-cover transition-transform duration-700 group-hover:scale-110"
              />
            </div>
          {/if}

          <div class="space-y-4 min-w-0">
            <h3
              class="text-2xl font-black leading-tight tracking-tight sm:text-3xl line-clamp-2"
            >
              {metadata.title}
            </h3>

            <p
              class="text-base text-muted-foreground line-clamp-3 leading-relaxed font-medium"
            >
              {metadata.description}
            </p>

            <div class="flex flex-wrap gap-3">
              <div
                class="flex items-center text-sm font-bold bg-muted px-4 py-2 rounded-full"
              >
                <Clock class="mr-2 h-4 w-4" />

                {metadata.duration
                  ? (metadata.duration / 60).toFixed(1) + " min"
                  : "Unknown"}
              </div>

              {#if url.includes("youtube.com") || url.includes("youtu.be")}
                <div
                  class="flex items-center text-sm font-bold bg-red-500/10 text-red-500 px-4 py-2 rounded-full border border-red-500/10"
                >
                  <Youtube class="mr-2 h-4 w-4" />

                  YouTube
                </div>
              {/if}
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="flex items-center justify-between px-2">
            <Label for="quality" class="text-lg font-bold"
              >Override Quality</Label
            >

            <span
              class="text-xs font-bold text-muted-foreground uppercase tracking-widest opacity-60"
              >Optional</span
            >
          </div>

          <Select.Root type="single" bind:value={selectedFormatId}>
            <Select.Trigger
              id="quality"
              class="h-16 w-full rounded-[2rem] border-2 bg-muted/30 px-6 text-lg font-bold transition-all hover:bg-muted/50 focus:ring-primary/20"
            >
              {selectedFormatLabel}
            </Select.Trigger>

            <Select.Content
              class="rounded-[2.5rem] border-2 p-3 shadow-2xl min-w-[320px]"
            >
              <div class="mb-3 px-2">
                <div class="relative">
                  <Search
                    class="absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground opacity-50"
                  />

                  <input
                    type="text"
                    bind:value={searchQuery}
                    placeholder="Search quality (e.g. 1080p)..."
                    class="h-12 w-full rounded-2xl bg-muted/50 pl-11 pr-4 text-sm font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
                    onpointerdown={(e) => e.stopPropagation()}
                    onkeydown={(e) => e.stopPropagation()}
                  />
                </div>
              </div>

              <div class="max-h-[350px] overflow-y-auto pr-1">
                <Select.Item
                  value=""
                  class="rounded-2xl py-4 px-5 font-bold transition-all focus:bg-primary focus:text-primary-foreground mb-1"
                >
                  <div class="flex items-center justify-between w-full">
                    <div class="flex flex-col gap-0.5">
                      <span class="text-base">Best Quality</span>

                      <span
                        class="text-[10px] opacity-70 font-medium tracking-wide uppercase"
                        >Automatic Selection</span
                      >
                    </div>

                    <div class="text-right">
                      <span
                        class="text-xs font-black opacity-40 uppercase tracking-widest"
                        >Default</span
                      >
                    </div>
                  </div>
                </Select.Item>

                {#if filteredOptions.length > 0}
                  {#each filteredOptions as option}
                    <Select.Item
                      value={option.format_id || ""}
                      class="rounded-2xl py-4 px-5 font-bold transition-all focus:bg-primary focus:text-primary-foreground mb-1 last:mb-0"
                    >
                      <div class="flex items-center justify-between w-full">
                        <div class="flex flex-col gap-0.5">
                          <span class="text-base"
                            >{option.resolution} •

                            <span class="uppercase opacity-60 text-xs"
                              >{option.extension}</span
                            ></span
                          >

                          <span
                            class="text-[10px] opacity-70 font-medium tracking-wide uppercase"
                            >{option.note || "Standard"}</span
                          >
                        </div>

                        <div class="text-right">
                          <span class="text-sm font-black tabular-nums"
                            >{formatFileSize(option.file_size)}</span
                          >
                        </div>
                      </div>
                    </Select.Item>
                  {/each}
                {:else}
                  <div class="py-12 text-center text-muted-foreground">
                    <FilePlay class="mx-auto h-8 w-8 opacity-20 mb-3" />

                    <p class="text-sm font-bold">No formats found</p>
                  </div>
                {/if}
              </div>
            </Select.Content>
          </Select.Root>

          <div
            class="flex items-center justify-between px-2 pt-6 border-t border-muted"
          >
            <div class="space-y-0.5">
              <Label class="text-lg font-bold">Re-encode Video</Label>
              <p class="text-sm text-muted-foreground font-medium">
                Convert to a different format
              </p>
            </div>
            <Switch bind:checked={reEncode} aria-label="Toggle re-encoding" />
          </div>

          {#if reEncode}
            <div
              class="mt-6 space-y-6 animate-in slide-in-from-top-2 duration-300"
            >
              <div
                class="flex items-center gap-2 px-2 text-sm font-bold text-muted-foreground uppercase tracking-widest"
              >
                <Settings2 class="h-4 w-4" />
                Encoding Options
              </div>

              <div class="grid gap-6 sm:grid-cols-2">
                <!-- Video Codec -->
                <div class="space-y-3">
                  <Label class="text-base font-bold px-2">Video Codec</Label>
                  <Select.Root type="single" bind:value={videoCodec}>
                    <Select.Trigger
                      class="h-auto! w-full rounded-2xl border-2 bg-muted/30 px-5 py-4 text-base font-bold transition-all hover:bg-muted/50"
                    >
                      <div class="flex flex-col items-start gap-0.5">
                        <span>{selectedVideoCodecInfo.label}</span>
                        <span
                          class="text-[10px] font-medium opacity-60 uppercase tracking-wide"
                          >{selectedVideoCodecInfo.description} • {selectedVideoCodecInfo.output}</span
                        >
                      </div>
                    </Select.Trigger>
                    <Select.Content class="rounded-2xl border-2 p-2 shadow-2xl">
                      {#each videoCodecOptions as option}
                        <Select.Item
                          value={option.value}
                          class="rounded-xl py-3 px-4 font-bold transition-all focus:bg-primary focus:text-primary-foreground mb-1 last:mb-0"
                        >
                          <div class="flex flex-col gap-0.5">
                            <span>{option.label}</span>
                            <span
                              class="text-[10px] font-medium opacity-60 uppercase tracking-wide"
                              >{option.description} • {option.output}</span
                            >
                          </div>
                        </Select.Item>
                      {/each}
                    </Select.Content>
                  </Select.Root>
                </div>

                <!-- Audio Codec -->
                <div class="space-y-3">
                  <Label class="text-base font-bold px-2">Audio Codec</Label>
                  <Select.Root type="single" bind:value={audioCodec}>
                    <Select.Trigger
                      class="h-auto! w-full rounded-2xl border-2 bg-muted/30 px-5 py-4 text-base font-bold transition-all hover:bg-muted/50"
                    >
                      <div class="flex flex-col items-start gap-0.5">
                        <span>{selectedAudioCodecInfo.label}</span>
                        <span
                          class="text-[10px] font-medium opacity-60 uppercase tracking-wide"
                          >{selectedAudioCodecInfo.description}</span
                        >
                      </div>
                    </Select.Trigger>
                    <Select.Content class="rounded-2xl border-2 p-2 shadow-2xl">
                      {#each audioCodecOptions as option}
                        <Select.Item
                          value={option.value}
                          class="rounded-xl py-3 px-4 font-bold transition-all focus:bg-primary focus:text-primary-foreground mb-1 last:mb-0"
                        >
                          <div class="flex flex-col gap-0.5">
                            <span>{option.label}</span>
                            <span
                              class="text-[10px] font-medium opacity-60 uppercase tracking-wide"
                              >{option.description}</span
                            >
                          </div>
                        </Select.Item>
                      {/each}
                    </Select.Content>
                  </Select.Root>
                </div>
              </div>

              <!-- CRF Slider -->
              <div class="space-y-4 px-2">
                <div class="flex items-center justify-between">
                  <Label class="text-base font-bold">Quality (CRF)</Label>
                  <span class="text-lg font-black tabular-nums"
                    >{crfValue[0]}</span
                  >
                </div>
                <Slider
                  type="multiple"
                  bind:value={crfValue}
                  min={currentCrfRange.min}
                  max={currentCrfRange.max}
                  step={1}
                  class="py-2"
                />
                <div
                  class="flex justify-between text-xs font-bold text-muted-foreground"
                >
                  <span>Higher Quality</span>
                  <span>Smaller File</span>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>

    <div
      class="space-y-6 animate-in slide-in-from-top-4 duration-500 delay-150"
    >
      <div
        class="relative overflow-hidden rounded-[2.5rem] border bg-card p-2 shadow-2xl shadow-primary/5 transition-all focus-within:ring-2 focus-within:ring-primary/20"
      >
        <div class="relative">
          <Type
            class="absolute left-6 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground"
          />

          <input
            type="text"
            bind:value={customName}
            placeholder="Video Name (Optional)"
            class="h-16 w-full rounded-[2rem] border-none bg-transparent pl-14 pr-14 text-lg font-medium focus:outline-none focus:ring-0"
            disabled={loading}
          />

          {#if customName}
            <button
              onclick={() => (customName = "")}
              class="absolute right-6 top-1/2 -translate-y-1/2 rounded-full p-2 hover:bg-muted transition-colors"
              aria-label="Clear name"
            >
              <X class="h-5 w-5 text-muted-foreground" />
            </button>
          {/if}
        </div>
      </div>

      <Button.Root
        onclick={startDownload}
        disabled={loading || !url}
        class="h-20 w-full rounded-[2.5rem] text-2xl font-black shadow-2xl shadow-primary/20 transition-all hover:scale-[1.01] active:scale-[0.99]"
      >
        <Download class="mr-3 h-8 w-8 stroke-[3px]" />

        Start Download
      </Button.Root>
    </div>
  {/if}
</div>
