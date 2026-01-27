<script lang="ts">
	import { MarketingNav } from '$lib/components/navigation';
	import HeadScripts from '$lib/components/HeadScripts.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import type { LayoutServerData } from './+layout.server';

	let { children, data }: { children: any; data: LayoutServerData } = $props();

	// Convert Levee menu items to nav format, fallback to hardcoded
	const navItems = $derived(data.headerMenu?.items
		.filter((item) => item.is_active)
		.map((item) => ({
			label: item.label,
			href: item.url ?? '#'
		})) ?? [{ label: 'Pricing', href: '/pricing' }]);
</script>

<HeadScripts />

<div class="layout-www">
	<div class="relative z-10">
		<MarketingNav items={navItems} />
		<main id="main-content">
			{@render children()}
		</main>
		<Footer />
	</div>
</div>
