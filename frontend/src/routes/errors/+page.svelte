<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import * as Card from "$lib/components/ui/card/index.js";
  import * as Button from "$lib/components/ui/button/index.js";
  import { Badge } from "$lib/components/ui/badge/index.js";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import { ytdlpApi } from "$lib/api-client";
  import {
    RefreshCw,
    AlertTriangle,
    Terminal,
    ChevronDown,
    History,
    ShieldAlert,
  } from "@lucide/svelte";

  let { data } = $props();
  let errors = $derived(data.errors || []);
  let expandedId = $state<string | null>(null);

  function refresh() {
    invalidateAll();
  }

  function formatDate(dateStr?: string) {
    if (!dateStr) return "-";
    return new Date(dateStr).toLocaleString();
  }

  function toggleExpand(id: string) {
    expandedId = expandedId === id ? null : id;
  }
</script>

<div class="space-y-10 animate-in fade-in duration-700">
  <div class="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
    <div>
      <h1 class="text-4xl font-extrabold tracking-tight">System Logs</h1>
      <p class="mt-1 text-lg text-muted-foreground font-medium">
        Monitor download errors and system status.
      </p>
    </div>
    <div class="flex gap-3">
      <Button.Root onclick={refresh} variant="secondary" size="lg" class="rounded-2xl h-12 px-6 font-bold shadow-sm">
        <RefreshCw class="mr-2 h-5 w-5" />
        Refresh
      </Button.Root>
    </div>
  </div>

  {#if errors.length === 0}
    <div class="flex min-h-[400px] flex-col items-center justify-center rounded-[3rem] border-2 border-dashed p-12 text-center animate-in zoom-in-95 duration-500">
      <div class="rounded-3xl bg-green-500/10 p-6 text-green-600 dark:text-green-500">
        <History class="h-12 w-12" />
      </div>
      <h2 class="mt-6 text-2xl font-bold tracking-tight">No errors recorded</h2>
      <p class="mt-2 text-muted-foreground max-w-[250px] mx-auto font-medium">
        System is running smoothly. All clear!
      </p>
    </div>
  {:else}
    <div class="grid gap-6">
      {#each errors as error (error.id)}
        {@const isExpanded = expandedId === error.id}
        <div class="group relative overflow-hidden rounded-[2.5rem] border bg-card transition-all hover:shadow-2xl hover:shadow-primary/5 {isExpanded ? 'ring-2 ring-primary/20' : ''}">
          <button
            class="flex w-full cursor-pointer items-start justify-between p-6 transition-colors hover:bg-muted/30 sm:p-8"
            onclick={() => error.id && toggleExpand(error.id)}
          >
            <div class="flex gap-6">
              <div class="mt-1 shrink-0 rounded-2xl bg-destructive/10 p-3 text-destructive shadow-sm shadow-destructive/10">
                <AlertTriangle class="h-6 w-6" />
              </div>
              <div class="space-y-2 text-left">
                <div class="flex flex-wrap items-center gap-3">
                  <span class="text-lg font-black uppercase tracking-wider text-destructive">Error</span>
                  <Badge variant="outline" class="rounded-full px-3 py-0.5 font-bold border-none bg-muted text-xs uppercase tracking-tight">
                    {error.videoId || "System"}
                  </Badge>
                  <span class="text-xs font-bold text-muted-foreground opacity-60">{formatDate(error.createdAt)}</span>
                </div>
                <p class="text-lg font-bold leading-tight tracking-tight sm:text-xl">
                  {error.errorMessage}
                </p>
              </div>
            </div>
            <div class="shrink-0 h-10 w-10 flex items-center justify-center rounded-xl bg-muted transition-transform duration-300 {isExpanded ? 'rotate-180' : ''}">
              <ChevronDown class="h-5 w-5" />
            </div>
          </button>

          {#if isExpanded}
            <div class="bg-muted/20 p-6 pt-2 animate-in slide-in-from-top-4 duration-300 sm:p-8 sm:pt-4">
              <Separator class="mb-8 opacity-50" />
              <div class="space-y-6">
                {#if error.command}
                  <div class="space-y-3">
                    <div class="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-muted-foreground">
                      <Terminal class="h-4 w-4" />
                      Executed Command
                    </div>
                    <div class="rounded-2xl bg-zinc-950 p-6 font-mono text-sm text-green-500 overflow-x-auto border-2 border-white/5 shadow-2xl">
                      <span class="select-none opacity-40 mr-3">$</span>{error.command}
                    </div>
                  </div>
                {/if}

                <div class="space-y-3">
                  <div class="flex items-center gap-2 text-xs font-black uppercase tracking-widest text-muted-foreground">
                    <History class="h-4 w-4" />
                    Process Output
                  </div>
                  <div class="max-h-[400px] overflow-y-auto rounded-3xl bg-zinc-950 p-6 font-mono text-xs leading-relaxed border-2 border-white/5 shadow-inner">
                    {#if error.output}
                      <pre class="whitespace-pre-wrap break-all text-zinc-400">{error.output}</pre>
                    {:else}
                      <div class="flex flex-col items-center justify-center py-12 text-zinc-700 italic">
                        <ShieldAlert class="mb-3 h-10 w-10 opacity-20" />
                        <span class="font-bold">No output captured</span>
                      </div>
                    {/if}
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>
