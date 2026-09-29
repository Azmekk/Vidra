<script lang="ts">
  import * as Pagination from "$lib/components/ui/pagination/index.js";
  import { ChevronLeft, ChevronRight } from "@lucide/svelte";

  let {
    totalCount,
    totalPages,
    currentPage = $bindable(1),
    limit,
    onPageChange,
  }: {
    totalCount: number;
    totalPages: number;
    currentPage: number;
    limit: number;
    onPageChange: (page: number) => void;
  } = $props();
</script>

{#if totalPages > 1}
  <div class="flex justify-center mt-12 pb-12">
    <Pagination.Root count={totalCount} perPage={limit} bind:page={currentPage}>
      {#snippet children({ pages })}
        <Pagination.Content>
          <Pagination.Item>
            <Pagination.PrevButton onclick={() => onPageChange(currentPage - 1)}>
              <ChevronLeft class="h-4 w-4" />
              <span class="hidden sm:inline">Previous</span>
            </Pagination.PrevButton>
          </Pagination.Item>
          {#each pages as page (page.key)}
            {#if page.type === "ellipsis"}
              <Pagination.Item>
                <Pagination.Ellipsis />
              </Pagination.Item>
            {:else}
              <Pagination.Item>
                <Pagination.Link
                  {page}
                  isActive={currentPage === page.value}
                  onclick={() => onPageChange(page.value)}
                >
                  {page.value}
                </Pagination.Link>
              </Pagination.Item>
            {/if}
          {/each}
          <Pagination.Item>
            <Pagination.NextButton onclick={() => onPageChange(currentPage + 1)}>
              <span class="hidden sm:inline">Next</span>
              <ChevronRight class="h-4 w-4" />
            </Pagination.NextButton>
          </Pagination.Item>
        </Pagination.Content>
      {/snippet}
    </Pagination.Root>
  </div>
{/if}
