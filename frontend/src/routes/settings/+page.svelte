<script lang="ts">
  import { onMount } from "svelte";
  import { settingsApi } from "$lib/api-client";
  import { setMode, mode } from "mode-watcher";
  import * as Button from "$lib/components/ui/button/index.js";
  import * as Select from "$lib/components/ui/select/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Slider } from "$lib/components/ui/slider/index.js";
  import { Switch } from "$lib/components/ui/switch/index.js";
  import {
    Settings2,
    Save,
    Globe,
    Palette,
    Video,
    Volume2,
    Gauge,
    LoaderCircle,
    Check,
  } from "@lucide/svelte";

  let loading = $state(true);
  let saving = $state(false);
  let saved = $state(false);
  let error = $state<string | null>(null);

  let proxyUrl = $state("");
  let defaultReEncode = $state(true);
  let defaultVideoCodec = $state("libx264");
  let defaultAudioCodec = $state("aac");
  let defaultCrf = $state<number[]>([23]);
  let theme = $state<"light" | "dark" | "system">("system");

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

  const themeOptions = [
    { value: "system", label: "System", description: "Follow system preference" },
    { value: "light", label: "Light", description: "Always light mode" },
    { value: "dark", label: "Dark", description: "Always dark mode" },
  ];

  const crfRanges: Record<
    string,
    { min: number; max: number; default: number }
  > = {
    libx264: { min: 18, max: 28, default: 23 },
    "libvpx-vp9": { min: 24, max: 40, default: 32 },
    vp9_qsv: { min: 18, max: 35, default: 25 },
  };

  const currentCrfRange = $derived(crfRanges[defaultVideoCodec] || crfRanges.libx264);
  const selectedVideoCodecInfo = $derived(
    videoCodecOptions.find((o) => o.value === defaultVideoCodec) ||
      videoCodecOptions[0],
  );
  const selectedAudioCodecInfo = $derived(
    audioCodecOptions.find((o) => o.value === defaultAudioCodec) ||
      audioCodecOptions[0],
  );
  const selectedThemeInfo = $derived(
    themeOptions.find((o) => o.value === theme) || themeOptions[0],
  );

  onMount(async () => {
    try {
      const res = await settingsApi.getSettings();
      proxyUrl = res.data.proxyUrl || "";
      defaultReEncode = res.data.defaultReEncode ?? true;
      defaultVideoCodec = res.data.defaultVideoCodec || "libx264";
      defaultAudioCodec = res.data.defaultAudioCodec || "aac";
      defaultCrf = [res.data.defaultCrf ?? 23];
      theme = (res.data.theme as "light" | "dark" | "system") || "system";
    } catch (e) {
      console.error("Failed to load settings", e);
      error = "Failed to load settings";
    } finally {
      loading = false;
    }
  });

  async function saveSettings() {
    saving = true;
    error = null;
    saved = false;
    try {
      await settingsApi.updateSettings({
        proxyUrl,
        defaultReEncode,
        defaultVideoCodec,
        defaultAudioCodec,
        defaultCrf: defaultCrf[0],
        theme,
      });

      // Apply theme change immediately
      setMode(theme);

      saved = true;
      setTimeout(() => {
        saved = false;
      }, 2000);
    } catch (e) {
      console.error("Failed to save settings", e);
      error = "Failed to save settings";
    } finally {
      saving = false;
    }
  }
</script>

<div
  class="mx-auto space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700"
>
  <div class="flex flex-col gap-6">
    <div class="flex items-center gap-4">
      <div
        class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary"
      >
        <Settings2 class="h-6 w-6" />
      </div>
      <h1 class="text-4xl font-extrabold tracking-tight">Settings</h1>
    </div>
    <p class="text-lg text-muted-foreground font-medium leading-relaxed">
      Configure default download options and application preferences.
    </p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <LoaderCircle class="h-8 w-8 animate-spin text-primary" />
    </div>
  {:else}
    <div class="space-y-8">
      <!-- Proxy Settings -->
      <div
        class="overflow-hidden rounded-[2.5rem] border bg-card p-8 shadow-2xl shadow-primary/5"
      >
        <div class="flex items-center gap-3 mb-6">
          <div class="rounded-xl bg-blue-500/10 p-2.5 text-blue-500">
            <Globe class="h-5 w-5" />
          </div>
          <h2 class="text-xl font-bold">Network</h2>
        </div>

        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="proxy" class="text-base font-bold">Proxy URL</Label>
            <p class="text-sm text-muted-foreground">
              Optional HTTP/SOCKS proxy for yt-dlp downloads (e.g., socks5://127.0.0.1:1080)
            </p>
          </div>
          <Input
            id="proxy"
            type="text"
            bind:value={proxyUrl}
            placeholder="socks5://127.0.0.1:1080"
            class="h-14 rounded-2xl bg-muted/50 border-none px-5 text-base font-medium"
          />
        </div>
      </div>

      <!-- Theme Settings -->
      <div
        class="overflow-hidden rounded-[2.5rem] border bg-card p-8 shadow-2xl shadow-primary/5"
      >
        <div class="flex items-center gap-3 mb-6">
          <div class="rounded-xl bg-purple-500/10 p-2.5 text-purple-500">
            <Palette class="h-5 w-5" />
          </div>
          <h2 class="text-xl font-bold">Appearance</h2>
        </div>

        <div class="space-y-4">
          <Label class="text-base font-bold">Theme</Label>
          <Select.Root type="single" bind:value={theme}>
            <Select.Trigger
              class="h-auto! w-full rounded-2xl border-2 bg-muted/30 px-5 py-4 text-base font-bold transition-all hover:bg-muted/50"
            >
              <div class="flex flex-col items-start gap-0.5">
                <span>{selectedThemeInfo.label}</span>
                <span
                  class="text-[10px] font-medium opacity-60 uppercase tracking-wide"
                  >{selectedThemeInfo.description}</span
                >
              </div>
            </Select.Trigger>
            <Select.Content class="rounded-2xl border-2 p-2 shadow-2xl">
              {#each themeOptions as option}
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

      <!-- Encoding Defaults -->
      <div
        class="overflow-hidden rounded-[2.5rem] border bg-card p-8 shadow-2xl shadow-primary/5"
      >
        <div class="flex items-center gap-3 mb-6">
          <div class="rounded-xl bg-green-500/10 p-2.5 text-green-500">
            <Video class="h-5 w-5" />
          </div>
          <h2 class="text-xl font-bold">Default Encoding</h2>
        </div>

        <div class="space-y-6">
          <div class="flex items-center justify-between">
            <div class="space-y-0.5">
              <Label class="text-base font-bold">Re-encode by Default</Label>
              <p class="text-sm text-muted-foreground">
                Automatically re-encode downloaded videos
              </p>
            </div>
            <Switch bind:checked={defaultReEncode} aria-label="Toggle default re-encoding" />
          </div>

          {#if defaultReEncode}
            <div
              class="space-y-6 pt-4 border-t border-muted animate-in slide-in-from-top-2 duration-300"
            >
              <div class="grid gap-6 sm:grid-cols-2">
                <!-- Video Codec -->
                <div class="space-y-3">
                  <div class="flex items-center gap-2">
                    <Video class="h-4 w-4 text-muted-foreground" />
                    <Label class="text-base font-bold">Video Codec</Label>
                  </div>
                  <Select.Root type="single" bind:value={defaultVideoCodec}>
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
                  <div class="flex items-center gap-2">
                    <Volume2 class="h-4 w-4 text-muted-foreground" />
                    <Label class="text-base font-bold">Audio Codec</Label>
                  </div>
                  <Select.Root type="single" bind:value={defaultAudioCodec}>
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
              <div class="space-y-4">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <Gauge class="h-4 w-4 text-muted-foreground" />
                    <Label class="text-base font-bold">Quality (CRF)</Label>
                  </div>
                  <span class="text-lg font-black tabular-nums"
                    >{defaultCrf[0]}</span
                  >
                </div>
                <Slider
                  type="multiple"
                  bind:value={defaultCrf}
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

      {#if error}
        <div
          class="flex items-center gap-4 rounded-[2rem] bg-destructive/10 p-6 text-destructive border border-destructive/20"
        >
          <span class="font-bold text-lg">{error}</span>
        </div>
      {/if}

      <Button.Root
        onclick={saveSettings}
        disabled={saving}
        class="h-16 w-full rounded-[2rem] text-xl font-black shadow-2xl shadow-primary/20 transition-all hover:scale-[1.01] active:scale-[0.99]"
      >
        {#if saving}
          <LoaderCircle class="mr-3 h-6 w-6 animate-spin" />
          Saving...
        {:else if saved}
          <Check class="mr-3 h-6 w-6" />
          Saved!
        {:else}
          <Save class="mr-3 h-6 w-6" />
          Save Settings
        {/if}
      </Button.Root>
    </div>
  {/if}
</div>
