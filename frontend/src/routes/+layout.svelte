<script lang="ts">
  import "./layout.css";
  import favicon from "$lib/assets/favicon1.png";
  import { page } from "$app/stores";
  import { ModeWatcher } from "mode-watcher";
  import ModeToggle from "$lib/components/mode-toggle.svelte";
  import ServerInfo from "$lib/components/server-info.svelte";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import * as Button from "$lib/components/ui/button/index.js";
  import { ytdlpApi } from "$lib/api-client";
  import {
    Video,
    Download,
    AlertCircle,
    Menu,
    RefreshCw,
  } from "@lucide/svelte";

  let { children } = $props();

  const links = [
    { href: "/", label: "Videos", icon: Video },
    { href: "/download", label: "Download", icon: Download },
    { href: "/errors", label: "Errors", icon: AlertCircle },
  ];

  let isMenuOpen = $state(false);
  let isUpdatingYtdlp = $state(false);

  async function updateYtdlp() {
    if (!confirm("Update yt-dlp binary? This may take a moment.")) return;
    isUpdatingYtdlp = true;
    try {
      const res = await ytdlpApi.updateYtdlp();
      const output = Object.values(res.data).join("\n");
      alert("Update completed:\n" + output);
    } catch (e) {
      console.error(e);
      alert("Update failed.");
    } finally {
      isUpdatingYtdlp = false;
    }
  }
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>Vidra</title>
</svelte:head>

<ModeWatcher />

<div
  class="relative flex min-h-screen flex-col bg-background font-sans antialiased"
>
  <header
    class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60"
  >
    <div class="mx-auto flex h-16 max-w-2xl items-center px-4">
      <div class="mr-4 flex flex-1 items-center">
        <a href="/" class="mr-6 flex items-center space-x-2">
          <div class="rounded-xl bg-primary p-1.5">
            <Video class="h-5 w-5 text-primary-foreground" />
          </div>
          <span class="font-bold tracking-tight text-xl">Vidra</span>
        </a>
        <nav class="hidden items-center space-x-6 text-sm font-medium md:flex">
          {#each links as link}
            <a
              href={link.href}
              class="transition-colors hover:text-foreground/80 {$page.url
                .pathname === link.href
                ? 'text-foreground'
                : 'text-foreground/60'}"
            >
              {link.label}
            </a>
          {/each}
        </nav>
      </div>

      <div class="flex items-center gap-2">
        <ServerInfo />
        <button
          onclick={updateYtdlp}
          disabled={isUpdatingYtdlp}
          title="Update yt-dlp"
          class="inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-xl bg-muted transition-all hover:bg-muted/80 active:scale-95 disabled:opacity-50"
        >
          <RefreshCw class="h-5 w-5 {isUpdatingYtdlp ? 'animate-spin' : ''}" />
        </button>
        <ModeToggle />
        <button
          class="inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-xl bg-muted transition-colors hover:bg-muted/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring md:hidden"
          onclick={() => (isMenuOpen = !isMenuOpen)}
        >
          <Menu class="h-5 w-5" />
          <span class="sr-only">Toggle Menu</span>
        </button>
      </div>
    </div>
  </header>

  {#if isMenuOpen}
    <div
      class="fixed inset-0 z-50 grid h-screen w-screen grid-rows-[auto_1fr] bg-background animate-in fade-in duration-200 md:hidden"
    >
      <div class="flex h-16 items-center justify-between border-b px-4">
        <a
          href="/"
          class="flex items-center space-x-2"
          onclick={() => (isMenuOpen = false)}
        >
          <div class="rounded-xl bg-primary p-1.5">
            <Video class="h-5 w-5 text-primary-foreground" />
          </div>
          <span class="font-bold text-xl">Vidra</span>
        </a>
        <button
          class="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-muted"
          onclick={() => (isMenuOpen = false)}
        >
          <Menu class="h-5 w-5" />
        </button>
      </div>
      <nav class="flex flex-col space-y-2 p-6">
        {#each links as link}
          <a
            href={link.href}
            class="flex items-center rounded-xl p-4 text-lg font-semibold transition-colors {$page
              .url.pathname === link.href
              ? 'bg-primary/10 text-primary'
              : 'text-foreground/60 hover:bg-muted'}"
            onclick={() => (isMenuOpen = false)}
          >
            <link.icon class="mr-3 h-5 w-5" />
            {link.label}
          </a>
        {/each}
      </nav>
    </div>
  {/if}

  <main class="flex-1">
    <div class="mx-auto max-w-2xl px-4 py-8">
      {@render children()}
    </div>
  </main>

  <footer class="border-t py-8">
    <div
      class="mx-auto max-w-2xl px-4 flex flex-col items-center justify-between gap-4 md:flex-row"
    >
      <p
        class="text-balance text-center text-xs font-medium text-muted-foreground md:text-left"
      >
        © {new Date().getFullYear()} Vidra • Simple Video Downloader
      </p>
    </div>
  </footer>
</div>
