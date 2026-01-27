<script lang="ts">
	import { marked } from 'marked';

	interface Props {
		content: string;
		class?: string;
	}

	let { content, class: className = '' }: Props = $props();

	// Configure marked for safe rendering
	marked.setOptions({
		breaks: true,
		gfm: true
	});

	let html = $derived(marked.parse(content || '') as string);
</script>

<div class="prose prose-sm max-w-none dark:prose-invert {className}">
	{@html html}
</div>
