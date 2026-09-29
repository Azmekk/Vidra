<script lang="ts">
	import { Sun, Moon } from "@lucide/svelte";
	import { toggleMode } from "mode-watcher";
	import * as Button from "$lib/components/ui/button/index.js";

	let { class: className = "" } = $props();
</script>

<Button.Root variant="secondary" size="icon" onclick={toggleMode} class={`h-10 w-10 cursor-pointer rounded-xl ${className}`}>
	<Sun
		class="h-5 w-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0"
	/>
	<Moon
		class="absolute h-5 w-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
	/>
	<span class="sr-only">Toggle theme</span>
</Button.Root>
