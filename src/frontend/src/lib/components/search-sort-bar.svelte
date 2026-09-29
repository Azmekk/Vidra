<script lang="ts">
  import * as Select from "$lib/components/ui/select/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Search, SortAsc, Loader } from "@lucide/svelte";

  let {
    search = $bindable(""),
    order = $bindable(""),
    isLoading,
    onOrderChange,
  }: {
    search: string;
    order: string;
    isLoading: boolean;
    onOrderChange: () => void;
  } = $props();
</script>

<div
  class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between bg-card p-6 rounded-[2rem] border shadow-sm"
>
  <div class="relative flex-1 max-w-md">
    <Search
      class="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground"
    />
    <Input
      type="text"
      placeholder="Search videos..."
      bind:value={search}
      class="h-12 rounded-xl pl-12 pr-12 bg-muted/50 border-none ring-offset-background focus-visible:ring-2 focus-visible:ring-primary/20 transition-all"
    />
    {#if isLoading}
      <div class="absolute right-4 top-1/2 -translate-y-1/2">
        <Loader class="h-5 w-5 animate-spin text-primary" />
      </div>
    {/if}
  </div>
  <div class="flex items-center gap-3">
    <SortAsc class="h-5 w-5 text-muted-foreground" />
    <Select.Root type="single" bind:value={order} onValueChange={onOrderChange}>
      <Select.Trigger
        class="h-12 min-w-[180px] rounded-xl bg-muted/50 border-none font-bold"
      >
        {order.replace("_", " ").toUpperCase()}
      </Select.Trigger>
      <Select.Content class="rounded-xl border shadow-xl">
        <Select.Item value="created_at_desc">Latest First</Select.Item>
        <Select.Item value="created_at_asc">Oldest First</Select.Item>
        <Select.Item value="name_asc">Name A-Z</Select.Item>
        <Select.Item value="name_desc">Name Z-A</Select.Item>
        <Select.Item value="status_asc">Status A-Z</Select.Item>
        <Select.Item value="status_desc">Status Z-A</Select.Item>
      </Select.Content>
    </Select.Root>
  </div>
</div>
