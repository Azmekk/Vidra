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
    class="w-64 rounded-[1.5rem] p-4 shadow-2xl border-2"
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
        <div
          class="flex items-center justify-between bg-muted/30 p-3 rounded-2xl"
        >
          <div class="flex items-center gap-3">
            <div
              class="rounded-full p-2 {info?.status === 'ok'
                ? 'bg-green-500/10 text-green-500'
                : 'bg-red-500/10 text-red-500'}"
            >
              <Server class="h-4 w-4" />
            </div>
            <span class="text-sm font-bold">Status</span>
          </div>
          <span
            class="text-xs font-black uppercase {info?.status === 'ok'
              ? 'text-green-500'
              : 'text-red-500'}"
          >
            {info?.status === "ok" ? "Online" : "Offline"}
          </span>
        </div>

        <!-- WebSocket Status -->
        <div
          class="flex items-center justify-between bg-muted/30 p-3 rounded-2xl"
        >
          <div class="flex items-center gap-3">
            <div
              class="rounded-full p-2 {wsStatus === 'connected'
                ? 'bg-blue-500/10 text-blue-500'
                : 'bg-amber-500/10 text-amber-500'}"
            >
              {#if wsStatus === "connected"}
                <Wifi class="h-4 w-4" />
              {:else}
                <WifiOff class="h-4 w-4" />
              {/if}
            </div>
            <span class="text-sm font-bold">WebSocket</span>
          </div>
          <span
            class="text-xs font-black uppercase {wsStatus === 'connected'
              ? 'text-blue-500'
              : 'text-amber-500'}"
          >
            {wsStatus === "connected" ? "Connected" : "Waiting"}
          </span>
        </div>

        <!-- Disk Usage -->
        <div
          class="flex items-center justify-between bg-muted/30 p-3 rounded-2xl"
        >
          <div class="flex items-center gap-3">
            <div class="rounded-full p-2 bg-purple-500/10 text-purple-500">
              <HardDrive class="h-4 w-4" />
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
          class="w-full py-2 text-xs font-bold uppercase tracking-widest opacity-40 hover:opacity-100 transition-opacity"
        >
          Refresh Info
        </button>
      </div>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
