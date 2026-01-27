<script lang="ts">
	import { setSEO, type SEOConfig } from '$lib/utils/seo';

	interface Props extends SEOConfig {
		jsonLd?: object | object[];
	}

	let { jsonLd, ...config }: Props = $props();

	const seo = $derived(setSEO(config));

	// Handle single or multiple JSON-LD schemas
	const jsonLdScripts = $derived.by(() => {
		if (!jsonLd) return [];
		return Array.isArray(jsonLd) ? jsonLd : [jsonLd];
	});
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<meta name="keywords" content={seo.keywords} />
	<link rel="canonical" href={seo.canonical} />

	{#if seo.noindex}
		<meta name="robots" content="noindex, nofollow" />
	{/if}

	<!-- Open Graph -->
	<meta property="og:type" content={seo.type} />
	<meta property="og:title" content={seo.title} />
	<meta property="og:description" content={seo.description} />
	<meta property="og:url" content={seo.url} />
	<meta property="og:image" content={seo.image} />
	<meta property="og:site_name" content={seo.siteName} />
	<meta property="og:locale" content={seo.locale} />

	<!-- Twitter Card -->
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:site" content={seo.twitterHandle} />
	<meta name="twitter:title" content={seo.title} />
	<meta name="twitter:description" content={seo.description} />
	<meta name="twitter:image" content={seo.image} />

	<!-- Article specific -->
	{#if seo.author}
		<meta name="author" content={seo.author} />
	{/if}
	{#if seo.publishedTime}
		<meta property="article:published_time" content={seo.publishedTime} />
	{/if}
	{#if seo.modifiedTime}
		<meta property="article:modified_time" content={seo.modifiedTime} />
	{/if}

	<!-- JSON-LD Structured Data -->
	{#each jsonLdScripts as schema}
		{@html `<script type="application/ld+json">${JSON.stringify(schema)}</script>`}
	{/each}
</svelte:head>
