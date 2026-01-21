<script lang="ts">
  import * as Button from "$lib/components/ui/button/index.js";
  import { Plus } from "@lucide/svelte";

  let {
    totalCount,
  }: {
    totalCount: number;
  } = $props();
</script>

<div
  class="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between"
>
  <div>
    <h1 class="text-4xl font-extrabold tracking-tight">Library</h1>
    <p class="mt-1 text-lg text-muted-foreground font-medium">
      Your downloaded collection ({totalCount} videos).
    </p>
  </div>
  <Button.Root
    href="/download"
    size="lg"
    class="rounded-2xl px-6 py-6 text-base font-bold shadow-lg shadow-primary/20 transition-all hover:scale-105 active:scale-95"
  >
    <Plus class="mr-2 h-5 w-5 stroke-[3px]" />
    Download New
  </Button.Root>
</div>
