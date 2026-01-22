<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { systemApi } from "$lib/api-client";
  import type { HandlersSystemInfoResponse } from "$api/index";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
  import { Server, HardDrive, Wifi, WifiOff, Loader2 } from "@lucide/svelte";
  import { browser } from "$app/environment";

  let info = $state<HandlersSystemInfoResponse | null>(null);
  let loading = $state(true);
  let wsStatus = $state<"connected" | "disconnected" | "connecting">(
    "connecting",
  );
  let interval: ReturnType<typeof setInterval>;

  async function fetchInfo() {
    try {
      const res = await systemApi.getSystemInfo();
      info = res.data;
    } catch (e) {
      console.error("Failed to fetch system info:", e);
    } finally {
      loading = false;
    }
  }

  function checkWs() {
    if (!browser) return;
    // This is a simple check, in a real app we might want to hook into the actual WS service
    // For now, we'll try to find if there's any active WS connection or just assume based on common state
    // Actually, let's just use a simple flag that the layout can pass or we can listen to events.
    // Since we don't have a global WS store yet, I'll just check if we can reach the WS endpoint.

    // Better: listen for the custom events if the websocket service emits them.
    // For this simple implementation, I'll just poll the system info which also acts as a health check.
  }

  onMount(() => {
    fetchInfo();
    interval = setInterval(fetchInfo, 30000); // Every 30s

    // Listen for websocket status events if they exist
    const handleWsOpen = () => (wsStatus = "connected");
    const handleWsClose = () => (wsStatus = "disconnected");

    window.addEventListener("ws_open", handleWsOpen);
    window.addEventListener("ws_close", handleWsClose);

    return () => {
      window.removeEventListener("ws_open", handleWsOpen);
      window.removeEventListener("ws_close", handleWsClose);
    };
  });

  onDestroy(() => {
    if (interval) clearInterval(interval);
  });

  function formatSize(gb: number) {
    if (gb < 1) {
      return (gb * 1024).toFixed(1) + " MB";
    }
    return gb.toFixed(2) + " GB";
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    class="inline-flex h-10 w-10 cursor-pointer items-center justify-center rounded-xl bg-muted transition-all hover:bg-muted/80 active:scale-95"
  >
    <Server class="h-5 w-5" />
    <span class="sr-only">Server Info</span>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="end"
    sideOffset={8}
    class="w-[calc(100vw-2rem)] sm:w-96 rounded-[1.5rem] p-4 sm:p-6 shadow-2xl border-2"
  >
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="font-bold text-sm uppercase tracking-wider opacity-60">
          System Status
        </h3>
        {#if loading && !info}
          <Loader2 class="h-4 w-4 animate-spin opacity-40" />
        {/if}
      </div>

      <div class="space-y-3">
        <!-- Server Status -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div
              class="rounded-full p-1.5 {info?.status === 'ok'
                ? 'bg-green-500/10 text-green-500'
                : 'bg-red-500/10 text-red-500'}"
            >
              <Server class="h-3.5 w-3.5" />
            </div>
            <span class="text-sm font-bold">Server</span>
          </div>
          <span
            class="text-xs font-black {info?.status === 'ok'
              ? 'text-green-500'
              : 'text-red-500'}"
          >
            {info?.status === "ok" ? "OK" : "OFF"}
          </span>
        </div>

        <!-- WebSocket Status -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div
              class="rounded-full p-1.5 {wsStatus === 'connected'
                ? 'bg-green-500/10 text-green-500'
                : 'bg-amber-500/10 text-amber-500'}"
            >
              {#if wsStatus === "connected"}
                <Wifi class="h-3.5 w-3.5" />
              {:else}
                <WifiOff class="h-3.5 w-3.5" />
              {/if}
            </div>
            <span class="text-sm font-bold">WebSocket</span>
          </div>
          <span
            class="text-xs font-black {wsStatus === 'connected'
              ? 'text-green-500'
              : 'text-amber-500'}"
          >
            {wsStatus === "connected" ? "OK" : "WAIT"}
          </span>
        </div>

        <!-- Disk Usage -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="rounded-full p-1.5 bg-purple-500/10 text-purple-500">
              <HardDrive class="h-3.5 w-3.5" />
            </div>
            <span class="text-sm font-bold">Storage</span>
          </div>
          <span class="text-xs font-black tabular-nums">
            {info ? formatSize(info.diskUsageGB ?? 0) : "0.0 GB"}
          </span>
        </div>
      </div>

      <div class="pt-2">
        <button
          onclick={fetchInfo}
          class="w-full py-3 sm:py-2 text-xs font-bold uppercase tracking-widest rounded-xl opacity-40 hover:opacity-100 hover:bg-muted/50 active:bg-muted/70 transition-all"
        >
          Refresh Info
        </button>
      </div>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
