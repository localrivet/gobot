<script lang="ts">
	let { children } = $props();
</script>

<div class="layout-auth">
	<div class="relative z-10 w-full max-w-md">
		<a href="/" class="block text-center mb-10">
			<span class="font-display text-3xl font-black text-gradient"> Gobot </span>
		</a>
		<main id="main-content">
			{@render children()}
		</main>
		<p class="text-center mt-8 text-sm text-base-content/60">
			<a href="/" class="hover:text-base-content transition-colors"> Back to home </a>
		</p>
	</div>
</div>
