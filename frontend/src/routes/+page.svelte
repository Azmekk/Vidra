<script lang="ts">
  import { onMount } from "svelte";
  import { videosApi } from "$lib/api-client";
  import { WsEventType } from "$lib/constants/websocket";
  import type {
    HandlersVideoResponse,
    ServicesDownloadProgressDTO,
  } from "$api/index";

  import PageHeader from "$lib/components/page-header.svelte";
  import SearchSortBar from "$lib/components/search-sort-bar.svelte";
  import VideosGrid from "$lib/components/videos-grid.svelte";
  import VideosPagination from "$lib/components/videos-pagination.svelte";

  let { data } = $props();
  let videos: HandlersVideoResponse[] = $state([]);
  let totalCount = $state(0);
  let totalPages = $state(0);
  let currentPage = $state(1);
  let limit = $state(10);

  // Sync state when data changes (SSR initial load and navigation)
  $effect(() => {
    videos = data.paginatedVideos.videos || [];
    totalCount = data.paginatedVideos.totalCount || 0;
    totalPages = data.paginatedVideos.totalPages || 0;
    currentPage = data.paginatedVideos.currentPage || 1;
    limit = data.paginatedVideos.limit || 10;
  });

  let progressMap: Record<string, ServicesDownloadProgressDTO> = $state({});
  let ws: WebSocket;

  let search = $state("");
  let debouncedSearch = $state("");
  let order = $state("created_at_desc");
  let isLoading = $state(false);

  let editingId = $state<string | null>(null);
  let editingName = $state("");

  let copiedId = $state<string | null>(null);
  let activeVideoId: string | null = $state(null);

  function connectWebSocket() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/ws`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log("WebSocket connected");
      window.dispatchEvent(new CustomEvent("ws_open"));
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === WsEventType.Progress) {
        progressMap[data.payload.id] = data.payload;
        if (data.payload.status === "completed") {
          (async () => {
            try {
              const res = await videosApi.getVideo(data.payload.id);
              videos = videos.map((v) =>
                v.id === data.payload.id ? res.data : v,
              );
            } catch (e) {
              console.error("Error fetching updated video:", e);
            }
          })();
        }
      } else if (data.type === WsEventType.VideoCreated) {
        if (currentPage === 1 && order === "created_at_desc") {
          videos = [data.payload, ...videos.slice(0, limit - 1)];
          totalCount++;
          totalPages = Math.ceil(totalCount / limit);
        } else {
          totalCount++;
          totalPages = Math.ceil(totalCount / limit);
        }
      } else if (data.type === WsEventType.VideoDeleted) {
        const videoExists = videos.some((v) => v.id === data.payload.id);
        if (videoExists) {
          fetchVideos(debouncedSearch, order, currentPage);
        } else {
          totalCount--;
          totalPages = Math.ceil(totalCount / limit);
        }
        delete progressMap[data.payload.id];
      }
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected, retrying in 3s...");
      window.dispatchEvent(new CustomEvent("ws_close"));
      setTimeout(connectWebSocket, 3000);
    };

    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
      ws.close();
    };
  }

  async function fetchVideos(query: string, sort: string, page: number) {
    isLoading = true;
    try {
      const res = await videosApi.listVideos(query, sort, page, limit);
      videos = res.data.videos || [];
      totalCount = res.data.totalCount || 0;
      totalPages = res.data.totalPages || 0;
      currentPage = res.data.currentPage || 1;
    } catch (e) {
      console.error("Error fetching videos", e);
    } finally {
      isLoading = false;
    }
  }

  async function renameVideo(id: string) {
    if (!editingName.trim()) return;
    try {
      await videosApi.updateVideo(id, { name: editingName });
      editingId = null;
      await fetchVideos(debouncedSearch, order, currentPage);
    } catch (e) {
      console.error("Error renaming video", e);
    }
  }

  function startEditing(video: HandlersVideoResponse) {
    editingId = video.id!;
    editingName = video.name!;
  }

  function cancelEditing() {
    editingId = null;
    editingName = "";
  }

  // Debounce search
  $effect(() => {
    const currentSearch = search;
    const handler = setTimeout(() => {
      debouncedSearch = currentSearch;
      currentPage = 1;
    }, 300);

    return () => clearTimeout(handler);
  });

  // Re-fetch when debouncedSearch, order, or currentPage changes
  $effect(() => {
    fetchVideos(debouncedSearch, order, currentPage);
  });

  function copyToClipboard(text: string, id: string) {
    navigator.clipboard.writeText(text);
    copiedId = id;
    setTimeout(() => {
      if (copiedId === id) copiedId = null;
    }, 2000);
  }

  async function deleteVideo(id: string) {
    if (!confirm("Are you sure you want to delete this video?")) return;
    try {
      await videosApi.deleteVideo(id);
    } catch (e) {
      console.error("Error deleting video", e);
    }
  }

  function togglePlay(id: string) {
    if (activeVideoId === id) {
      activeVideoId = null;
    } else {
      activeVideoId = id;
    }
  }

  function handleOrderChange() {
    currentPage = 1;
  }

  function handlePageChange(page: number) {
    fetchVideos(debouncedSearch, order, page);
  }

  onMount(() => {
    connectWebSocket();

    return () => {
      if (ws) ws.close();
    };
  });
</script>

<div class="space-y-10">
  <PageHeader {totalCount} />

  <SearchSortBar
    bind:search
    bind:order
    {isLoading}
    onOrderChange={handleOrderChange}
  />

  <VideosGrid
    {videos}
    {progressMap}
    {activeVideoId}
    {copiedId}
    {editingId}
    bind:editingName
    onTogglePlay={togglePlay}
    onStartEditing={startEditing}
    onCancelEditing={cancelEditing}
    onRename={renameVideo}
    onDelete={deleteVideo}
    onCopyId={copyToClipboard}
  />

  <VideosPagination
    {totalCount}
    {totalPages}
    bind:currentPage
    {limit}
    onPageChange={handlePageChange}
  />
</div>
